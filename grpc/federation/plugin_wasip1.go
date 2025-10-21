//go:build wasip1

package federation

import (
	"context"
	"fmt"
	"runtime"
	"unsafe"

	"google.golang.org/grpc/metadata"

	grpcfedcel "github.com/mercari/grpc-federation/grpc/federation/cel"
)

//go:wasmimport grpcfederation write
func grpc_federation_write(ptr, size uint32)

//go:wasmimport grpcfederation read_length
func grpc_federation_read_length() uint32

//go:wasmimport grpcfederation read
func grpc_federation_read(uint32)

func writePluginContent(content []byte) {
	if content == nil {
		grpc_federation_write(0, 0)
		return
	}
	grpc_federation_write(
		uint32(uintptr(unsafe.Pointer(&content[0]))),
		uint32(len(content)),
	)
}

func readPluginContent() string {
	length := grpc_federation_read_length()
	if length == 0 {
		return ""
	}
	buf := make([]byte, length)
	grpc_federation_read(
		uint32(uintptr(unsafe.Pointer(&buf[0]))),
	)
	return string(buf)
}

type PluginHandler func(ctx context.Context, req *CELPluginRequest) (*CELPluginResponse, error)

func PluginMainLoop(verSchema CELPluginVersionSchema, handler PluginHandler) {
	for {
		content := readPluginContent()
		if content == grpcfedcel.ExitCommand {
			return
		}
		if content == grpcfedcel.GCCommand {
			runtime.GC()
			writePluginContent(nil)
			continue
		}
		if content == grpcfedcel.VersionCommand {
			b, _ := EncodeCELPluginVersion(verSchema)
			writePluginContent(b)
			continue
		}
		res, err := handlePluginFunc(content, handler)
		if err != nil {
			res = ToErrorCELPluginResponse(err)
		}
		encoded, err := EncodeCELPluginResponse(res)
		if err != nil {
			panic(fmt.Sprintf("failed to encode cel plugin response: %s", err.Error()))
		}
		writePluginContent(encoded)
	}
}

func handlePluginFunc(content string, handler PluginHandler) (res *CELPluginResponse, e error) {
	defer func() {
		if r := recover(); r != nil {
			res = ToErrorCELPluginResponse(fmt.Errorf("%v", r))
		}
	}()

	req, err := DecodeCELPluginRequest([]byte(content))
	if err != nil {
		return nil, err
	}
	md := make(metadata.MD)
	for _, m := range req.GetMetadata() {
		md[m.GetKey()] = m.GetValues()
	}
	ctx := metadata.NewIncomingContext(context.Background(), md)
	return handler(ctx, req)
}
