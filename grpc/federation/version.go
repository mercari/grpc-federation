package federation

import (
	"github.com/mercari/grpc-federation/grpc/federation/cel"
)

const CELPluginProtocolVersion = cel.PluginProtocolVersion

type CELPluginVersionSchema = cel.PluginVersionSchema

var Version string = "dev"
