package federation_test

import (
	"runtime/debug"
	"testing"

	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
)

func TestVersion(t *testing.T) {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		if buildInfo.Main.Version != "" {
			if grpcfed.Version != buildInfo.Main.Version {
				t.Fatalf("expected version is %s but got %s", buildInfo.Main.Version, grpcfed.Version)
			}
			return
		}
		if grpcfed.Version != "(devel)" {
			t.Fatalf("failed to get default version information: %s", grpcfed.Version)
		}
	}
}
