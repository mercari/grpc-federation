package federation

import (
	"runtime/debug"

	"github.com/mercari/grpc-federation/grpc/federation/cel"
)

const CELPluginProtocolVersion = cel.PluginProtocolVersion

type CELPluginVersionSchema = cel.PluginVersionSchema

const devVersion = "(devel)"

var Version string

func init() {
	if Version != "" {
		// set by go build with ldflags.
		return
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		// set by go install.
		Version = buildInfo.Main.Version
	}
	if Version == "" {
		Version = devVersion
	}
}
