package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"
	"unsafe"

	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
	"google.golang.org/grpc/metadata"

	"github.com/goccy/wasi-go-net/wasip1"

	_ "github.com/goccy/wasi-go-net/http"
)

type plugin struct{}

func (_ *plugin) Example_Net_HttpGet(ctx context.Context, url string) (string, error) {
	/*
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
	*/
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

//go:wasmimport wasi_go_net server_is_ready
//go:noescape
func server_is_ready(uint32, uint32)

func main() {
	http.DefaultTransport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		DialContext:           wasip1.DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       5 * time.Second,
		TLSHandshakeTimeout:   90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	resp, err := http.Get("https://example.com")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stderr, "GOT RESPONSE!!!", resp)
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

	ln, err := wasip1.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	addr := ln.Addr().String()
	fmt.Fprintf(os.Stderr, "addr = %s\n", addr)
	http.HandleFunc("/version", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Println("called version")
		b, _ := grpcfed.EncodeCELPluginVersion(grpcfed.CELPluginVersionSchema{
			ProtocolVersion:   grpcfed.CELPluginProtocolVersion,
			FederationVersion: "(devel)",
			Functions: []string{
				"example_net_httpGet_string_string",
			},
		})
		w.Write(b)
	})
	ch := make(chan struct{})
	http.HandleFunc("/exit", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		ch <- struct{}{}
	})
	http.HandleFunc("/gc", func(w http.ResponseWriter, _ *http.Request) {
		runtime.GC()
		w.WriteHeader(http.StatusOK)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		b, err := io.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		res, err := handleNetPlugin(b, new(plugin))
		if err != nil {
			res = grpcfed.ToErrorCELPluginResponse(err)
		}
		encoded, err := grpcfed.EncodeCELPluginResponse(res)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(encoded)
	})
	server := http.Server{}
	go func() {
		if err := server.Serve(ln); err != nil {
			panic(err)
		}
	}()
	for {
		if _, err := http.Get(fmt.Sprintf("http://%s", addr)); err == nil {
			ptr := unsafe.Pointer(unsafe.StringData(addr))
			server_is_ready(uint32(uintptr(ptr)), uint32(len(addr)))
			break
		}
	}

	<-ch
}

func handleNetPlugin(content []byte, plug *plugin) (res *grpcfed.CELPluginResponse, e error) {
	defer func() {
		if e := recover(); e != nil {
			res = grpcfed.ToErrorCELPluginResponse(errors.New(fmt.Sprintf("%v", e)))
		}
	}()

	req, err := grpcfed.DecodeCELPluginRequest(content)
	if err != nil {
		return nil, err
	}
	md := make(metadata.MD)
	for _, m := range req.GetMetadata() {
		md[m.GetKey()] = m.GetValues()
	}
	ctx := metadata.NewIncomingContext(context.Background(), md)
	switch req.GetMethod() {
	case "example_net_httpGet_string_string":
		if len(req.GetArgs()) != 1 {
			return nil, fmt.Errorf("%s: invalid argument number: %d. expected number is %d", req.GetMethod(), len(req.GetArgs()), 1)
		}
		arg0, err := grpcfed.ToString(req.GetArgs()[0])
		if err != nil {
			return nil, err
		}
		ret, err := plug.Example_Net_HttpGet(ctx, arg0)
		if err != nil {
			return nil, err
		}
		return grpcfed.ToStringCELPluginResponse(ret)
	}
	return nil, fmt.Errorf("unexpected method name: %s", req.GetMethod())
}
