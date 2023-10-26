package protodeps

import (
	_ "embed"
)

//go:embed googleapis/google/rpc/code.proto
var GoogleRPCCodeProtoFile []byte

//go:embed googleapis/google/rpc/error_details.proto
var GoogleRPCErrorDetailsProtoFile []byte
