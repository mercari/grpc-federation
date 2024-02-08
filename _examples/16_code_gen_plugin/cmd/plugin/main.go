package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/mercari/grpc-federation/grpc/federation/generator"
)

//go:embed resolver.go.tmpl
var tmpl []byte

func main() {
	req, err := generator.ToCodeGeneratorRequest(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range req.GRPCFederationFiles {
		for _, svc := range file.Services {
			fmt.Println("service name", svc.Name)
			for _, method := range svc.Methods {
				fmt.Println("method name", method.Name)
				if method.Rule != nil {
					fmt.Println("method timeout", method.Rule.Timeout)
				}
			}
		}
		for _, msg := range file.Messages {
			fmt.Println("msg name", msg.Name)
		}
	}
	if err := os.WriteFile("resolver_test.go", tmpl, 0o600); err != nil {
		log.Fatal(err)
	}
}
