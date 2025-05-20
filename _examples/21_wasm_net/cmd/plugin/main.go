package main

import (
	"context"
	"io"
	"net/http"

	pluginpb "example/plugin"
)

type plugin struct{}

func (_ *plugin) Example_Net_HttpGet(ctx context.Context, url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func main() {
	pluginpb.RegisterNetPlugin(new(plugin))
}
