package federation

import (
	_ "embed"
)

//go:embed federation.proto
var ProtoFile []byte

//go:embed private.proto
var PrivateProtoFile []byte
