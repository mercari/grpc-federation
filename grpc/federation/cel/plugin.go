package cel

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/mercari/grpc-federation/grpc/federation/cel/plugin"
)

type CELPlugin struct {
	cfg         CELPluginConfig
	mod         wazero.CompiledModule
	wasmRuntime wazero.Runtime
}

type CELFunction struct {
	Name     string
	ID       string
	Args     []*cel.Type
	Return   *cel.Type
	IsMethod bool
}

type CELPluginConfig struct {
	Name      string
	Wasm      WasmConfig
	Functions []*CELFunction
	CacheDir  string
}

type WasmConfig struct {
	Reader io.Reader
	Sha256 string
}

var (
	ErrWasmContentMismatch = errors.New(
		`grpc-federation: wasm file content mismatch`,
	)
)

type conn struct {
	net.Conn
	end bool
	buf []byte
	n   int64
	err error
}

type addr struct {
	net.Addr
}

var connMap = make(map[uint32]*conn)
var addrMap = make(map[uint32]*addr)

func NewCELPlugin(ctx context.Context, cfg CELPluginConfig) (*CELPlugin, error) {
	if cfg.Wasm.Reader == nil {
		return nil, fmt.Errorf("grpc-federation: WasmConfig.Reader field is required")
	}
	wasmFile, err := io.ReadAll(cfg.Wasm.Reader)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(wasmFile)
	gotHash := hex.EncodeToString(hash[:])
	if cfg.Wasm.Sha256 != gotHash {
		return nil, fmt.Errorf(`expected [%s] but got [%s]: %w`, cfg.Wasm.Sha256, gotHash, ErrWasmContentMismatch)
	}
	runtimeCfg := wazero.NewRuntimeConfig().WithCloseOnContextDone(true)
	if cache := getCompilationCache(cfg.Name, cfg.CacheDir); cache != nil {
		runtimeCfg = runtimeCfg.WithCompilationCache(cache)
	}

	r := wazero.NewRuntimeWithConfig(ctx, runtimeCfg)
	ptrToString := func(mod api.Module, ptr, size uint32) string {
		bytes, _ := mod.Memory().Read(ptr, size)
		return string(bytes)
	}

	ptrToBytes := func(mod api.Module, ptr, size uint32) []byte {
		bytes, _ := mod.Memory().Read(ptr, size)
		return bytes
	}

	goNet := r.NewHostModuleBuilder("wasi_go_net")
	goNet = goNet.NewFunctionBuilder().WithGoModuleFunction(
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			networkPtr := uint32(stack[0])
			networkLen := uint32(stack[1])
			addressPtr := uint32(stack[2])
			addressLen := uint32(stack[3])
			network := ptrToString(mod, networkPtr, networkLen)
			address := ptrToString(mod, addressPtr, addressLen)
			fmt.Println("network", network, "address", address)
			var d net.Dialer
			c, err := d.DialContext(ctx, network, address)
			if err != nil {
				fmt.Println("err", err)
			}
			//fmt.Println("conn", c)
			ret := &conn{Conn: c}
			id := uint32(len(connMap) + 1)
			connMap[id] = ret
			stack[0] = uint64(id)
		}),
		[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32},
		[]api.ValueType{api.ValueTypeI32},
	).Export("dial")
	goNet = goNet.NewFunctionBuilder().WithGoModuleFunction(
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			id := uint32(stack[0])
			conn, exists := connMap[id]
			if !exists {
				fmt.Println("failed to find conn")
				return
			}
			bytesPtr := uint32(stack[1])
			bytesLen := uint32(stack[2])
			b := ptrToBytes(mod, bytesPtr, bytesLen)
			conn.buf = make([]byte, len(b))
			conn.end = false
			go func() {
				n, err := conn.Read(conn.buf)
				conn.n = int64(n)
				conn.err = err
			}()
			/*
				if !mod.Memory().Write(uint32(stack[1]), conn.buf) {
					fmt.Println("failed to write buf")
				}
				if !mod.Memory().WriteUint32Le(uint32(stack[3]), uint32(n)) {
					fmt.Println("failed to write n")
				}
			*/
		}),
		[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32},
		[]api.ValueType{},
	).Export("conn_read")

	goNet = goNet.NewFunctionBuilder().WithGoModuleFunction(
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			id := uint32(stack[0])
			conn, exists := connMap[id]
			if !exists {
				fmt.Println("failed to find conn")
				return
			}
			if conn.err != nil {
				fmt.Println("err", conn.err)
				e := []byte(conn.err.Error())
				if !mod.Memory().Write(uint32(stack[4]), e) {
					fmt.Println("failed to write err content")
				}
				if !mod.Memory().WriteUint32Le(uint32(stack[5]), uint32(len(e))) {
					fmt.Println("failed to write err length")
				}
				fmt.Println("failed to read", err)
				return
			}
			if !mod.Memory().Write(uint32(stack[1]), conn.buf) {
				fmt.Println("failed to write buf")
			}
			if !mod.Memory().WriteUint32Le(uint32(stack[3]), uint32(conn.n)) {
				fmt.Println("failed to write n")
			}
		}),
		[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32},
		[]api.ValueType{},
	).Export("conn_get_read")

	goNet = goNet.NewFunctionBuilder().WithGoModuleFunction(
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			id := uint32(stack[0])
			conn, exists := connMap[id]
			if !exists {
				fmt.Println("failed to find conn")
				return
			}
			bytesPtr := uint32(stack[1])
			bytesLen := uint32(stack[2])
			b := ptrToBytes(mod, bytesPtr, bytesLen)
			n, err := conn.Write(b)
			if err != nil {
				e := []byte(err.Error())
				if !mod.Memory().Write(uint32(stack[4]), e) {
					fmt.Println("failed to write err content")
				}
				if !mod.Memory().WriteUint32Le(uint32(stack[5]), uint32(len(e))) {
					fmt.Println("failed to write err length")
				}
				fmt.Println("failed to write", err)
				return
			}
			if !mod.Memory().WriteUint32Le(uint32(stack[3]), uint32(n)) {
				fmt.Println("failed to write n")
			}
		}),
		[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32},
		[]api.ValueType{},
	).Export("conn_write")
	goNet = goNet.NewFunctionBuilder().WithGoModuleFunction(
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			id := uint32(stack[0])
			conn, exists := connMap[id]
			if !exists {
				fmt.Println("failed to find conn")
				return
			}
			if err := conn.Close(); err != nil {
				fmt.Println("close error", err)
			}
		}),
		[]api.ValueType{api.ValueTypeI32},
		[]api.ValueType{api.ValueTypeI32},
	).Export("conn_close")

	goNet = goNet.NewFunctionBuilder().WithGoModuleFunction(
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			id := uint32(stack[0])
			conn, exists := connMap[id]
			if !exists {
				fmt.Println("failed to find conn")
				return
			}
			t := time.Unix(int64(stack[1]), 0)
			if err := conn.SetDeadline(t); err != nil {
				fmt.Println("failed to set deadline", err)
			}
		}),
		[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32},
		[]api.ValueType{},
	).Export("conn_set_deadline")

	goNet = goNet.NewFunctionBuilder().WithGoModuleFunction(
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			id := uint32(stack[0])
			conn, exists := connMap[id]
			if !exists {
				fmt.Println("failed to find conn")
				return
			}
			t := time.Unix(int64(stack[1]), 0)
			if err := conn.SetReadDeadline(t); err != nil {
				fmt.Println("failed to set deadline", err)
			}
		}),
		[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32},
		[]api.ValueType{},
	).Export("conn_set_read_deadline")

	goNet = goNet.NewFunctionBuilder().WithGoModuleFunction(
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			id := uint32(stack[0])
			conn, exists := connMap[id]
			if !exists {
				fmt.Println("failed to find conn")
				return
			}
			t := time.Unix(int64(stack[1]), 0)
			if err := conn.SetWriteDeadline(t); err != nil {
				fmt.Println("failed to set deadline", err)
			}
		}),
		[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32},
		[]api.ValueType{},
	).Export("conn_set_write_deadline")

	goNet = goNet.NewFunctionBuilder().WithGoModuleFunction(
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			id := uint32(stack[0])
			conn, exists := connMap[id]
			if !exists {
				fmt.Println("failed to find conn")
				return
			}
			a := conn.LocalAddr()
			ret := &addr{Addr: a}
			retID := uint32(len(addrMap) + 1)
			addrMap[retID] = ret
			stack[0] = uint64(retID)
		}),
		[]api.ValueType{api.ValueTypeI32},
		[]api.ValueType{api.ValueTypeI32},
	).Export("conn_local_addr")

	goNet = goNet.NewFunctionBuilder().WithGoModuleFunction(
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			id := uint32(stack[0])
			conn, exists := connMap[id]
			if !exists {
				fmt.Println("failed to find conn")
				return
			}
			a := conn.RemoteAddr()
			ret := &addr{Addr: a}
			retID := uint32(len(addrMap) + 1)
			addrMap[retID] = ret
			stack[0] = uint64(retID)
		}),
		[]api.ValueType{api.ValueTypeI32},
		[]api.ValueType{api.ValueTypeI32},
	).Export("conn_remote_addr")

	goNet = goNet.NewFunctionBuilder().WithGoModuleFunction(
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
		}),
		[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32},
		[]api.ValueType{},
	).Export("addr_network")

	goNet = goNet.NewFunctionBuilder().WithGoModuleFunction(
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
		}),
		[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32},
		[]api.ValueType{},
	).Export("addr_string")

	if _, err := goNet.Instantiate(ctx); err != nil {
		return nil, fmt.Errorf("grpc-federation: failed to add wasi-network module: %w", err)
	}

	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	mod, err := r.CompileModule(ctx, wasmFile)
	if err != nil {
		return nil, fmt.Errorf("grpc-federation: failed to compile module: %w", err)
	}

	return &CELPlugin{
		cfg:         cfg,
		mod:         mod,
		wasmRuntime: r,
	}, nil
}

