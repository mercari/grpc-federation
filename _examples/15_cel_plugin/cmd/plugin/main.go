package main

import (
	"context"
	pluginpb "example/plugin"
	"fmt"
	"os"
	"regexp"
	"unsafe"

	"google.golang.org/grpc/metadata"
)

type plugin struct{}

func (_ *plugin) Val() string {
	return "hello grpc-federation plugin"
}

func (_ *plugin) Example_Regexp_Compile(ctx context.Context, expr string) (*pluginpb.Regexp, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		fmt.Fprintf(os.Stderr, "plugin: got metadata is %+v", md)
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &pluginpb.Regexp{
		Ptr: uint32(uintptr(unsafe.Pointer(re))),
	}, nil
}

func (_ *plugin) Example_Regexp_Regexp_MatchString(ctx context.Context, re *pluginpb.Regexp, s string) (bool, error) {
	return (*regexp.Regexp)(unsafe.Pointer(uintptr(re.Ptr))).MatchString(s), nil
}
func main() {
	pluginpb.RegisterRegexpPlugin(&plugin{})
}
