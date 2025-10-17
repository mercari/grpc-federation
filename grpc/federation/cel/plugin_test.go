package cel

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/cel-go/common/types"
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
