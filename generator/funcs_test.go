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

func TestParentCtx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc        string
		level       int
		expected    string
		expectedErr string
	}{
		{
			desc:     "level is more than or equal to 2",
			level:    2,
			expected: "ctx1",
		},
		{
			desc:     "level is 1",
			level:    1,
			expected: "ctx",
		},
		{
			desc:        "level is less than or equal to 0",
			level:       0,
			expectedErr: "level cannot be less than or equal to 0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := ParentCtx(tc.level)
			if err != nil {
				if tc.expectedErr == "" {
					t.Fatalf("failed to call ParentCtx: %v", err)
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

func TestAdd(t *testing.T) {
	got := Add(1, 2)
	if got != 3 {
		t.Fatalf("received unexpecd result: got: %d, expected: %d", got, 3)
	}
}
