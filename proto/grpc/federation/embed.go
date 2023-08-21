package federation

import (
	_ "embed"
)

//go:embed federation.proto
var ProtoFile []byte
