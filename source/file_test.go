package source_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mercari/grpc-federation/source"
)

func TestFile_FindLocationByPos(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc     string
		file     string
		pos      source.Position
		expected *source.Location
	}{
		{
			desc: "find nothing",
			file: "service.proto",
			pos: source.Position{
				Line: 21,
				Col:  1,
			},
			expected: nil,
		},
		{
			desc: "find a MethodOption string timeout",
			file: "service.proto",
			pos: source.Position{
				Line: 19,
				Col:  47,
			},
			expected: &source.Location{
				FileName: "testdata/service.proto",
				Service: &source.Service{
					Name: "FederationService",
					Method: &source.Method{
						Name:   "GetPost",
						Option: &source.MethodOption{Timeout: true},
					},
				},
			},
		},
		{
			desc: "find a MethodOption message timeout",
			file: "service.proto",
			pos: source.Position{
				Line: 23,
				Col:  16,
			},
			expected: &source.Location{
				FileName: "testdata/service.proto",
				Service: &source.Service{
					Name: "FederationService",
					Method: &source.Method{
						Name:   "GetPost2",
						Option: &source.MethodOption{Timeout: true},
					},
				},
			},
		},
		{
			desc: "find a CallExprOption method",
			file: "service.proto",
			pos: source.Position{
				Line: 51,
				Col:  20,
			},
			expected: &source.Location{
				FileName: "testdata/service.proto",
				Message: &source.Message{
					Name: "Post",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx:  0,
							Call: &source.CallExprOption{Method: true},
						},
					},
				},
			},
		},
		{
			desc: "find a CallExprOption request field",
			file: "service.proto",
			pos: source.Position{
				Line: 52,
				Col:  29,
			},
			expected: &source.Location{
				FileName: "testdata/service.proto",
				Message: &source.Message{
					Name: "Post",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Request: &source.RequestOption{Field: true},
							},
						},
					},
				},
			},
		},
		{
			desc: "find a CallExprOption request by",
			file: "service.proto",
			pos: source.Position{
				Line: 52,
				Col:  39,
			},
			expected: &source.Location{
				FileName: "testdata/service.proto",
				Message: &source.Message{
					Name: "Post",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Request: &source.RequestOption{By: true},
							},
						},
					},
				},
			},
		},
		{
			desc: "find a FieldOption by",
			file: "service.proto",
			pos: source.Position{
				Line: 42,
				Col:  47,
			},
			expected: &source.Location{
				FileName: "testdata/service.proto",
				Message: &source.Message{
					Name: "GetPostResponse",
					Field: &source.Field{
						Name:   "post",
						Option: &source.FieldOption{By: true},
					},
				},
			},
		},
		{
			desc: "find a MessageExprOption message",
			file: "service.proto",
			pos: source.Position{
				Line: 37,
				Col:  15,
			},
			expected: &source.Location{
				FileName: "testdata/service.proto",
				Message: &source.Message{
					Name: "GetPostResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Message: &source.MessageExprOption{
								Name: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "find a MessageExprOption arg name",
			file: "service.proto",
			pos: source.Position{
				Line: 38,
				Col:  24,
			},
			expected: &source.Location{
				FileName: "testdata/service.proto",
				Message: &source.Message{
					Name: "GetPostResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Message: &source.MessageExprOption{
								Args: &source.ArgumentOption{Name: true},
							},
						},
					},
				},
			},
		},
		{
			desc: "find a MessageExprOption arg by",
			file: "service.proto",
			pos: source.Position{
				Line: 38,
				Col:  34,
			},
			expected: &source.Location{
				FileName: "testdata/service.proto",
				Message: &source.Message{
					Name: "GetPostResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Message: &source.MessageExprOption{
								Args: &source.ArgumentOption{By: true},
							},
						},
					},
				},
			},
		},
		{
			desc: "find a MessageExprOption arg inline",
			file: "service.proto",
			pos: source.Position{
				Line: 60,
				Col:  26,
			},
			expected: &source.Location{
				FileName: "testdata/service.proto",
				Message: &source.Message{
					Name: "Post",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 2,
							Message: &source.MessageExprOption{
								Args: &source.ArgumentOption{Inline: true},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			path := filepath.Join("testdata", tc.file)
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			sourceFile, err := source.NewFile(path, content)
			if err != nil {
				t.Fatal(err)
			}
			got := sourceFile.FindLocationByPos(tc.pos)
			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}
}
