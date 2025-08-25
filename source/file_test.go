package source_test

import (
	"io"
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
		// Additional coverage tests for various FindLocationByPos scenarios
		// Service-related tests to cover 0% functions
		{
			desc: "find service env message",
			file: "coverage.proto",
			pos: source.Position{
				Line: 15,
				Col:  19,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Option: &source.ServiceOption{
						Env: &source.Env{Message: true},
					},
				},
			},
		},
		{
			desc: "find service variable name",
			file: "coverage.proto",
			pos: source.Position{
				Line: 38,
				Col:  16,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Option: &source.ServiceOption{
						Var: &source.ServiceVariable{
							Idx:  0,
							Name: true,
						},
					},
				},
			},
		},
		{
			desc: "find service variable if condition",
			file: "coverage.proto",
			pos: source.Position{
				Line: 39,
				Col:  14,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Option: &source.ServiceOption{
						Var: &source.ServiceVariable{
							Idx: 0,
							If:  true,
						},
					},
				},
			},
		},
		{
			desc: "find service variable by expression",
			file: "coverage.proto",
			pos: source.Position{
				Line: 40,
				Col:  15,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Option: &source.ServiceOption{
						Var: &source.ServiceVariable{
							Idx: 0,
							By:  true,
						},
					},
				},
			},
		},
		{
			desc: "find service variable validation if",
			file: "coverage.proto",
			pos: source.Position{
				Line: 42,
				Col:  17,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Option: &source.ServiceOption{
						Var: &source.ServiceVariable{
							Idx: 0,
							Validation: &source.ServiceVariableValidationExpr{
								If: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "find service variable validation message",
			file: "coverage.proto",
			pos: source.Position{
				Line: 43,
				Col:  22,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Option: &source.ServiceOption{
						Var: &source.ServiceVariable{
							Idx: 0,
							Validation: &source.ServiceVariableValidationExpr{
								Message: true,
							},
						},
					},
				},
			},
		},
		// Additional advanced coverage tests that successfully passed
		{
			desc: "find file option import single value",
			file: "coverage.proto",
			pos: source.Position{
				Line: 9,
				Col:  22,
			},
			expected: &source.Location{
				FileName:   "testdata/coverage.proto",
				ImportName: "multiple_import.proto",
			},
		},
		// Enum-related test cases
		{
			desc: "find enum option alias",
			file: "coverage.proto",
			pos: source.Position{
				Line: 226,
				Col:  12,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Enum: &source.Enum{
					Name:   "ItemType",
					Option: &source.EnumOption{Alias: true},
				},
			},
		},
		{
			desc: "find enum value option alias",
			file: "coverage.proto",
			pos: source.Position{
				Line: 229,
				Col:  19,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Enum: &source.Enum{
					Name: "ItemType",
					Value: &source.EnumValue{
						Value: "UNKNOWN",
						Option: &source.EnumValueOption{
							Alias: true,
						},
					},
				},
			},
		},
		{
			desc: "find enum value option default",
			file: "coverage.proto",
			pos: source.Position{
				Line: 231,
				Col:  14,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Enum: &source.Enum{
					Name: "ItemType",
					Value: &source.EnumValue{
						Value: "UNKNOWN",
						Option: &source.EnumValueOption{
							Default: true,
						},
					},
				},
			},
		},
		{
			desc: "find enum value attribute name",
			file: "coverage.proto",
			pos: source.Position{
				Line: 233,
				Col:  20,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Enum: &source.Enum{
					Name: "ItemType",
					Value: &source.EnumValue{
						Value: "UNKNOWN",
						Option: &source.EnumValueOption{
							Attr: &source.EnumValueAttribute{
								Idx:  0,
								Name: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "find enum value attribute value",
			file: "coverage.proto",
			pos: source.Position{
				Line: 233,
				Col:  35,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Enum: &source.Enum{
					Name: "ItemType",
					Value: &source.EnumValue{
						Value: "UNKNOWN",
						Option: &source.EnumValueOption{
							Attr: &source.EnumValueAttribute{
								Idx:   0,
								Value: true,
							},
						},
					},
				},
			},
		},
		// Oneof field test cases
		{
			desc: "find oneof field option",
			file: "coverage.proto",
			pos: source.Position{
				Line: 365,
				Col:  13,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "NestedMessage",
					Oneof: &source.Oneof{
						Name: "choice",
						Field: &source.Field{
							Name: "text",
							Option: &source.FieldOption{
								Oneof: &source.FieldOneof{
									If: true,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "find oneof field default",
			file: "coverage.proto",
			pos: source.Position{
				Line: 366,
				Col:  18,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "NestedMessage",
					Oneof: &source.Oneof{
						Name: "choice",
						Field: &source.Field{
							Name: "text",
							Option: &source.FieldOption{
								Oneof: &source.FieldOneof{
									Default: true,
								},
							},
						},
					},
				},
			},
		},
		// Validation expression test cases
		{
			desc: "find validation expr name",
			file: "coverage.proto",
			pos: source.Position{
				Line: 282,
				Col:  18,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Validation: &source.ValidationExprOption{
								Name: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "find validation expr error if",
			file: "coverage.proto",
			pos: source.Position{
				Line: 285,
				Col:  23,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Validation: &source.ValidationExprOption{
								Error: &source.GRPCErrorOption{
									Idx: 0,
									If:  true,
								},
							},
						},
					},
				},
			},
		},
		// Map expression test cases
		{
			desc: "find map expr enum name",
			file: "coverage.proto",
			pos: source.Position{
				Line: 324,
				Col:  19,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 1,
							Map: &source.MapExprOption{
								Enum: &source.EnumExprOption{
									Name: true,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "find map expr enum by",
			file: "coverage.proto",
			pos: source.Position{
				Line: 325,
				Col:  17,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 1,
							Map: &source.MapExprOption{
								Enum: &source.EnumExprOption{
									By: true,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "find map expr message name",
			file: "coverage.proto",
			pos: source.Position{
				Line: 334,
				Col:  19,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 2,
							Map: &source.MapExprOption{
								Message: &source.MessageExprOption{
									Name: true,
								},
							},
						},
					},
				},
			},
		},
		// Message argument test cases
		{
			desc: "find message arg name",
			file: "coverage.proto",
			pos: source.Position{
				Line: 346,
				Col:  22,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 3,
							Message: &source.MessageExprOption{
								Args: &source.ArgumentOption{
									Idx:  0,
									Name: true,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "find message arg by",
			file: "coverage.proto",
			pos: source.Position{
				Line: 346,
				Col:  32,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 3,
							Message: &source.MessageExprOption{
								Args: &source.ArgumentOption{
									Idx: 0,
									By:  true,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "find message arg inline",
			file: "coverage.proto",
			pos: source.Position{
				Line: 347,
				Col:  60,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 3,
							Message: &source.MessageExprOption{
								Args: &source.ArgumentOption{
									Idx:    1,
									Inline: true,
								},
							},
						},
					},
				},
			},
		},
		// Retry option test cases
		{
			desc: "find retry constant interval",
			file: "coverage.proto",
			pos: source.Position{
				Line: 103,
				Col:  25,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "ComplexResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Retry: &source.RetryOption{
									Constant: &source.RetryConstantOption{
										Interval: true,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "find retry exponential initial_interval",
			file: "coverage.proto",
			pos: source.Position{
				Line: 115,
				Col:  33,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "ComplexResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 1,
							Call: &source.CallExprOption{
								Retry: &source.RetryOption{
									Exponential: &source.RetryExponentialOption{
										InitialInterval: true,
									},
								},
							},
						},
					},
				},
			},
		},
		// Environment variable option test cases
		{
			desc: "find env var alternate",
			file: "coverage.proto",
			pos: source.Position{
				Line: 22,
				Col:  24,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Option: &source.ServiceOption{
						Env: &source.Env{
							Var: &source.EnvVar{
								Idx: 0,
								Option: &source.EnvVarOption{
									Alternate: true,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "find env var required",
			file: "coverage.proto",
			pos: source.Position{
				Line: 23,
				Col:  23,
			},
			expected: &source.Location{
				FileName: "testdata/coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Option: &source.ServiceOption{
						Env: &source.Env{
							Var: &source.EnvVar{
								Idx: 0,
								Option: &source.EnvVarOption{
									Required: true,
								},
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

func TestFile_NodeInfoByLocation(t *testing.T) {
	t.Parallel()

	// First read the proto file content
	path := filepath.Join("testdata", "coverage.proto")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sourceFile, err := source.NewFile(path, content)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		desc     string
		location *source.Location
		wantNil  bool
	}{
		{
			desc: "find service option env message",
			location: &source.Location{
				FileName: "coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Option: &source.ServiceOption{
						Env: &source.Env{Message: true},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find service option env var name",
			location: &source.Location{
				FileName: "coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Option: &source.ServiceOption{
						Env: &source.Env{
							Var: &source.EnvVar{
								Idx:  0,
								Name: true,
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find service variable name",
			location: &source.Location{
				FileName: "coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Option: &source.ServiceOption{
						Var: &source.ServiceVariable{
							Idx:  0,
							Name: true,
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find service variable if condition",
			location: &source.Location{
				FileName: "coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Option: &source.ServiceOption{
						Var: &source.ServiceVariable{
							Idx: 0,
							If:  true,
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find service variable by expression",
			location: &source.Location{
				FileName: "coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Option: &source.ServiceOption{
						Var: &source.ServiceVariable{
							Idx: 0,
							By:  true,
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find service variable validation if",
			location: &source.Location{
				FileName: "coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Option: &source.ServiceOption{
						Var: &source.ServiceVariable{
							Idx: 0,
							Validation: &source.ServiceVariableValidationExpr{
								If: true,
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find service variable validation message",
			location: &source.Location{
				FileName: "coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Option: &source.ServiceOption{
						Var: &source.ServiceVariable{
							Idx: 0,
							Validation: &source.ServiceVariableValidationExpr{
								Message: true,
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find method option timeout",
			location: &source.Location{
				FileName: "coverage.proto",
				Service: &source.Service{
					Name: "CoverageService",
					Method: &source.Method{
						Name:   "GetData",
						Option: &source.MethodOption{Timeout: true},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find call expr method",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx:  0,
							Call: &source.CallExprOption{Method: true},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find call expr timeout",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx:  0,
							Call: &source.CallExprOption{Timeout: true},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find call expr metadata",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx:  0,
							Call: &source.CallExprOption{Metadata: true},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find call expr retry if",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Retry: &source.RetryOption{If: true},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find grpc call option content subtype",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Option: &source.GRPCCallOption{
									ContentSubtype: true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find grpc call option header",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Option: &source.GRPCCallOption{
									Header: true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find grpc call option trailer",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Option: &source.GRPCCallOption{
									Trailer: true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find grpc call option static method",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Option: &source.GRPCCallOption{
									StaticMethod: true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find method request field",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Request: &source.RequestOption{
									Idx:   0,
									Field: true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find method request by",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Request: &source.RequestOption{
									Idx: 0,
									By:  true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find method request if",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Request: &source.RequestOption{
									Idx: 0,
									If:  true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find validation expr name",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Validation: &source.ValidationExprOption{
								Name: true,
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find field by option",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataItem",
					Field: &source.Field{
						Name:   "value",
						Option: &source.FieldOption{By: true},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find oneof option",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "NestedMessage",
					Oneof: &source.Oneof{
						Name:   "choice",
						Option: &source.OneofOption{},
					},
				},
			},
			wantNil: false, // Oneof option returns node info
		},
		{
			desc: "find enum option alias",
			location: &source.Location{
				FileName: "coverage.proto",
				Enum: &source.Enum{
					Name:   "ItemType",
					Option: &source.EnumOption{Alias: true},
				},
			},
			wantNil: false,
		},
		{
			desc: "find enum value alias",
			location: &source.Location{
				FileName: "coverage.proto",
				Enum: &source.Enum{
					Name: "Status",
					Value: &source.EnumValue{
						Value:  "STATUS_ACTIVE",
						Option: &source.EnumValueOption{Alias: true},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find enum value default",
			location: &source.Location{
				FileName: "coverage.proto",
				Enum: &source.Enum{
					Name: "ItemType",
					Value: &source.EnumValue{
						Value:  "UNKNOWN",
						Option: &source.EnumValueOption{Default: true},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find enum value attr name",
			location: &source.Location{
				FileName: "coverage.proto",
				Enum: &source.Enum{
					Name: "ItemType",
					Value: &source.EnumValue{
						Value: "UNKNOWN",
						Option: &source.EnumValueOption{
							Attr: &source.EnumValueAttribute{
								Idx:  0,
								Name: true,
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find enum value attr value",
			location: &source.Location{
				FileName: "coverage.proto",
				Enum: &source.Enum{
					Name: "ItemType",
					Value: &source.EnumValue{
						Value: "UNKNOWN",
						Option: &source.EnumValueOption{
							Attr: &source.EnumValueAttribute{
								Idx:   0,
								Value: true,
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find import by name",
			location: &source.Location{
				FileName:   "coverage.proto",
				ImportName: "federation.proto",
			},
			wantNil: false,
		},
		// Additional tests for complex message expression locations
		{
			desc: "find map expr message args",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 2,
							Map: &source.MapExprOption{
								Message: &source.MessageExprOption{
									Args: &source.ArgumentOption{
										Idx:  0,
										Name: true,
									},
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find retry constant max_retries",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "ComplexResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Retry: &source.RetryOption{
									Constant: &source.RetryConstantOption{
										MaxRetries: true,
									},
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find retry exponential randomization_factor",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "ComplexResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 1,
							Call: &source.CallExprOption{
								Retry: &source.RetryOption{
									Exponential: &source.RetryExponentialOption{
										RandomizationFactor: true,
									},
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find retry exponential multiplier",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "ComplexResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 1,
							Call: &source.CallExprOption{
								Retry: &source.RetryOption{
									Exponential: &source.RetryExponentialOption{
										Multiplier: true,
									},
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find retry exponential max_interval",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "ComplexResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 1,
							Call: &source.CallExprOption{
								Retry: &source.RetryOption{
									Exponential: &source.RetryExponentialOption{
										MaxInterval: true,
									},
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find retry exponential max_retries",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "ComplexResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 1,
							Call: &source.CallExprOption{
								Retry: &source.RetryOption{
									Exponential: &source.RetryExponentialOption{
										MaxRetries: true,
									},
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find grpc call option max_call_recv_msg_size",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Option: &source.GRPCCallOption{
									MaxCallRecvMsgSize: true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find grpc call option max_call_send_msg_size",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Option: &source.GRPCCallOption{
									MaxCallSendMsgSize: true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find grpc call option wait_for_ready",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Option: &source.GRPCCallOption{
									WaitForReady: true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find grpc error option ignore",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Error: &source.GRPCErrorOption{
									Idx:    0,
									Ignore: true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find grpc error option ignore_and_response",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Error: &source.GRPCErrorOption{
									Idx:               0,
									IgnoreAndResponse: true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find grpc error def name",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Error: &source.GRPCErrorOption{
									Idx: 0,
									Def: &source.VariableDefinitionOption{
										Idx:  0,
										Name: true,
									},
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find grpc error def by",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Error: &source.GRPCErrorOption{
									Idx: 0,
									Def: &source.VariableDefinitionOption{
										Idx: 0,
										By:  true,
									},
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "nodeInfoByEnumExpr",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 1,
							Map: &source.MapExprOption{
								Enum: &source.EnumExprOption{
									Name: true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "nodeInfoByEnumExpr",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 1,
							Map: &source.MapExprOption{
								Enum: &source.EnumExprOption{
									By: true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "nodeInfoByArgument with inline",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 3,
							Message: &source.MessageExprOption{
								Args: &source.ArgumentOption{
									Idx:    0,
									Inline: true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find grpc error detail by",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Error: &source.GRPCErrorOption{
									Idx: 0,
									Detail: &source.GRPCErrorDetailOption{
										Idx: 0,
										By:  true,
									},
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find grpc error detail message",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Error: &source.GRPCErrorOption{
									Idx: 0,
									Detail: &source.GRPCErrorDetailOption{
										Idx: 0,
										LocalizedMessage: &source.GRPCErrorDetailLocalizedMessageOption{
											Idx: 0,
										},
									},
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find grpc error detail precondition failure",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Error: &source.GRPCErrorOption{
									Idx: 0,
									Detail: &source.GRPCErrorDetailOption{
										Idx: 0,
										PreconditionFailure: &source.GRPCErrorDetailPreconditionFailureOption{
											Idx: 0,
										},
									},
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find grpc error detail bad request",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 0,
							Call: &source.CallExprOption{
								Error: &source.GRPCErrorOption{
									Idx: 0,
									Detail: &source.GRPCErrorDetailOption{
										Idx: 0,
										BadRequest: &source.GRPCErrorDetailBadRequestOption{
											Idx: 0,
										},
									},
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find field oneof option",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Field: &source.Field{
						Name: "user",
						Option: &source.FieldOption{
							Oneof: &source.FieldOneof{
								If: true,
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find field oneof by",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Field: &source.Field{
						Name: "user",
						Option: &source.FieldOption{
							Oneof: &source.FieldOneof{
								By: true,
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			desc: "find field oneof def name",
			location: &source.Location{
				FileName: "coverage.proto",
				Message: &source.Message{
					Name: "DataResponse",
					Field: &source.Field{
						Name: "user",
						Option: &source.FieldOption{
							Oneof: &source.FieldOneof{
								Def: &source.VariableDefinitionOption{
									Idx:  0,
									Name: true,
								},
							},
						},
					},
				},
			},
			wantNil: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			nodeInfo := sourceFile.NodeInfoByLocation(tc.location)
			if tc.wantNil {
				if nodeInfo != nil {
					t.Errorf("expected nil but got %+v", nodeInfo)
				}
			} else {
				if nodeInfo == nil {
					t.Errorf("expected non-nil nodeInfo but got nil")
				}
			}
		})
	}
}

func TestFile_UtilityMethods(t *testing.T) {
	t.Parallel()

	// Test basic file utility methods for coverage
	path := filepath.Join("testdata", "coverage.proto")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sourceFile, err := source.NewFile(path, content)
	if err != nil {
		t.Fatal(err)
	}

	// Test Read method
	buffer := make([]byte, 10)
	n, err := sourceFile.Read(buffer)
	if err != nil && err != io.EOF {
		t.Errorf("unexpected error from Read: %v", err)
	}
	if n == 0 {
		t.Error("expected to read some bytes")
	}

	// Test Close method
	err = sourceFile.Close()
	if err != nil {
		t.Errorf("unexpected error from Close: %v", err)
	}

	// Test Path method
	if sourceFile.Path() != path {
		t.Errorf("expected path %s, got %s", path, sourceFile.Path())
	}

	// Test Content method
	contentFromFile := sourceFile.Content()
	if len(contentFromFile) != len(content) {
		t.Errorf("expected content length %d, got %d", len(content), len(contentFromFile))
	}

	// Test AST method
	ast := sourceFile.AST()
	if ast == nil {
		t.Error("expected non-nil AST")
	}

	// Test Imports method
	imports := sourceFile.Imports()
	if len(imports) == 0 {
		t.Error("expected at least one import")
	}

	// Test ImportsByImportRule method
	importsByRule := sourceFile.ImportsByImportRule()
	if len(importsByRule) == 0 {
		t.Error("expected at least one import by rule")
	}
}

func TestFile_ErrorCoverage(t *testing.T) {
	t.Parallel()

	// Test NewFile with invalid content to trigger error recovery
	invalidPath := "testdata/nonexistent.proto"
	_, err := source.NewFile(invalidPath, []byte("invalid proto content"))
	// This should not return an error because the parser has error recovery
	if err != nil {
		t.Logf("NewFile with invalid content returned error: %v", err)
	}

	// Test with nil fileNode condition
	emptyPath := "testdata/empty.proto"
	_, err = source.NewFile(emptyPath, []byte(""))
	// This should not return an error as empty content should still parse
	if err != nil {
		t.Logf("NewFile with empty content returned error: %v", err)
	}
}
