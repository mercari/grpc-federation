//go:build !wasip1

package net

import (
	"net"
	"net/http"
)

var (
	Listen = net.Listen
)

func DefaultTransport() http.RoundTripper {
	return http.DefaultTransport
}
