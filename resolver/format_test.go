package resolver_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/genproto/googleapis/rpc/code"

	"github.com/mercari/grpc-federation/internal/testutil"
	"github.com/mercari/grpc-federation/resolver"
)

func TestProtoFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		messageOptionToFormatMap map[string]string
	}{
		{
			name: "simple_aggregation.proto",
			messageOptionToFormatMap: map[string]string{
				"GetPostResponse": `
  option (grpc.federation.message) = {
    def {
      name: "post"
      message {
        name: "Post"
        args { name: "id", by: "$.id" }
      }
    }
    def {
      name: "uuid"
      by: "grpc.federation.uuid.newRandom()"
    }
    def {
      name: "map_value"
      by: "{1:'a', 2:'b', 3:'c'}"
    }
    def {
      name: "e"
      enum {
        name: "org.federation.Item.ItemType"
        by: "org.user.Item.ItemType.value('ITEM_TYPE_2')"
      }
    }
    def {
      name: "id"
      by: "100"
    }
  }`,
				"Post": `
  option (grpc.federation.message) = {
    def {
      name: "res"
      call {
        method: "org.post.PostService/GetPost"
        request { field: "id", by: "$.id" }
      }
    }
    def {
      name: "post"
      autobind: true
      by: "res.post"
    }
    def {
      name: "user"
      message {
        name: "User"
        args { inline: "post" }
      }
    }
    def {
      name: "z"
      message {
        name: "Z"
      }
    }
    def {
      name: "m"
      autobind: true
      message {
        name: "M"
        args: [
          { name: "x", by: "10" },
          { name: "y", by: "1" }
        ]
      }
    }
  }`,
				"M": `
  option (grpc.federation.message) = {}`,
				"Z": `
  option (grpc.federation.message) = {
    custom_resolver: true
  }`,
				"User": `
  option (grpc.federation.message) = {
    def {
      name: "res"
      call {
        method: "org.user.UserService/GetUser"
        request { field: "id", by: "$.user_id" }
      }
    }
    def {
      name: "user"
      autobind: true
      by: "res.user"
    }
    def {
      name: "_def2"
      message {
        name: "M"
        args: [
          { name: "x", by: "uint(2)" },
          { name: "y", by: "org.user.Item.ItemType.value('ITEM_TYPE_2')" }
        ]
      }
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
    def {
      name: "cp"
      message {
        name: "CreatePost"
        args: [
          { name: "title", by: "$.title" },
          { name: "content", by: "$.content" },
          { name: "user_id", by: "$.user_id" },
          { name: "type", by: "$.type" }
        ]
      }
    }
    def {
      name: "res"
      call {
        method: "org.post.PostService/CreatePost"
        request { field: "post", by: "cp" }
      }
    }
    def {
      name: "p"
      by: "res.post"
    }
  }`,
				"CreatePost": `
  option (grpc.federation.message) = {
    alias: "org.post.CreatePost"
  }`,
			},
		},
		{
			name: "custom_resolver.proto",
			messageOptionToFormatMap: map[string]string{
				"GetPostResponse": `
  option (grpc.federation.message) = {
    def {
      name: "post"
      message {
        name: "Post"
        args { name: "id", by: "$.id" }
      }
    }
  }`,
				"Post": `
  option (grpc.federation.message) = {
    def {
      name: "res"
      call {
        method: "org.post.PostService/GetPost"
        request { field: "id", by: "$.id" }
      }
    }
    def {
      name: "post"
      autobind: true
      by: "res.post"
    }
    def {
      name: "user"
      message {
        name: "User"
        args { inline: "post" }
      }
    }
  }`,
				"User": `
  option (grpc.federation.message) = {
    def {
      name: "res"
      call {
        method: "org.user.UserService/GetUser"
        request { field: "id", by: "$.user_id" }
      }
    }
    def {
      name: "u"
      by: "res.user"
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
    def {
      name: "post"
      message {
        name: "Post"
        args: [
          { name: "id", by: "$.id" },
          { name: "a", by: "$.a" },
          { name: "b", by: "$.b" }
        ]
      }
    }
  }`,
				"Post": `
  option (grpc.federation.message) = {
    def {
      name: "res"
      call {
        method: "org.post.PostService/GetPost"
        request: [
          { field: "id", by: "$.id" },
          { field: "a", by: "$.a", if: "$.a != null" },
          { field: "b", by: "$.b", if: "$.b != null" }
        ]
      }
    }
    def {
      name: "post"
      autobind: true
      by: "res.post"
    }
  }`,
			},
		},
		{
			name: "validation.proto",
			messageOptionToFormatMap: map[string]string{
				"GetPostResponse": `
  option (grpc.federation.message) = {
    def {
      name: "post"
      message {
        name: "Post"
      }
    }
    def {
      name: "_def1"
      validation {
        error {
          code: FAILED_PRECONDITION
          if: "post.id != 'some-id'"
          message: "'validation message 1'"
        }
      }
    }
    def {
      name: "_def2"
      validation {
        error {
          code: FAILED_PRECONDITION
          if: "true"
          message: "'validation message 2'"
          details {
            if: "post.title != 'some-title'"
            message: [
              {...},
              {...}
            ]
            precondition_failure {...}
            bad_request {...}
            localized_message {...}
          }
        }
      }
    }
  }`,
			},
		},
		{
			name: "switch.proto",
			messageOptionToFormatMap: map[string]string{
				"GetPostResponse": `
  option (grpc.federation.message) = {
    def {
      name: "switch"
      switch {
        case {
          if: "$.id == 'blue'"
          by: "1"
        }
        case {
          if: "$.id == 'red'"
          by: "2"
        }
        default {
          by: "3"
        }
      }
    }
  }`,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testdataDir := filepath.Join(testutil.RepoRoot(), "testdata")
			r := resolver.New(testutil.Compile(t, filepath.Join(testdataDir, test.name)), resolver.ImportPathOption(testdataDir))
			result, err := r.Resolve()
			if err != nil {
				t.Fatal(err)
			}
			msgMap := make(map[string]*resolver.Message)
			for _, file := range result.Files {
				for _, msg := range file.Messages {
					msgMap[msg.Name] = msg
				}
			}
			for msgName, format := range test.messageOptionToFormatMap {
				t.Run(msgName, func(t *testing.T) {
					expected := strings.TrimPrefix(format, "\n")
					msg, exists := msgMap[msgName]
					if !exists {
						t.Fatalf("failed to find %s message", msgName)
					}
					got := msg.Rule.ProtoFormat(&resolver.ProtoFormatOption{
						IndentLevel:    1,
						IndentSpaceNum: 2,
					})
					if diff := cmp.Diff(got, expected); diff != "" {
						fmt.Println(got)
						t.Errorf("(-got, +want)\n%s", diff)
					}
				})
			}
		})
	}
}

