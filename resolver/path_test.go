package resolver_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mercari/grpc-federation/resolver"
)

func extractFieldType(t *testing.T, src *resolver.Type, selectors ...string) *resolver.Type {
	t.Helper()
	ref := src.Ref
	var typ *resolver.Type
	for _, sel := range selectors {
		if ref == nil {
			t.Fatal("required reference to message")
		}
		field := ref.Field(sel)
		if field == nil {
			t.Fatalf("failed to get field %s from %s message type", sel, ref.Name)
		}
		typ = field.Type
		ref = field.Type.Ref
	}
	return typ
}

func TestPath(t *testing.T) {
	t.Run("selectors", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			tests := []struct {
				name     string
				path     string
				expected []string
			}{
				{"root only", "$", []string{"$"}},
				{"chain", "$.a.b.c", []string{"$", "a", "b", "c"}},
				{"index", "$.a.b[0].c", []string{"$", "a", "b", "[0]", "c"}},
				{"index all", "$.a.b[*].c", []string{"$", "a", "b", "[*]", "c"}},
				{"omit root", "a.b.c", []string{"a", "b", "c"}},
			}
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					path, err := resolver.NewPathBuilder(test.path).Build()
					if err != nil {
						t.Fatal(err)
					}
					if diff := cmp.Diff(path.Selectors(), test.expected); diff != "" {
						t.Errorf("(-got, +want)\n%s", diff)
					}
				})
			}
		})
		t.Run("error", func(t *testing.T) {
			tests := []struct {
				name string
				path string
			}{
				{"white space", "      "},
				{"invalid selector", "a."},
				{"invalid index", "a["},
			}
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					_, err := resolver.NewPathBuilder(test.path).Build()
					if err == nil {
						t.Fatal("expected error")
					}
				})
			}
		})
	})
	t.Run("type", func(t *testing.T) {
		srcType := resolver.NewMessageType(
			resolver.NewMessage("T", []*resolver.Field{
				{
					Name: "a",
					Type: resolver.NewMessageType(
						resolver.NewMessage("U", []*resolver.Field{
							{
								Name: "b",
								Type: resolver.NewMessageType(
									resolver.NewMessage("V", []*resolver.Field{
										{
											Name: "c",
											Type: resolver.StringType,
										},
									}),
									false,
								),
							},
							{
								Name: "array",
								Type: resolver.NewMessageType(
									resolver.NewMessage("V", []*resolver.Field{
										{
											Name: "c",
											Type: resolver.Int64Type,
										},
									}),
									true,
								),
							},
						}),
						false,
					),
				},
			}),
			false,
		)
		tests := []struct {
			name     string
			path     string
			expected *resolver.Type
		}{
			{"root only", "$", srcType},
			{"select message", "$.a.b", extractFieldType(t, srcType, "a", "b")},
			{"select string", "$.a.b.c", resolver.StringType},
			{"index", "$.a.array[0].c", resolver.Int64Type},
			{"index all", "$.a.array[*].c", resolver.Int64Type},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				path, err := resolver.NewPathBuilder(test.path).Build()
				if err != nil {
					t.Fatal(err)
				}
				typ, err := path.Type(srcType)
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(typ, test.expected); diff != "" {
					t.Errorf("(-got, +want)\n%s", diff)
				}
			})
		}
	})
}
