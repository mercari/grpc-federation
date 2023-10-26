package rpc

import (
	_ "embed"
)

//go:embed code.proto
var CodeProtoFile []byte

//go:embed error_details.proto
var ErrorDetailsProtoFile []byte
