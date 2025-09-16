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
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/goccy/wasi-go/imports"
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
	cfg            CELPluginConfig
	mod            wazero.CompiledModule
	wasmRuntime    wazero.Runtime
	instance       *CELPluginInstance
	backupInstance chan *CELPluginInstance
	instanceMu     sync.RWMutex
	celRegistry    *types.Registry
}

type CELFunction struct {
	Name     string
	ID       string
	Args     []*cel.Type
	Return   *cel.Type
	IsMethod bool
}

type CELPluginConfig struct {
	Name       string
	Wasm       WasmConfig
	Functions  []*CELFunction
	CacheDir   string
	Capability *CELPluginCapability
}

type CELPluginCapability struct {
	Env        *CELPluginEnvCapability
	FileSystem *CELPluginFileSystemCapability
	Network    *CELPluginNetworkCapability
}

type CELPluginEnvCapability struct {
	All   bool
	Names []string
}

type CELPluginFileSystemCapability struct {
	MountPath string
}

type CELPluginNetworkCapability struct {
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
	var runtimeCfg wazero.RuntimeConfig
	if cfg.CacheDir == "" {
		runtimeCfg = wazero.NewRuntimeConfigInterpreter()
		if cache := getCompilationCache(cfg.Name, cfg.CacheDir); cache != nil {
			runtimeCfg = runtimeCfg.WithCompilationCache(cache)
		}
	} else {
		runtimeCfg = wazero.NewRuntimeConfig()
		if cache := getCompilationCache(cfg.Name, cfg.CacheDir); cache != nil {
			runtimeCfg = runtimeCfg.WithCompilationCache(cache)
		}
	}
	r := wazero.NewRuntimeWithConfig(ctx, runtimeCfg.WithDebugInfoEnabled(false))
	mod, err := r.CompileModule(ctx, wasmFile)
	if err != nil {
		return nil, err
	}
	if cfg.Capability == nil || cfg.Capability.Network == nil {
		wasi_snapshot_preview1.MustInstantiate(ctx, r)
	}

	ret := &CELPlugin{
		cfg:            cfg,
		mod:            mod,
		wasmRuntime:    r,
		backupInstance: make(chan *CELPluginInstance, 1),
	}

	host := r.NewHostModuleBuilder("grpcfederation")
	host.NewFunctionBuilder().WithGoModuleFunction(
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			instance, err := ret.getInstance(ctx)
			if err != nil {
				panic(err)
			}
			outW := instance.outW
			//nolint:gosec
			b, ok := mod.Memory().Read(uint32(stack[0]), uint32(stack[1]))
			if !ok {
				panic("failed to read memory from plugin")
			}
			if _, err := outW.Write(b); err != nil {
				panic(fmt.Sprintf("failed to write plugin response to host buffer: %s", err))
			}
		}),
		[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32},
		[]api.ValueType{},
	).Export("grpc_federation_write")
	if _, err := host.Instantiate(ctx); err != nil {
		return nil, err
	}
	return ret, nil
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

func (p *CELPlugin) Close() {
	p.instanceMu.Lock()
	p.instance.Close()
	p.instanceMu.Unlock()

	backup := <-p.backupInstance
	backup.Close()
}

// If we treat a CELPluginInstance directly as a CELLibrary, any CELProgram once created will be cached and reused,
// so the CELPluginInstance that was initially bound to the CELProgram will be captured and continually used.
// Therefore, in order to switch the CELPluginInstance during operation, we need to treat the CELPlugin as a CELLibrary.
func (p *CELPlugin) CompileOptions() []cel.EnvOption {
	var opts []cel.EnvOption
	for _, fn := range p.cfg.Functions {
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
						instance, err := p.getInstance(ctx)
						if err != nil {
							return types.NewErr(err.Error())
						}
						return instance.Call(ctx, fn, md, append([]ref.Val{self}, args...)...)
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
						instance, err := p.getInstance(ctx)
						if err != nil {
							return types.NewErr(err.Error())
						}
						return instance.Call(ctx, fn, md, args...)
					}),
				)...,
			)
		}
	}
	return opts
}

func (p *CELPlugin) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

// CreateInstance is called when initializing the gRPC Federation service.
// In addition to calling ValidatePlugin to check the plugin version, it also creates a backupInstance to switch to in case an error occurs in the main instance.
func (p *CELPlugin) CreateInstance(ctx context.Context, celRegistry *types.Registry) (*CELPluginInstance, error) {
	p.celRegistry = celRegistry
	instance, err := p.createInstance(ctx)
	if err != nil {
		return nil, err
	}
	p.instanceMu.Lock()
	p.instance = instance
	p.instanceMu.Unlock()

	p.createBackupInstance(ctx)
	return instance, nil
}

func (p *CELPlugin) Cleanup() {
	if instance := p.getCurrentInstance(); instance != nil {
		instance.enqueueGC()
	}
}

func (p *CELPlugin) createBackupInstance(ctx context.Context) {
	go func() {
		instance, err := p.createInstance(ctx)
		if err != nil {
			p.backupInstance <- nil
		} else {
			p.backupInstance <- instance
		}
	}()
}

func (p *CELPlugin) getInstance(ctx context.Context) (*CELPluginInstance, error) {
	if instance := p.getCurrentInstance(); instance != nil {
		return instance, nil
	}

	instance := <-p.backupInstance
	if instance == nil {
		return nil, errors.New("grpc-federation: failed to get backup cel plugin instance")
	}

	p.instanceMu.Lock()
	p.instance = instance
	p.instanceMu.Unlock()

	p.createBackupInstance(ctx)
	return instance, nil
}