func toCodePtr(c code.Code) *code.Code {
	return &c
}

func TestValidationExpr_ProtoFormat(t *testing.T) {
	t.Parallel()
	expr := &resolver.ValidationExpr{
		Name: "_validation0",
		Error: &resolver.GRPCError{
			Code: toCodePtr(code.Code_FAILED_PRECONDITION),
			Details: resolver.GRPCErrorDetails{
				{
					If: &resolver.CELValue{
						Expr: "1 == 1",
					},
				},
				{
					If: &resolver.CELValue{
						Expr: "2 == 2",
					},
				},
			},
		},
	}

	opt := resolver.DefaultProtoFormatOption
	got := expr.ProtoFormat(opt)
	if diff := cmp.Diff(got, `validation {
  name: "_validation0"
  error {
    code: FAILED_PRECONDITION
    details: [
      {
        if: "1 == 1"
      },
      {
        if: "2 == 2"
      }
    ]
  }
}`); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestGRPCError_ProtoFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc     string
		err      *resolver.GRPCError
		expected string
	}{
		{
			desc: "Rule is set",
			err: &resolver.GRPCError{
				Code: toCodePtr(code.Code_FAILED_PRECONDITION),
				If: &resolver.CELValue{
					Expr: "1 == 1",
				},
			},
			expected: `error {
  code: FAILED_PRECONDITION
  if: "1 == 1"
}`,
		},
		{
			desc: "Details are set",
			err: &resolver.GRPCError{
				Code: toCodePtr(code.Code_FAILED_PRECONDITION),
				Details: resolver.GRPCErrorDetails{
					{
						If: &resolver.CELValue{
							Expr: "1 == 1",
						},
					},
					{
						If: &resolver.CELValue{
							Expr: "2 == 2",
						},
					},
				},
			},
			expected: `error {
  code: FAILED_PRECONDITION
  details: [
    {
      if: "1 == 1"
    },
    {
      if: "2 == 2"
    }
  ]
}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			opt := resolver.DefaultProtoFormatOption
			got := tc.err.ProtoFormat(opt)
			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}
}

func TestGRPCErrorDetails_ProtoFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc     string
		details  resolver.GRPCErrorDetails
		expected string
	}{
		{
			desc: "single detail",
			details: resolver.GRPCErrorDetails{
				{
					If: &resolver.CELValue{
						Expr: "1 == 1",
					},
				},
			},
			expected: `details {
  if: "1 == 1"
}`,
		},
		{
			desc: "multiple details",
			details: resolver.GRPCErrorDetails{
				{
					If: &resolver.CELValue{
						Expr: "1 == 1",
					},
				},
				{
					If: &resolver.CELValue{
						Expr: "2 == 2",
					},
				},
			},
			expected: `details: [
  {
    if: "1 == 1"
  },
  {
    if: "2 == 2"
  }
]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			opt := resolver.DefaultProtoFormatOption
			got := tc.details.ProtoFormat(opt)
			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}
}

