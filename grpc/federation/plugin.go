//go:build !wasip1

package federation

import "context"

type PluginHandler func(ctx context.Context, req *CELPluginRequest) (*CELPluginResponse, error)

func PluginMainLoop(verSchema CELPluginVersionSchema, handler PluginHandler) {}

func WritePluginContent(content []byte) {}

func ReadPluginContent() string {
	return ""
}
