package cel

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc/metadata"
)

func TestPlugin(t *testing.T) {
	ctx := context.Background()
	srcPath := filepath.Join("testdata", "plugin", "main.go")

	wasmPath := filepath.Join(t.TempDir(), "test.wasm")
	cmd := exec.CommandContext(ctx, "go", "build", "-o", wasmPath, srcPath)
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build main.go: %q: %v", b, err)
	}
	defer os.Remove(wasmPath)

	f, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(f)
	fn := &CELFunction{
		Name:   "foo",
		ID:     "foo",
		Return: types.IntType,
	}
	plugin, err := NewCELPlugin(ctx, CELPluginConfig{
		Name: "test",
		Wasm: WasmConfig{
			Reader: bytes.NewReader(f),
			Sha256: hex.EncodeToString(hash[:]),
		},
		Functions: []*CELFunction{fn},
		CacheDir:  t.TempDir(),
		Capability: &CELPluginCapability{
			// enable wasi-go host function.
			Network: &CELPluginNetworkCapability{},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := plugin.CreateInstance(ctx, nil); err != nil {
		t.Fatal(err)
	}
	v := plugin.Call(ctx, fn)
	iv, ok := v.(types.Int)
	if !ok {
		t.Fatalf("unexpected return type %T; want types.Int", v)
	}
	if int(iv) != 10 {
		t.Fatalf("failed to get response value: %v", iv)
	}
}

// newInstanceForSemaphoreTest constructs a minimally-initialized
// CELPluginInstance suitable for exercising the semaphore-acquisition path of
// Call/ValidatePlugin without spinning up a real WASM module. It only
// initializes fields that the methods touch before sem.Acquire returns.
func newInstanceForSemaphoreTest() *CELPluginInstance {
	return &CELPluginInstance{
		name:             "sem-test",
		sem:              semaphore.NewWeighted(1),
		reqCh:            make(chan []byte, 1),
		resCh:            make(chan []byte),
		gcErrCh:          make(chan error, 1),
		instanceModErrCh: make(chan error, 1),
		gcQueue:          make(chan struct{}, 1),
	}
}

// TestCallReleasesOnCtxCancel verifies that when a goroutine is blocked on
// Call() waiting for the semaphore (i.e. another call holds the instance),
// canceling that goroutine's ctx unblocks it promptly with ctx.Err(), instead
// of leaking the goroutine + arguments + span until the previous holder
// finally returns.
//
// Before this change, Call() acquired sync.Mutex.Lock() which is not
// context-aware, so a parked goroutine could not be woken by ctx.Done().
func TestCallReleasesOnCtxCancel(t *testing.T) {
	t.Parallel()

	inst := newInstanceForSemaphoreTest()

	// Simulate an in-flight call holding the instance lock.
	if err := inst.sem.Acquire(context.Background(), 1); err != nil {
		t.Fatalf("failed to pre-acquire semaphore: %v", err)
	}
	defer inst.sem.Release(1)

	ctx, cancel := context.WithCancel(context.Background())

	type result struct {
		val ref.Val
		err error
	}
	done := make(chan result, 1)
	go func() {
		fn := &CELFunction{Name: "f", ID: "f", Return: types.IntType}
		v, err := inst.Call(ctx, fn, metadata.MD{})
		done <- result{val: v, err: err}
	}()

	// Give the goroutine a moment to actually park on Acquire.
	time.Sleep(20 * time.Millisecond)

	select {
	case <-done:
		t.Fatal("Call returned before ctx cancel; expected it to be parked on the semaphore")
	default:
	}

	cancel()

	select {
	case r := <-done:
		if !errors.Is(r.err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", r.err)
		}
		if r.val != nil {
			t.Fatalf("expected nil value on cancel, got %v", r.val)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Call did not return within 2s after ctx cancel; semaphore is not context-aware")
	}
}

// TestCallReleasesOnCtxDeadline verifies deadline propagation: when the
// caller's ctx has a deadline, Call() blocked on the semaphore must return
// context.DeadlineExceeded shortly after the deadline.
func TestCallReleasesOnCtxDeadline(t *testing.T) {
	t.Parallel()

	inst := newInstanceForSemaphoreTest()

	if err := inst.sem.Acquire(context.Background(), 1); err != nil {
		t.Fatalf("failed to pre-acquire semaphore: %v", err)
	}
	defer inst.sem.Release(1)

	const deadline = 50 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()

	type result struct {
		err error
	}
	done := make(chan result, 1)
	start := time.Now()
	go func() {
		fn := &CELFunction{Name: "f", ID: "f", Return: types.IntType}
		_, err := inst.Call(ctx, fn, metadata.MD{})
		done <- result{err: err}
	}()

	select {
	case r := <-done:
		elapsed := time.Since(start)
		if !errors.Is(r.err, context.DeadlineExceeded) {
			t.Fatalf("expected context.DeadlineExceeded, got %v", r.err)
		}
		// Sanity check: should be roughly the deadline, not a hang.
		if elapsed > 2*time.Second {
			t.Fatalf("Call returned only after %v, expected ~%v", elapsed, deadline)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Call did not return within 2s after deadline; semaphore is not context-aware")
	}
}

// TestValidatePluginReleasesOnCtxCancel verifies the same context-aware
// behavior on ValidatePlugin's semaphore acquisition path.
func TestValidatePluginReleasesOnCtxCancel(t *testing.T) {
	t.Parallel()

	inst := newInstanceForSemaphoreTest()

	if err := inst.sem.Acquire(context.Background(), 1); err != nil {
		t.Fatalf("failed to pre-acquire semaphore: %v", err)
	}
	defer inst.sem.Release(1)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- inst.ValidatePlugin(ctx)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ValidatePlugin did not return within 2s after ctx cancel")
	}
}

// TestStartGCSkipsOnAcquireTimeout verifies that the GC code path skips when
// the semaphore is held by an in-flight call for longer than gcWaitForSemTimeout.
func TestStartGCSkipsOnAcquireTimeout(t *testing.T) {
	// Lower gcWaitForSemTimeout for the duration of this test so it runs quickly.
	t.Cleanup(func() { gcWaitForSemTimeout = 10 * time.Second })
	gcWaitForSemTimeout = 50 * time.Millisecond

	inst := newInstanceForSemaphoreTest()

	// Simulate a long-running call holding the semaphore.
	if err := inst.sem.Acquire(context.Background(), 1); err != nil {
		t.Fatalf("failed to pre-acquire semaphore: %v", err)
	}
	defer inst.sem.Release(1)

	done := make(chan error, 1)
	start := time.Now()
	go func() {
		done <- inst.startGC(context.Background())
	}()

	select {
	case err := <-done:
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("expected nil error on GC skip, got %v", err)
		}
		// startGC must have given up around gcWaitForSemTimeout, not blocked forever.
		if elapsed > 1*time.Second {
			t.Fatalf("startGC returned only after %v, expected ~%v", elapsed, gcWaitForSemTimeout)
		}
		// And it must NOT have called cancelFn or pushed to gcErrCh in the
		// skip path (those are reserved for the GC-hang path).
		select {
		case <-inst.gcErrCh:
			t.Fatal("gcErrCh should not receive a value when GC is skipped on Acquire timeout")
		default:
		}
	case <-time.After(2 * time.Second):
		t.Fatal("startGC did not return within 2s; expected skip after Acquire timeout")
	}
}
