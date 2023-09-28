package resolver_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mercari/grpc-federation/internal/testutil"
	"github.com/mercari/grpc-federation/resolver"
)

func TestProtoFormat(t *testing.T) {
	tests := []struct {
		name                     string
		messageOptionToFormatMap map[string]string
	}{
		{
			name: "simple_aggregation.proto",
			messageOptionToFormatMap: map[string]string{
				"GetPostResponse": `
  option (grpc.federation.message) = {
    messages {
      name: "post"
      message: "Post"
      args { name: "id", by: "$.id" }
    }
  }`,
				"Post": `
  option (grpc.federation.message) = {
    resolver: {
      method: "org.post.PostService/GetPost"
      request { field: "id", by: "$.id" }
      response { name: "post", field: "post", autobind: true }
    }
    messages: [
      {
        name: "user"
        message: "User"
        args { inline: "post" }
      },
      {
        name: "z"
        message: "Z"
      },
      {
        name: "m"
        message: "M"
        autobind: true
      }
    ]
  }`,
				"M": `
  option (grpc.federation.message) = {}`,
				"Z": `
  option (grpc.federation.message) = {
    custom_resolver: true
  }`,
				"User": `
  option (grpc.federation.message) = {
    resolver: {
      method: "org.user.UserService/GetUser"
      request { field: "id", by: "$.user_id" }
      response { name: "user", field: "user", autobind: true }
    }
  }`,
			},
		},
		{
			name: "minimum.proto",
			messageOptionToFormatMap: map[string]string{
				"GetPostResponse": `
  option (grpc.federation.message) = {
    custom_resolver: true
  }`,
			},
		},
		{
			name: "create_post.proto",
			messageOptionToFormatMap: map[string]string{
				"CreatePostResponse": `
  option (grpc.federation.message) = {
    resolver: {
      method: "org.post.PostService/CreatePost"
      request { field: "post", by: "cp" }
      response { name: "p", field: "post" }
    }
    messages {
      name: "cp"
      message: "CreatePost"
      args: [
        { name: "title", by: "$.title" },
        { name: "content", by: "$.content" },
        { name: "user_id", by: "$.user_id" }
      ]
    }
  }`,
				"CreatePost": `
  option (grpc.federation.message) = {}`,
			},
		},
		{
			name: "custom_resolver.proto",
			messageOptionToFormatMap: map[string]string{
				"GetPostResponse": `
  option (grpc.federation.message) = {
    messages {
      name: "post"
      message: "Post"
      args { name: "id", by: "$.id" }
    }
  }`,
				"Post": `
  option (grpc.federation.message) = {
    resolver: {
      method: "org.post.PostService/GetPost"
      request { field: "id", by: "$.id" }
      response { name: "post", field: "post", autobind: true }
    }
    messages {
      name: "user"
      message: "User"
      args { inline: "post" }
    }
  }`,
				"User": `
  option (grpc.federation.message) = {
    resolver: {
      method: "org.user.UserService/GetUser"
      request { field: "id", by: "$.user_id" }
      response { name: "u", field: "user" }
    }
    custom_resolver: true
  }`,
			},
		},
		{
			name: "alias.proto",
			messageOptionToFormatMap: map[string]string{
				"GetPostResponse": `
  option (grpc.federation.message) = {
    messages {
      name: "post"
      message: "Post"
      args { name: "id", by: "$.id" }
    }
  }`,
				"Post": `
  option (grpc.federation.message) = {
    resolver: {
      method: "org.post.PostService/GetPost"
      request { field: "id", by: "$.id" }
      response { name: "post", field: "post", autobind: true }
    }
  }`,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := resolver.New(testutil.Compile(t, filepath.Join(testutil.RepoRoot(), "testdata", test.name)))
			result, err := r.Resolve()
			if err != nil {
				t.Fatal(err)
			}
			for _, svc := range result.Services {
				for _, msg := range svc.Messages {
					t.Run(msg.Name, func(t *testing.T) {
						expected, exists := test.messageOptionToFormatMap[msg.Name]
						if !exists {
							t.Fatalf("failed to find %s message option data", msg.Name)
						}
						expected = strings.TrimPrefix(expected, "\n")
						got := msg.Rule.ProtoFormat(&resolver.ProtoFormatOption{
							IndentLevel:    1,
							IndentSpaceNum: 2,
						})
						if diff := cmp.Diff(got, expected); diff != "" {
							t.Errorf("(-got, +want)\n%s", diff)
						}
					})
				}
			}
		})
	}
}

func TestDependencyTreeFormat(t *testing.T) {
	tests := []struct {
		name                     string
		messageOptionToFormatMap map[string]string
	}{
		{
			name: "async.proto",
			messageOptionToFormatMap: map[string]string{
				"GetResponse": `
a ─┐
   c ─┐
b ─┐  │
   d ─┤
      e ─┐
a ─┐     │
   c ─┐  │
b ─┐  │  │
   d ─┤  │
      f ─┤
      g ─┤
         h ─┐
      i ─┐  │
         j ─┤
`,
				"A": `
aa ─┐
ab ─┤
`,
				"AA": "",
				"AB": "",
				"B":  "",
				"C":  "",
				"D":  "",
				"E":  "",
				"F":  "",
				"G":  "",
				"H":  "",
				"I":  "",
				"J":  "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := resolver.New(testutil.Compile(t, filepath.Join(testutil.RepoRoot(), "testdata", test.name)))
			result, err := r.Resolve()
			if err != nil {
				t.Fatal(err)
			}
			for _, svc := range result.Services {
				for _, msg := range svc.Messages {
					t.Run(msg.Name, func(t *testing.T) {
						expected, exists := test.messageOptionToFormatMap[msg.Name]
						if !exists {
							t.Fatalf("failed to find %s message option data", msg.Name)
						}
						expected = strings.TrimPrefix(expected, "\n")
						got := resolver.DependencyGraphTreeFormat(msg.Rule.Resolvers)
						if diff := cmp.Diff(got, expected); diff != "" {
							t.Errorf("(-got, +want)\n%s", diff)
						}
					})
				}
			}
		})
	}
}