func TestGRPCErrorDetail_ProtoFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc     string
		detail   *resolver.GRPCErrorDetail
		expected string
	}{
		{
			desc: "single detail",
			detail: &resolver.GRPCErrorDetail{
				If: &resolver.CELValue{
					Expr: "1 == 1",
				},
				PreconditionFailures: []*resolver.PreconditionFailure{
					{},
				},
				BadRequests: []*resolver.BadRequest{
					{},
				},
				LocalizedMessages: []*resolver.LocalizedMessage{
					{},
				},
			},
			expected: `  if: "1 == 1"
  precondition_failure {...}
  bad_request {...}
  localized_message {...}`,
		},
		{
			desc: "multiple detail",
			detail: &resolver.GRPCErrorDetail{
				If: &resolver.CELValue{
					Expr: "2 == 2",
				},
				PreconditionFailures: []*resolver.PreconditionFailure{
					{},
					{},
				},
				BadRequests: []*resolver.BadRequest{
					{},
					{},
				},
				LocalizedMessages: []*resolver.LocalizedMessage{
					{},
					{},
				},
			},
			expected: `  if: "2 == 2"
  precondition_failure: [
    {...},
    {...}
  ]
  bad_request: [
    {...},
    {...}
  ]
  localized_message: [
    {...},
    {...}
  ]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			opt := resolver.DefaultProtoFormatOption
			got := tc.detail.ProtoFormat(opt)
			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}
}

func TestDependencyTreeFormat(t *testing.T) {
	t.Parallel()
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
			msgMap := make(map[string]*resolver.Message)
			for _, file := range result.Files {
				for _, msg := range file.Messages {
					msgMap[msg.Name] = msg
				}
			}
			for msgName, format := range test.messageOptionToFormatMap {
				t.Run(msgName, func(t *testing.T) {
					expected := strings.TrimPrefix(format, "\n")
					msg, exists := msgMap[msgName]
					if !exists {
						t.Fatalf("failed to find message from %s", msgName)
					}
					got := resolver.DependencyGraphTreeFormat(msg.Rule.DefSet.DefinitionGroups())
					if diff := cmp.Diff(got, expected); diff != "" {
						t.Errorf("(-got, +want)\n%s", diff)
					}
				})
			}
		})
	}
}
