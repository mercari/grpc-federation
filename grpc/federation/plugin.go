//go:build !wasip1

package federation

func WritePluginContent(content []byte) {}

func ReadPluginContent() string {
	return ""
}
