//go:build !wasip1

package net

import (
	"net"
	"net/http"
	"unsafe"
)

var (
	Listen = net.Listen
)

func StringToPtr(s string) uint32 {
	return uint32(uintptr(unsafe.Pointer(unsafe.StringData(s))))
}

func DefaultTransport() http.RoundTripper {
	return http.DefaultTransport
}