func getCompilationCache(name, baseDir string) wazero.CompilationCache {
	if baseDir == "" {
		tmpDir := os.TempDir()
		if tmpDir == "" {
			return nil
		}
		baseDir = tmpDir
	}
	cacheDir := filepath.Join(baseDir, "grpc-federation", name)
	if _, err := os.Stat(cacheDir); err != nil {
		if err := os.MkdirAll(cacheDir, 0o755); err != nil {
			return nil
		}
	}
	cache, err := wazero.NewCompilationCacheWithDir(cacheDir)
	if err != nil {
		return nil
	}
	return cache
}

func (p *CELPlugin) CreateInstance(ctx context.Context, celRegistry *types.Registry) *CELPluginInstance {
	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()
	modCfg := wazero.NewModuleConfig().
		WithSysWalltime().
		WithFSConfig(wazero.NewFSConfig().WithFSMount(os.DirFS("/"), "/")).
		WithStdin(stdinR).
		WithStdout(stdoutW).
		WithStderr(os.Stderr).
		WithArgs("plugin")
	for _, kv := range os.Environ() {
		i := strings.IndexByte(kv, '=')
		modCfg = modCfg.WithEnv(kv[:i], kv[i+1:])
	}

	// setting the buffer size to 1 ensures that the function can exit even if there is no receiver.
	instanceModErrCh := make(chan error, 1)
	go func() {
		_, err := p.wasmRuntime.InstantiateModule(ctx, p.mod, modCfg)
		instanceModErrCh <- err
	}()

	const gcQueueLength = 1

	gcQueue := make(chan struct{}, gcQueueLength)
	instance := &CELPluginInstance{
		name:             p.cfg.Name,
		functions:        p.cfg.Functions,
		celRegistry:      celRegistry,
		stdin:            stdinW,
		stdout:           stdoutR,
		instanceModErrCh: instanceModErrCh,
		gcQueue:          gcQueue,
	}
	// start GC thread.
	// It is enqueued into gcQueue using the `instance.GC()` function.
	go func() {
		for range gcQueue {
			instance.startGC()
		}
	}()
	return instance
}

