package generator_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mercari/grpc-federation/generator"
	"github.com/mercari/grpc-federation/internal/testutil"
	"github.com/mercari/grpc-federation/resolver"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"simple_aggregation"},
		{"minimum"},
		{"create_post"},
		{"custom_resolver"},
		{"async"},
		{"alias"},
		{"autobind"},
		{"literal"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := resolver.New(testutil.Compile(t, filepath.Join(testutil.RepoRoot(), "testdata", test.name+".proto")))
			result, err := r.Resolve()
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Services) != 1 {
				t.Fatalf("faield to get services. expected 1 but got %d", len(result.Services))
			}
			service := result.Services[0]
			out, err := generator.NewCodeGenerator().Generate(service)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join("testdata", fmt.Sprintf("expected_%s.go", test.name))
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(string(out), string(data)); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}
}
