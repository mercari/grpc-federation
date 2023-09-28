package resolver

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/mercari/grpc-federation/source"
)

func TestCandidates(t *testing.T) {
	fooMethod := &Method{
		Name: "Call",
		Request: &Message{
			Fields: []*Field{
				{Name: "foo_req_param", Type: Int64Type},
			},
		},
		Response: &Message{
			Fields: []*Field{
				{Name: "foo_res_param", Type: StringType},
			},
		},
	}
	barMethod := &Method{
		Name: "Call",
		Request: &Message{
			Fields: []*Field{
				{Name: "bar_req_param", Type: Int64Type},
			},
		},
		Response: &Message{
			Fields: []*Field{
				{Name: "bar_res_param", Type: StringType},
			},
		},
	}
	msgY := &Message{
		File: &File{Name: "foo.proto", Package: &Package{Name: "foo"}},
		Name: "y",
		Rule: &MessageRule{
			MessageArgument: &Message{
				Fields: []*Field{
					{Name: "y_arg0", Type: StringType},
					{Name: "y_arg1", Type: Int64Type},
				},
			},
		},
	}
	r := &Resolver{
		files: []*descriptorpb.FileDescriptorProto{
			{
				Name:    proto.String("foo.proto"),
				Package: proto.String("foo"),
			},
		},
		cachedMessageMap: map[string]*Message{
			"foo.x": {
				File: &File{Name: "foo.proto", Package: &Package{Name: "foo"}},
				Name: "x",
				Fields: []*Field{
					{Name: "field_a", Type: StringType},
				},
				Rule: &MessageRule{
					MessageArgument: &Message{
						Fields: []*Field{
							{Name: "x_arg0", Type: StringType},
							{Name: "x_arg1", Type: Int64Type},
						},
					},
					MethodCall: &MethodCall{
						Method: barMethod,
					},
					MessageDependencies: []*MessageDependency{
						{Name: "y", Message: msgY},
					},
				},
			},
			"foo.y": msgY,
			"foo.z": {
				File: &File{Name: "foo.proto", Package: &Package{Name: "foo"}},
				Name: "z",
			},
		},
		cachedMethodMap: map[string]*Method{
			"foo.FooExample/Call": fooMethod,
			"bar.BarExample/Call": barMethod,
		},
	}
	tests := []struct {
		name     string
		location *source.Location
		expected []string
	}{
		{
			name: "message_option",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name:   "x",
					Option: &source.MessageOption{},
				},
			},
			expected: []string{"resolver", "messages"},
		},
		{
			name: "resolver",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						Resolver: &source.ResolverOption{},
					},
				},
			},
			expected: []string{"method", "request", "response"},
		},
		{
			name: "resolver.method",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						Resolver: &source.ResolverOption{
							Method: true,
						},
					},
				},
			},
			expected: []string{"bar.BarExample/Call"},
		},
		{
			name: "resolver.request",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						Resolver: &source.ResolverOption{
							Request: &source.RequestOption{},
						},
					},
				},
			},
			expected: []string{"field", "by"},
		},
		{
			name: "resolver.request.field",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						Resolver: &source.ResolverOption{
							Request: &source.RequestOption{
								Field: true,
							},
						},
					},
				},
			},
			expected: []string{"bar_req_param"},
		},
		{
			name: "resolver.request.by",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						Resolver: &source.ResolverOption{
							Request: &source.RequestOption{
								By: true,
							},
						},
					},
				},
			},
			expected: []string{"$.x_arg1"},
		},
		{
			name: "resolver.response.name",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						Resolver: &source.ResolverOption{
							Response: &source.ResponseOption{
								Name: true,
							},
						},
					},
				},
			},
			expected: []string{},
		},
		{
			name: "resolver.response.field",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						Resolver: &source.ResolverOption{
							Response: &source.ResponseOption{
								Field: true,
							},
						},
					},
				},
			},
			expected: []string{"bar_res_param"},
		},
		{
			name: "resolver.response.autobind",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						Resolver: &source.ResolverOption{
							Response: &source.ResponseOption{
								AutoBind: true,
							},
						},
					},
				},
			},
			expected: []string{"true", "false"},
		},
		{
			name: "messages",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						Messages: &source.MessageDependencyOption{},
					},
				},
			},
			expected: []string{"name", "message", "args"},
		},
		{
			name: "messages.name",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						Messages: &source.MessageDependencyOption{Name: true},
					},
				},
			},
			expected: []string{},
		},
		{
			name: "messages.message",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						Messages: &source.MessageDependencyOption{Message: true},
					},
				},
			},
			expected: []string{"z"},
		},
		{
			name: "messages.args",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						Messages: &source.MessageDependencyOption{
							Args: &source.ArgumentOption{},
						},
					},
				},
			},
			expected: []string{"name", "by", "inline"},
		},
		{
			name: "messages.args.name",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						Messages: &source.MessageDependencyOption{
							Args: &source.ArgumentOption{Name: true},
						},
					},
				},
			},
			expected: []string{},
		},
		{
			name: "messages.args.by",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						Messages: &source.MessageDependencyOption{
							Args: &source.ArgumentOption{By: true},
						},
					},
				},
			},
			expected: []string{"$.x_arg0", "$.x_arg1"},
		},
		{
			name: "field",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Field: &source.Field{
						Name:   "field_a",
						Option: &source.FieldOption{By: true},
					},
				},
			},
			expected: []string{"$.x_arg0"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := r.Candidates(test.location)
			if diff := cmp.Diff(test.expected, got); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}
}