func (p *CELPlugin) getCurrentInstance() *CELPluginInstance {
	p.instanceMu.RLock()
	defer p.instanceMu.RUnlock()

	if p.instance != nil && !p.instance.closed {
		return p.instance
	}
	return nil
}

func (p *CELPlugin) createInstance(ctx context.Context) (*CELPluginInstance, error) {
	inR, inW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	outR, outW, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	modCfg, networkModCfg := addModuleConfigByCapability(
		p.cfg.Capability,
		wazero.NewModuleConfig().
			WithSysWalltime().
			WithStdin(inR).
			WithStdout(os.Stdout).
			WithStderr(os.Stderr).
			WithArgs("plugin"),
		imports.NewBuilder().
			WithStdio(int(inR.Fd()), int(os.Stdout.Fd()), int(os.Stderr.Fd())).
			WithSocketsExtension("wasmedgev2", p.mod),
	)

	if p.cfg.Capability != nil && p.cfg.Capability.Network != nil {
		var err error
		ctx, _, err = networkModCfg.Instantiate(ctx, p.wasmRuntime)
		if err != nil {
			return nil, err
		}
	}

	const gcQueueLength = 1
	instance := &CELPluginInstance{
		name:             p.cfg.Name,
		functions:        p.cfg.Functions,
		celRegistry:      p.celRegistry,
		inW:              inW,
		outR:             outR,
		outW:             outW,
		gcQueue:          make(chan struct{}, gcQueueLength),
		instanceModErrCh: make(chan error, 1),
	}

	// setting the buffer size to 1 ensures that the function can exit even if there is no receiver.
	go func() {
		_, err := p.wasmRuntime.InstantiateModule(ctx, p.mod, modCfg)
		instance.instanceModErrCh <- err
	}()

	// start GC thread.
	// It is enqueued into gcQueue using the `instance.GC()` function.
	go func() {
		for range instance.gcQueue {
			instance.startGC()
		}
	}()
	return instance, nil
}

func addModuleConfigByCapability(capability *CELPluginCapability, cfg wazero.ModuleConfig, nwcfg *imports.Builder) (wazero.ModuleConfig, *imports.Builder) {
	cfg, nwcfg = addModuleConfigByEnvCapability(capability, cfg, nwcfg)
	cfg, nwcfg = addModuleConfigByFileSystemCapability(capability, cfg, nwcfg)
	return cfg, nwcfg
}

var ignoreEnvNameMap = map[string]struct{}{
	// If a value greater than 1 is passed to GOMAXPROCS, a panic occurs on the plugin side,
	// so make sure not to pass it explicitly.
	"GOMAXPROCS": {},
}

func addModuleConfigByEnvCapability(capability *CELPluginCapability, cfg wazero.ModuleConfig, nwcfg *imports.Builder) (wazero.ModuleConfig, *imports.Builder) {
	if capability == nil || capability.Env == nil {
		return cfg, nwcfg
	}

	type Env struct {
		key   string
		value string
	}

	envCfg := capability.Env
	srcEnvs := os.Environ()
	envs := make([]Env, 0, len(srcEnvs))
	envMap := make(map[string]Env)
	for _, kv := range srcEnvs {
		i := strings.IndexByte(kv, '=')
		key := kv[:i]
		value := kv[i+1:]
		if _, exists := ignoreEnvNameMap[key]; exists {
			continue
		}
		env := Env{key: key, value: value}
		envs = append(envs, env)
		envMap[key] = env
	}
	if envCfg.All {
		filteredAllEnvs := make([]string, 0, len(envs))
		for _, env := range envs {
			cfg = cfg.WithEnv(env.key, env.value)
			filteredAllEnvs = append(filteredAllEnvs, env.key+"="+env.value)
		}
		nwcfg = nwcfg.WithEnv(filteredAllEnvs...)
	} else {
		var filteredEnvs []string
		for _, name := range envCfg.Names {
			envName := strings.ToUpper(name)
			if env, exists := envMap[envName]; exists {
				cfg = cfg.WithEnv(env.key, env.value)
				filteredEnvs = append(filteredEnvs, env.key+"="+env.value)
			}
		}
		nwcfg = nwcfg.WithEnv(filteredEnvs...)
	}
	return cfg, nwcfg
}

func addModuleConfigByFileSystemCapability(capability *CELPluginCapability, cfg wazero.ModuleConfig, nwcfg *imports.Builder) (wazero.ModuleConfig, *imports.Builder) {
	if capability == nil || capability.FileSystem == nil {
		return cfg, nwcfg
	}
	fs := capability.FileSystem
	mountPath := "/"
	if fs.MountPath != "" {
		mountPath = fs.MountPath
	}
	return cfg.WithFSConfig(
		wazero.NewFSConfig().WithFSMount(os.DirFS(mountPath), ""),
	), nwcfg.WithDirs(mountPath)
}

type CELPluginInstance struct {
	name             string
	functions        []*CELFunction
	celRegistry      *types.Registry
	inW              *os.File
	outR             *os.File
	outW             *os.File
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
		_, err := i.inW.Write(cmd)
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

func (i *CELPluginInstance) Close() error {
	if i == nil {
		return nil
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	defer func() { i.closeResources(nil) }()

	if i.closed {
		return i.instanceModErr
	}

	// start termination process.
	_, _ = i.inW.WriteString(exitCommand)
	<-i.instanceModErrCh

	return nil
}

func (i *CELPluginInstance) closeResources(instanceModErr error) {
	i.instanceModErr = instanceModErr
	i.closed = true
	i.inW.Close()
	i.outR.Close()
	i.outW.Close()
	close(i.gcQueue)
}

func (i *CELPluginInstance) LibraryName() string {
	return i.name
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
		reader := bufio.NewReader(i.outR)
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