type CELPluginInstance struct {
	name             string
	functions        []*CELFunction
	celRegistry      *types.Registry
	stdin            *io.PipeWriter
	stdout           *io.PipeReader
	instanceModErrCh chan error
	instanceModErr   error
	closed           bool
	mu               sync.Mutex
	gcQueue          chan struct{}
}

const PluginProtocolVersion = 1

type PluginVersionSchema struct {
	ProtocolVersion   int      `json:"protocolVersion"`
	FederationVersion string   `json:"grpcFederationVersion"`
	Functions         []string `json:"functions"`
}

var (
	versionCommand = "version\n"
	exitCommand    = "exit\n"
	gcCommand      = "gc\n"
)

func (i *CELPluginInstance) ValidatePlugin(ctx context.Context) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if err := i.write([]byte(versionCommand)); err != nil {
		return fmt.Errorf("failed to send cel protocol version command: %w", err)
	}
	content, err := i.recvContent()
	if err != nil {
		return fmt.Errorf("failed to receive cel protocol version command: %w", err)
	}
	var v PluginVersionSchema
	if err := json.Unmarshal([]byte(content), &v); err != nil {
		return fmt.Errorf("failed to decode cel plugin's version schema: %w", err)
	}
	if v.ProtocolVersion != PluginProtocolVersion {
		return fmt.Errorf(
			"grpc-federation: cel plugin protocol version mismatch: expected version %d but got %d. plugin's gRPC Federation version is %s",
			PluginProtocolVersion,
			v.ProtocolVersion,
			v.FederationVersion,
		)
	}
	implementedMethodMap := make(map[string]struct{})
	for _, fn := range v.Functions {
		implementedMethodMap[fn] = struct{}{}
	}

	var missingFunctions []string
	for _, fn := range i.functions {
		if _, exists := implementedMethodMap[fn.ID]; !exists {
			missingFunctions = append(missingFunctions, fn.ID)
		}
	}
	if len(missingFunctions) != 0 {
		return fmt.Errorf("grpc-federation: cel plugin functions are missing: [%v]", missingFunctions)
	}
	return nil
}

