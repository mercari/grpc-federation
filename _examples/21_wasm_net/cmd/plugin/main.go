package main

import (
	"context"
	"crypto/tls"
	pluginpb "example/plugin"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
	"unsafe"
)

type plugin struct{}

//go:wasmimport wasi_go_net dial
func dial(networkPtr, networkLen, addressPtr, addressLen uint32) uint32

//go:wasmimport wasi_go_net conn_read
func connRead(connAddr, bytesAddr, bytesLen uint32)

//go:wasmimport wasi_go_net conn_get_read
func connGetRead(connAddr, bytesAddr, bytesLen, retN, retErrAddr, retErrLen uint32)

//go:wasmimport wasi_go_net conn_write
func connWrite(connAddr, bytesAddr, bytesLen, retN, retErrAddr, retErrLen uint32)

//go:wasmimport wasi_go_net conn_close
func connClose(connAddr uint32) uint32

//go:wasmimport wasi_go_net conn_set_deadline
func connSetDeadline(connAddr, t, retErrAddr, retErrLen uint32)

//go:wasmimport wasi_go_net conn_set_read_deadline
func connSetReadDeadline(connAddr, t, retErrAddr, retErrLen uint32)

//go:wasmimport wasi_go_net conn_set_write_deadline
func connSetWriteDeadline(connAddr, t, retErrAddr, retErrLen uint32)

//go:wasmimport wasi_go_net conn_local_addr
func connLocalAddr(connAddr uint32) uint32

//go:wasmimport wasi_go_net conn_remote_addr
func connRemoteAddr(connAddr uint32) uint32

//go:wasmimport wasi_go_net addr_network
func addrNetwork(addr, retPtr, retLen uint32)

//go:wasmimport wasi_go_net addr_string
func addrString(addr, retPtr, retLen uint32)

func stringToPtr(s string) (uint32, uint32) {
	ptr := unsafe.Pointer(unsafe.StringData(s))
	return uint32(uintptr(ptr)), uint32(len(s))
}

type conn struct {
	addr uint32
}

