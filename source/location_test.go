package source_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mercari/grpc-federation/source"
)

func TestFile(t *testing.T) {
	path := filepath.Join("testdata", "service.proto")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sourceFile, err := source.NewFile(path, content)
	if err != nil {
		t.Fatal(err)
	}
	loc := &source.Location{
		FileName: filepath.Base(path),
		Message: &source.Message{
			Name: "GetPostResponse",
			Option: &source.MessageOption{
				Messages: &source.MessageDependencyOption{
					Idx:  0,
					Name: true,
				},
			},
		},
	}
	t.Run("filter", func(t *testing.T) {
		n := sourceFile.NodeInfoByLocation(loc)
		if n == nil {
			t.Fatal("failed to get node info")
		}
		if n.RawText() != `"post"` {
			t.Fatalf("failed to get text %s", n.RawText())
		}
	})
	t.Run("finder", func(t *testing.T) {
		foundLoc := sourceFile.FindLocationByPos(source.Position{
			Line: 25,
			Col:  17,
		})
		if diff := cmp.Diff(foundLoc, loc); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
}