func (i *CELPluginInstance) write(cmd []byte) error {
	if i.closed {
		return i.instanceModErr
	}

	writeCh := make(chan error)
	go func() {
		_, err := i.stdin.Write(cmd)
		writeCh <- err
	}()
	select {
	case err := <-i.instanceModErrCh:
		// If the module instance is terminated,
		// it is considered that the termination process has been completed.
		i.closeResources(err)
		return err
	case err := <-writeCh:
		return err
	}
}

func (i *CELPluginInstance) Close(ctx context.Context) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	defer func() { i.closeResources(nil) }()
	if err := i.write([]byte(exitCommand)); err != nil {
		return err
	}
	close(i.gcQueue)
	return nil
}

func (i *CELPluginInstance) closeResources(instanceModErr error) {
	i.instanceModErr = instanceModErr
	i.closed = true
	i.stdin.Close()
	i.stdout.Close()
}

func (i *CELPluginInstance) LibraryName() string {
	return i.name
}

// Trigger GC command.
// The execution of GC is managed by queue, and if it exceeds the capacity of the queue, it will be ignored.
func (i *CELPluginInstance) GC() {
	i.enqueueGC()
}

func (i *CELPluginInstance) enqueueGC() {
	select {
	case i.gcQueue <- struct{}{}:
	default:
		// If the capacity of gcQueue is exceeded, the trigger event is discarded.
	}
}

func (i *CELPluginInstance) startGC() {
	i.mu.Lock()
	defer i.mu.Unlock()

	_ = i.write([]byte(gcCommand))
}

