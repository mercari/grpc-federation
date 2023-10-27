package rpc

import (
	_ "embed"
)

//go:embed code.proto
var GoogleRPCCodeProtoFile []byte

//go:embed error_details.proto
var GoogleRPCErrorDetailsProtoFile []byte
