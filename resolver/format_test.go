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
      args: [
        { name: "id", by: "$.id" },
        { name: "double_value", double: 1.23 },
        { name: "float_value", float: 4.56 },
        { name: "int32_value", int32: 1 },
        { name: "int64_value", int64: 2 },
        { name: "uint32_value", uint32: 3 },
        { name: "uint64_value", uint64: 4 },
        { name: "sint32_value", sint32: -1 },
        { name: "sint64_value", sint64: -2 },
        { name: "fixed32_value", fixed32: 5 },
        { name: "fixed64_value", fixed64: 6 },
        { name: "sfixed32_value", sfixed32: 7 },
        { name: "sfixed64_value", sfixed64: 8 },
        { name: "bool_value", bool: true },
        { name: "string_value", string: "hello" },
        { name: "bytes_value", bytes: "hello" },
        { name: "message_value", message: { name: "org.post.Post", fields: [{ field: "content", string: "xxxyyyzzz" }] } },
        { name: "double_list_value", double_list: { values: [1.23, 4.56] } },
        { name: "float_list_value", float_list: { values: [7.89, 1.23] } },
        { name: "int32_list_value", int32_list: { values: [1, 2, 3] } },
        { name: "int64_list_value", int64_list: { values: [4, 5, 6] } },
        { name: "uint32_list_value", uint32_list: { values: [7, 8, 9] } },
        { name: "uint64_list_value", uint64_list: { values: [10, 11, 12] } },
        { name: "sint32_list_value", sint32_list: { values: [-1, -2, -3] } },
        { name: "sint64_list_value", sint64_list: { values: [-4, -5, -6] } },
        { name: "fixed32_list_value", fixed32_list: { values: [11, 12, 13] } },
        { name: "fixed64_list_value", fixed64_list: { values: [14, 15, 16] } },
        { name: "sfixed32_list_value", sfixed32_list: { values: [-1, -2, -3] } },
        { name: "sfixed64_list_value", sfixed64_list: { values: [-4, -5, -6] } },
        { name: "bool_list_value", bool_list: { values: [false, true] } },
        { name: "string_list_value", string_list: { values: ["hello", "world"] } },
        { name: "bytes_list_value", bytes_list: { values: ["hello", "world"] } },
        { name: "message_list_value", message_list: { values: [{ name: "org.post.Post", fields: [{ field: "content", string: "aaabbbccc" }] }] } }
      ]
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