func (i *CELPluginInstance) CompileOptions() []cel.EnvOption {
	var opts []cel.EnvOption
	for _, fn := range i.functions {
		fn := fn
		if fn.IsMethod {
			opts = append(opts,
				BindMemberFunction(
					fn.Name,
					MemberOverloadFunc(fn.ID, fn.Args[0], fn.Args[1:], fn.Return, func(ctx context.Context, self ref.Val, args ...ref.Val) ref.Val {
						md, ok := metadata.FromIncomingContext(ctx)
						if !ok {
							md = make(metadata.MD)
						}
						return i.Call(ctx, fn, md, append([]ref.Val{self}, args...)...)
					}),
				)...,
			)
		} else {
			opts = append(opts,
				BindFunction(
					fn.Name,
					OverloadFunc(fn.ID, fn.Args, fn.Return, func(ctx context.Context, args ...ref.Val) ref.Val {
						md, ok := metadata.FromIncomingContext(ctx)
						if !ok {
							md = make(metadata.MD)
						}
						return i.Call(ctx, fn, md, args...)
					}),
				)...,
			)
		}
	}
	return opts
}

func (i *CELPluginInstance) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

func (i *CELPluginInstance) Call(ctx context.Context, fn *CELFunction, md metadata.MD, args ...ref.Val) ref.Val {
	i.mu.Lock()
	defer i.mu.Unlock()

	if err := i.sendRequest(fn, md, args...); err != nil {
		return types.NewErr(err.Error())
	}
	return i.recvResponse(fn)
}

func (i *CELPluginInstance) sendRequest(fn *CELFunction, md metadata.MD, args ...ref.Val) error {
	req := &plugin.CELPluginRequest{Method: fn.ID}
	for key, values := range md {
		req.Metadata = append(req.Metadata, &plugin.CELPluginGRPCMetadata{
			Key:    key,
			Values: values,
		})
	}
	for idx, arg := range args {
		pluginArg, err := i.refToCELPluginValue(fn.Args[idx], arg)
		if err != nil {
			return err
		}
		req.Args = append(req.Args, pluginArg)
	}

	encoded, err := protojson.Marshal(req)
	if err != nil {
		return err
	}
	if err := i.write(append(encoded, '\n')); err != nil {
		return err
	}
	return nil
}

func (i *CELPluginInstance) recvResponse(fn *CELFunction) ref.Val {
	content, err := i.recvContent()
	if err != nil {
		return types.NewErr(err.Error())
	}

	var res plugin.CELPluginResponse
	if err := protojson.Unmarshal([]byte(content), &res); err != nil {
		return types.NewErr(fmt.Sprintf("grpc-federation: failed to decode response: %s", err.Error()))
	}
	if res.Error != "" {
		return types.NewErr(res.Error)
	}
	return i.celPluginValueToRef(fn, fn.Return, res.Value)
}

func (i *CELPluginInstance) recvContent() (string, error) {
	if i.closed {
		return "", errors.New("grpc-federation: plugin has already been closed")
	}

	type readResult struct {
		response string
		err      error
	}
	readCh := make(chan readResult)
	go func() {
		reader := bufio.NewReader(i.stdout)
		content, err := reader.ReadString('\n')
		if err != nil {
			readCh <- readResult{err: fmt.Errorf("grpc-federation: failed to receive response from wasm plugin: %w", err)}
			return
		}
		if content == "" {
			readCh <- readResult{err: errors.New("grpc-federation: receive empty response from wasm plugin")}
			return
		}
		readCh <- readResult{response: content}
	}()
	select {
	case err := <-i.instanceModErrCh:
		// If the module instance is terminated,
		// it is considered that the termination process has been completed.
		i.closeResources(err)
		return "", err
	case result := <-readCh:
		return result.response, result.err
	}
}

