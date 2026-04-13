package federation

import (
	_ "embed"
)

//go:embed federation.proto
var ProtoFile []byte

//go:embed private.proto
var PrivateProtoFile []byte

//go:embed time.proto
var TimeProtoFile []byte

//go:embed plugin.proto
var PluginProtoFile []byte