func (c *conn) Read(b []byte) (int, error) {
	bytesAddr := uint32(uintptr(unsafe.Pointer(&b[0])))
	bytesLen := uint32(len(b))
	//bytesAddr, bytesLen := stringToPtr(string(b))
	fmt.Fprintf(os.Stderr, "connRead\n")
	connRead(c.addr, bytesAddr, bytesLen)
	for {
		// nonblocking
		var (
			n       uint32
			errAddr uint32
			errLen  uint32
		)
		connGetRead(c.addr, bytesAddr, bytesLen,
			uint32(uintptr(unsafe.Pointer(&n))),
			uint32(uintptr(unsafe.Pointer(&errAddr))),
			uint32(uintptr(unsafe.Pointer(&errLen))),
		)
		if n != 0 {
			fmt.Fprintf(os.Stderr, "read: n = %d errAddr = %d. errLen = %d\n", n, errAddr, errLen)
			return int(n), nil
		}
		if errAddr != 0 {
			fmt.Fprintln(os.Stderr, "found error", errAddr)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return 0, nil
}

func (c *conn) Write(b []byte) (int, error) {
	bytesAddr := uint32(uintptr(unsafe.Pointer(&b[0])))
	bytesLen := uint32(len(b))
	var (
		n       uint32
		errAddr uint32
		errLen  uint32
	)
	connWrite(c.addr, bytesAddr, bytesLen,
		uint32(uintptr(unsafe.Pointer(&n))),
		uint32(uintptr(unsafe.Pointer(&errAddr))),
		uint32(uintptr(unsafe.Pointer(&errLen))),
	)
	fmt.Fprintf(os.Stderr, "write: n = %d, b = %v. errAddr = %d. errLen = %d\n", n, len(b), errAddr, errLen)
	return int(n), nil
}

func (c *conn) Close() error {
	connClose(c.addr)
	//fmt.Fprintln(os.Stderr, "called close")
	return nil
}

type addr struct {
	ptr uint32
}

func (a *addr) Network() string {
	var (
		addr   uint32
		length uint32
	)
	addrNetwork(a.ptr, uint32(uintptr(unsafe.Pointer(&addr))), uint32(uintptr(unsafe.Pointer(&length))))
	//fmt.Fprintf(os.Stderr, "addr.Network: addr: %d. len: %d\n", addr, length)
	ret := unsafe.String((*byte)(unsafe.Pointer(uintptr(addr))), length)
	fmt.Fprintln(os.Stderr, "addr.Netowk: ret", ret)
	return ret
}

func (a *addr) String() string {
	var (
		addr   uint32
		length uint32
	)
	addrString(a.ptr, uint32(uintptr(unsafe.Pointer(&addr))), uint32(uintptr(unsafe.Pointer(&length))))
	fmt.Fprintf(os.Stderr, "addr.String: addr: %d. len: %d\n", addr, length)
	ret := unsafe.String((*byte)(unsafe.Pointer(uintptr(addr))), length)
	fmt.Fprintln(os.Stderr, "addr.String: ret", ret)
	return ret
}

func (c *conn) LocalAddr() net.Addr {
	fmt.Fprintln(os.Stderr, "called LocalAddr")
	res := connLocalAddr(c.addr)
	return &addr{ptr: res}
}

func (c *conn) RemoteAddr() net.Addr {
	fmt.Fprintln(os.Stderr, "called RemoteAddr")
	res := connRemoteAddr(c.addr)
	return &addr{ptr: res}
}

func (c *conn) SetDeadline(t time.Time) error {
	//fmt.Fprintln(os.Stderr, "called SetDeadline", t)
	var (
		errAddr uint32
		errLen  uint32
	)
	connSetDeadline(c.addr,
		uint32(t.Unix()),
		uint32(uintptr(unsafe.Pointer(&errAddr))),
		uint32(uintptr(unsafe.Pointer(&errLen))),
	)
	return nil
}

func (c *conn) SetReadDeadline(t time.Time) error {
	//fmt.Fprintln(os.Stderr, "called SetReadDeadline", t)
	var (
		errAddr uint32
		errLen  uint32
	)
	connSetReadDeadline(c.addr,
		uint32(t.Unix()),
		uint32(uintptr(unsafe.Pointer(&errAddr))),
		uint32(uintptr(unsafe.Pointer(&errLen))),
	)
	return nil
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	//fmt.Fprintln(os.Stderr, "called SetWriteDeadline", t)
	var (
		errAddr uint32
		errLen  uint32
	)
	connSetWriteDeadline(c.addr,
		uint32(t.Unix()),
		uint32(uintptr(unsafe.Pointer(&errAddr))),
		uint32(uintptr(unsafe.Pointer(&errLen))),
	)
	return nil
}

func (_ *plugin) Example_Net_HttpGet(ctx context.Context, url string) (string, error) {
	http.DefaultTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(_ context.Context, network string, address string) (net.Conn, error) {
			//fmt.Fprintf(os.Stderr, "called dialcontext\n")
			networkPtr, networkLen := stringToPtr(network)
			addressPtr, addressLen := stringToPtr(address)
			connPtr := dial(networkPtr, networkLen, addressPtr, addressLen)
			return &conn{addr: connPtr}, nil
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       5 * time.Second,
		TLSHandshakeTimeout:   90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	net.DefaultResolver = &net.Resolver{
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			//fmt.Fprintln(os.Stderr, string(debug.Stack()))
			//fmt.Fprintf(os.Stderr, "called dial\n")
			networkPtr, networkLen := stringToPtr(network)
			addressPtr, addressLen := stringToPtr(address)
			connPtr := dial(networkPtr, networkLen, addressPtr, addressLen)
			return &conn{addr: connPtr}, nil
		},
	}
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	fmt.Fprintln(os.Stderr, "GOT RESPONSE!!!", resp)
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func main() {
	pluginpb.RegisterNetPlugin(&plugin{})
}
