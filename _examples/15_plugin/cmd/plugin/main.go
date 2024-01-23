package main

import (
	pluginpb "example/plugin"
	"regexp"
	"unsafe"
)

type plugin struct{}

func (_ *plugin) Val() string {
	return "hello grpc-federation plugin"
}

func (_ *plugin) Example_Regexp_Compile(expr string) (*pluginpb.Regexp, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &pluginpb.Regexp{
		Ptr: uint64(uintptr(unsafe.Pointer(re))),
	}, nil
}

func (_ *plugin) Example_Regexp_Regexp_MatchString(re *pluginpb.Regexp, s string) (bool, error) {
	return (*regexp.Regexp)(unsafe.Pointer(uintptr(re.Ptr))).MatchString(s), nil
}

func main() {
	pluginpb.RegisterRegexpPlugin(&plugin{})
}
