//go:build wasip1

package net

import (
	"crypto/tls"
	"net/http"
	"unsafe"

	"github.com/goccy/wasi-go-net/wasip1"
)

var (
	Listen = wasip1.Listen
)

func StringToPtr(s string) uint32 {
	return uint32(uintptr(unsafe.Pointer(unsafe.StringData(s))))
}

func DefaultTransport() http.RoundTripper {
	return &http.Transport{
		DialContext: wasip1.DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}
