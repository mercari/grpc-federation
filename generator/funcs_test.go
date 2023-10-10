package generator

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCreateMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc        string
		pairs       []any
		expected    map[string]any
		expectedErr string
	}{
		{
			desc:  "success",
			pairs: []any{"key1", "value1", "key2", "value2"},
			expected: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			desc:        "invalid number of arguments",
			pairs:       []any{"key1", "value1", "key2"},
			expectedErr: "the number of arguments must be divisible by two",
		},
		{
			desc:        "invalid key type",
			pairs:       []any{1, "value1"},
			expectedErr: "cannot use type int as map key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := CreateMap(tc.pairs...)
			if err != nil {
				if tc.expectedErr == "" {
					t.Fatalf("failed to call CreateMap: %v", err)
				}

				if !strings.Contains(err.Error(), tc.expectedErr) {
					t.Fatalf("received an unexpected error: %q should have contained %q", err.Error(), tc.expectedErr)
				}
				return
			}

			if tc.expectedErr != "" {
				t.Fatal("expected to receive an error but got nil")
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}
}