func (i *CELPluginInstance) refToCELPluginValue(typ *cel.Type, v ref.Val) (*plugin.CELPluginValue, error) {
	switch typ.Kind() {
	case types.ListKind:
		elemType := typ.Parameters()[0]
		slice := reflect.ValueOf(v.Value())
		list := &plugin.CELPluginListValue{}
		for idx := 0; idx < slice.Len(); idx++ {
			src := slice.Index(idx).Interface()
			val := i.celRegistry.NativeToValue(src)
			if types.IsError(val) {
				return nil, fmt.Errorf("failed to convert %T to CEL value: %v", src, val.Value())
			}
			value, err := i.refToCELPluginValue(elemType, val)
			if err != nil {
				return nil, err
			}
			list.Values = append(list.Values, value)
		}
		return &plugin.CELPluginValue{
			Value: &plugin.CELPluginValue_List{
				List: list,
			},
		}, nil
	case types.BoolKind:
		vv := v.(types.Bool)
		return &plugin.CELPluginValue{
			Value: &plugin.CELPluginValue_Bool{
				Bool: bool(vv),
			},
		}, nil
	case types.DoubleKind:
		vv := v.(types.Double)
		return &plugin.CELPluginValue{
			Value: &plugin.CELPluginValue_Double{
				Double: float64(vv),
			},
		}, nil
	case types.IntKind:
		vv := v.(types.Int)
		return &plugin.CELPluginValue{
			Value: &plugin.CELPluginValue_Int64{
				Int64: int64(vv),
			},
		}, nil
	case types.BytesKind:
		vv := v.(types.Bytes)
		return &plugin.CELPluginValue{
			Value: &plugin.CELPluginValue_Bytes{
				Bytes: []byte(vv),
			},
		}, nil
	case types.StringKind:
		vv := v.(types.String)
		return &plugin.CELPluginValue{
			Value: &plugin.CELPluginValue_String_{
				String_: string(vv),
			},
		}, nil
	case types.UintKind:
		vv := v.(types.Uint)
		return &plugin.CELPluginValue{
			Value: &plugin.CELPluginValue_Uint64{
				Uint64: uint64(vv),
			},
		}, nil
	case types.StructKind:
		switch vv := v.Value().(type) {
		case proto.Message:
			any, ok := vv.(*anypb.Any)
			if !ok {
				anyValue, err := anypb.New(vv)
				if err != nil {
					return nil, fmt.Errorf("grpc-federation: failed to create any instance: %w", err)
				}
				any = anyValue
			}
			return &plugin.CELPluginValue{
				Value: &plugin.CELPluginValue_Message{
					Message: any,
				},
			}, nil
		default:
			return nil, fmt.Errorf(
				`grpc-federation: currently unsupported native proto message "%T"`,
				v.Value(),
			)
		}
	}
	return nil, fmt.Errorf(`grpc-federation: found unexpected cel function's argument type "%s"`, typ)
}

func (i *CELPluginInstance) celPluginValueToRef(fn *CELFunction, typ *cel.Type, v *plugin.CELPluginValue) ref.Val {
	switch typ.Kind() {
	case types.ListKind:
		elemType := typ.Parameters()[0]
		values := make([]ref.Val, 0, len(v.GetList().GetValues()))
		for _, vv := range v.GetList().GetValues() {
			value := i.celPluginValueToRef(fn, elemType, vv)
			if types.IsError(value) {
				// return error value
				return value
			}
			values = append(values, value)
		}
		return types.NewRefValList(i.celRegistry, values)
	case types.BoolKind:
		return types.Bool(v.GetBool())
	case types.BytesKind:
		return types.Bytes(v.GetBytes())
	case types.DoubleKind:
		return types.Double(v.GetDouble())
	case types.ErrorKind:
		return types.NewErr(v.GetString_())
	case types.IntKind:
		return types.Int(v.GetInt64())
	case types.StringKind:
		return types.String(v.GetString_())
	case types.UintKind:
		return types.Uint(v.GetUint64())
	case types.StructKind:
		return i.celRegistry.NativeToValue(v.GetMessage())
	}
	return types.NewErr("grpc-federation: unknown result type %s from %s function", typ, fn.Name)
}
