//go:build update_testdata

package generator_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mercari/grpc-federation/generator"
	"github.com/mercari/grpc-federation/internal/testutil"
	"github.com/mercari/grpc-federation/resolver"
)

// TestUpdateTestdata regenerates expected_*.go files for the listed test cases.
// Run with: go test ./generator/... -run TestUpdateTestdata -tags update_testdata
func TestUpdateTestdata(t *testing.T) {
	cases := []string{"optional"}
	for _, name := range cases {
		testdataDir := filepath.Join(testutil.RepoRoot(), "testdata")
		files := testutil.Compile(t, filepath.Join(testdataDir, name+".proto"))

		var dependentFiles []string
		for _, f := range files {
			fileName := f.GetName()
			if filepath.IsAbs(fileName) {
				dependentFiles = append(dependentFiles, fileName)
				continue
			}
			if !strings.Contains(fileName, "/") {
				dependentFiles = append(dependentFiles, fileName)
				continue
			}
		}

		r := resolver.New(files, resolver.ImportPathOption(testdataDir))
		result, err := r.Resolve()
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(result.Files))
		}
		result.Files[0].Name = filepath.Base(result.Files[0].Name)
		out, err := generator.NewCodeGenerator().Generate(result.Files[0])
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join("testdata", fmt.Sprintf("expected_%s.go", name))
		if err := os.WriteFile(path, out, 0644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %d bytes to %s", len(out), path)
	}
}
