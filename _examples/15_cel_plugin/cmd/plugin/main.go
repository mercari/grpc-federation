package main

import (
	"context"
	pluginpb "example/plugin"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
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

func (_ *plugin) Example_Regexp_NewExample(_ context.Context) (*pluginpb.Example, error) {
	return &pluginpb.Example{}, nil
}

func (_ *plugin) Example_Regexp_NewExamples(_ context.Context) ([]*pluginpb.Example, error) {
	return []*pluginpb.Example{{}, {}}, nil
}

func (_ *plugin) Example_Regexp_FilterExamples(_ context.Context, v []*pluginpb.Example) ([]*pluginpb.Example, error) {
	return v, nil
}

func (_ *plugin) Example_Regexp_Example_Concat(_ context.Context, _ *pluginpb.Example, v []string) (string, error) {
	return strings.Join(v, ""), nil
}

func (_ *plugin) Example_Regexp_Example_MySplit(_ context.Context, _ *pluginpb.Example, s string, sep string) ([]string, error) {
	return strings.Split(s, sep), nil
}

func (_ *plugin) Example_Regexp_Block(ctx context.Context) (bool, error) {
	time.Sleep(100 * time.Second)
	return true, nil
}

func main() {
	pluginpb.RegisterRegexpPlugin(&plugin{})
}
