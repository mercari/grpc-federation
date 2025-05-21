package net

import (
	"fmt"
	"io"
	"net/http"
	"unsafe"
)

//go:wasmimport wasi_go_net server_is_ready
//go:noescape
func server_is_ready(uint32, uint32)

func ReadRequestBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()

	return io.ReadAll(r.Body)
}

func Error(w http.ResponseWriter, e error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(e.Error()))
}

func WaitForReadyPluginServer(plgServerAddr string) {
	for {
		if _, err := http.Get(fmt.Sprintf("http://%s", plgServerAddr)); err == nil {
			server_is_ready(
				uint32(uintptr(unsafe.Pointer(unsafe.StringData(plgServerAddr)))),
				uint32(len(plgServerAddr)),
			)
			break
		}
	}
}
