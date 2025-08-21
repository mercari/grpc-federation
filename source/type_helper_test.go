package source_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mercari/grpc-federation/source"
)

func TestLocation_IsDefinedTypeName(t *testing.T) {
	t.Parallel()

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
		want     bool
	}{
		{
			desc: "IsDefinedTypeName returns true for MessageExprOption with Name=true",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 3,
							Message: &source.MessageExprOption{
								Name: true,
							},
						},
					},
				},
			},
			want: true,
		},
		{
			desc: "IsDefinedTypeName returns true for EnumExprOption with Name=true",
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
			want: true,
		},
		{
			desc: "IsDefinedTypeName returns false for MessageExprOption with Name=false",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 3,
							Message: &source.MessageExprOption{
								Name: false,
							},
						},
					},
				},
			},
			want: false,
		},
		{
			desc: "IsDefinedTypeName returns false for EnumExprOption with Name=false",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 1,
							Map: &source.MapExprOption{
								Enum: &source.EnumExprOption{
									Name: false,
								},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			desc: "IsDefinedTypeName returns false for EnumExprOption with By=true but Name=false",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Option: &source.MessageOption{
						Def: &source.VariableDefinitionOption{
							Idx: 1,
							Map: &source.MapExprOption{
								Enum: &source.EnumExprOption{
									Name: false,
									By:   true,
								},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			desc: "IsDefinedTypeName returns false for unrelated location",
			location: &source.Location{
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
			want: false,
		},
		{
			desc: "IsDefinedTypeName returns false for FieldOneof with Name=false",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "NestedMessage",
					Oneof: &source.Oneof{
						Name: "choice",
						Field: &source.Field{
							Name: "value",
							Option: &source.FieldOption{
								Oneof: &source.FieldOneof{
									Def: &source.VariableDefinitionOption{
										Idx: 0,
										Message: &source.MessageExprOption{
											Name: false,
										},
									},
								},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			desc: "IsDefinedTypeName returns true for FieldOneof with MessageExprOption Name=true",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "NestedMessage",
					Oneof: &source.Oneof{
						Name: "choice",
						Field: &source.Field{
							Name: "value",
							Option: &source.FieldOption{
								Oneof: &source.FieldOneof{
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
				},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			nodeInfo := sourceFile.NodeInfoByLocation(tc.location)
			if nodeInfo == nil {
				t.Fatalf("nodeInfo is nil for location: %+v", tc.location)
			}

			got := tc.location.IsDefinedTypeName()
			if got != tc.want {
				t.Errorf("IsDefinedTypeName() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLocation_IsDefinedFieldType(t *testing.T) {
	t.Parallel()

	path := filepath.Join("testdata", "coverage.proto")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = source.NewFile(path, content)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		desc     string
		location *source.Location
		want     bool
	}{
		{
			desc: "IsDefinedFieldType returns true for Field with Type=true",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Field: &source.Field{
						Name: "id",
						Type: true,
					},
				},
			},
			want: true,
		},
		{
			desc: "IsDefinedFieldType returns false for Field with Type=false",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "User",
					Field: &source.Field{
						Name: "id",
						Type: false,
					},
				},
			},
			want: false,
		},
		{
			desc: "IsDefinedFieldType returns true for Oneof Field with Type=true",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "NestedMessage",
					Oneof: &source.Oneof{
						Name: "choice",
						Field: &source.Field{
							Name: "value",
							Type: true,
						},
					},
				},
			},
			want: true,
		},
		{
			desc: "IsDefinedFieldType returns false for Oneof Field with Type=false",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "NestedMessage",
					Oneof: &source.Oneof{
						Name: "choice",
						Field: &source.Field{
							Name: "value",
							Type: false,
						},
					},
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.location.IsDefinedFieldType()
			if got != tc.want {
				t.Errorf("IsDefinedFieldType() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLocation_IsDefinedTypeAlias(t *testing.T) {
	t.Parallel()

	path := filepath.Join("testdata", "coverage.proto")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = source.NewFile(path, content)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		desc     string
		location *source.Location
		want     bool
	}{
		{
			desc: "IsDefinedTypeAlias returns true for MessageOption with Alias=true",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "ComplexResponse",
					Option: &source.MessageOption{
						Alias: true,
					},
				},
			},
			want: true,
		},
		{
			desc: "IsDefinedTypeAlias returns false for MessageOption with Alias=false",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "ComplexResponse",
					Option: &source.MessageOption{
						Alias: false,
					},
				},
			},
			want: false,
		},
		{
			desc: "IsDefinedTypeAlias returns true for FieldOption with Alias=true",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "DataItem",
					Field: &source.Field{
						Name: "id",
						Option: &source.FieldOption{
							Alias: true,
						},
					},
				},
			},
			want: true,
		},
		{
			desc: "IsDefinedTypeAlias returns false for FieldOption with Alias=false",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Message: &source.Message{
					Name: "DataItem",
					Field: &source.Field{
						Name: "id",
						Option: &source.FieldOption{
							Alias: false,
						},
					},
				},
			},
			want: false,
		},
		{
			desc: "IsDefinedTypeAlias returns true for EnumOption with Alias=true",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Enum: &source.Enum{
					Name: "Status",
					Option: &source.EnumOption{
						Alias: true,
					},
				},
			},
			want: true,
		},
		{
			desc: "IsDefinedTypeAlias returns false for EnumOption with Alias=false",
			location: &source.Location{
				FileName: "testdata/coverage.proto",
				Enum: &source.Enum{
					Name: "Status",
					Option: &source.EnumOption{
						Alias: false,
					},
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.location.IsDefinedTypeAlias()
			if got != tc.want {
				t.Errorf("IsDefinedTypeAlias() = %v, want %v", got, tc.want)
			}
		})
	}
}
