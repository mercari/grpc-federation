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
					DefSet: &VariableDefinitionSet{
						Defs: []*VariableDefinition{
							{
								Expr: &VariableExpr{
									Call: &CallExpr{
										Method: barMethod,
									},
								},
							},
							{
								Name: "y",
								Expr: &VariableExpr{
									Message: &MessageExpr{
										Message: msgY,
									},
								},
							},
						},
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
			expected: []string{"def", "custom_resolver", "alias"},
		},
		{
			name: "def.call",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						VariableDefinitions: &source.VariableDefinitionOption{
							Call: &source.CallExprOption{},
						},
					},
				},
			},
			expected: []string{"method", "request", "timeout", "retry"},
		},
		{
			name: "def.call.method",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						VariableDefinitions: &source.VariableDefinitionOption{
							Call: &source.CallExprOption{
								Method: true,
							},
						},
					},
				},
			},
			expected: []string{"bar.BarExample/Call"},
		},
		{
			name: "def.call.request",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						VariableDefinitions: &source.VariableDefinitionOption{
							Call: &source.CallExprOption{
								Request: &source.RequestOption{},
							},
						},
					},
				},
			},
			expected: []string{"field", "by"},
		},
		{
			name: "def.call.request.field",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						VariableDefinitions: &source.VariableDefinitionOption{
							Call: &source.CallExprOption{
								Request: &source.RequestOption{
									Field: true,
								},
							},
						},
					},
				},
			},
			expected: []string{"bar_req_param"},
		},
		{
			name: "def.call.request.by",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						VariableDefinitions: &source.VariableDefinitionOption{
							Call: &source.CallExprOption{
								Request: &source.RequestOption{
									By: true,
								},
							},
						},
					},
				},
			},
			expected: []string{"$.x_arg1"},
		},
		{
			name: "def.message",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						VariableDefinitions: &source.VariableDefinitionOption{
							Message: &source.MessageExprOption{},
						},
					},
				},
			},
			expected: []string{"name", "args"},
		},
		{
			name: "def.message.name",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						VariableDefinitions: &source.VariableDefinitionOption{
							Message: &source.MessageExprOption{
								Name: true,
							},
						},
					},
				},
			},
			expected: []string{"y", "z"},
		},
		{
			name: "def.message.args",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						VariableDefinitions: &source.VariableDefinitionOption{
							Message: &source.MessageExprOption{
								Args: &source.ArgumentOption{},
							},
						},
					},
				},
			},
			expected: []string{"name", "by", "inline"},
		},
		{
			name: "def.message.args.name",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						VariableDefinitions: &source.VariableDefinitionOption{
							Message: &source.MessageExprOption{
								Args: &source.ArgumentOption{Name: true},
							},
						},
					},
				},
			},
			expected: []string{},
		},
		{
			name: "def.message.args.by",
			location: &source.Location{
				FileName: "foo.proto",
				Message: &source.Message{
					Name: "x",
					Option: &source.MessageOption{
						VariableDefinitions: &source.VariableDefinitionOption{
							Message: &source.MessageExprOption{
								Args: &source.ArgumentOption{By: true},
							},
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
			expected: []string{"y", "$.x_arg0"},
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
