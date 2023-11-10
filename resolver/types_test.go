package resolver_test

import (
	"testing"

	exprv1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"

	"github.com/mercari/grpc-federation/resolver"
)

func TestValidationError_ReferenceNames(t *testing.T) {
	v := &resolver.ValidationError{
		Rule: &resolver.CELValue{
			CheckedExpr: &exprv1.CheckedExpr{
				ReferenceMap: map[int64]*exprv1.Reference{
					0: {Name: "name1"},
				},
			},
		},
		Details: []*resolver.ValidationErrorDetail{
			{
				Rule: &resolver.CELValue{
					CheckedExpr: &exprv1.CheckedExpr{
						ReferenceMap: map[int64]*exprv1.Reference{
							0: {Name: "name2"},
						},
					},
				},
				PreconditionFailures: []*resolver.PreconditionFailure{
					{
						Violations: []*resolver.PreconditionFailureViolation{
							{
								Type: &resolver.CELValue{
									CheckedExpr: &exprv1.CheckedExpr{
										ReferenceMap: map[int64]*exprv1.Reference{
											0: {Name: "name3"},
										},
									},
								},
								Subject: &resolver.CELValue{
									CheckedExpr: &exprv1.CheckedExpr{
										ReferenceMap: map[int64]*exprv1.Reference{
											0: {Name: "name4"},
										},
									},
								},
								Description: &resolver.CELValue{
									CheckedExpr: &exprv1.CheckedExpr{
										ReferenceMap: map[int64]*exprv1.Reference{
											0: {Name: "name5"},
										},
									},
								},
							},
							{
								Type: &resolver.CELValue{
									CheckedExpr: &exprv1.CheckedExpr{
										ReferenceMap: map[int64]*exprv1.Reference{
											0: {Name: "name3"},
										},
									},
								},
								Subject: &resolver.CELValue{
									CheckedExpr: &exprv1.CheckedExpr{
										ReferenceMap: map[int64]*exprv1.Reference{
											0: {Name: "name4"},
										},
									},
								},
								Description: &resolver.CELValue{
									CheckedExpr: &exprv1.CheckedExpr{
										ReferenceMap: map[int64]*exprv1.Reference{
											0: {Name: "name6"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	expected := []string{
		"name1",
		"name2",
		"name3",
		"name4",
		"name5",
		"name6",
	}
	got := v.ReferenceNames()
	// the map order is not guaranteed in Go
	if len(expected) != len(got) {
		t.Errorf("the number of reference names is different: expected: %d, got: %d", len(expected), len(got))
	}
	for _, e := range expected {
		var found bool
		for _, g := range got {
			if g == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%q does not exist in the received reference names: %v", e, got)
		}
	}
}
