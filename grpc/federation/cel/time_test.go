package cel_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
)

func TestTime(t *testing.T) {
	tests := []struct {
		name string
		expr string
		args map[string]any
		cmp  func(any) error
	}{
		{
			name: "location utc",
			expr: "grpc.federation.time.UTC",
			cmp: func(got any) error {
				gotV, ok := got.(*cellib.Location)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.UTC.String()
				if diff := cmp.Diff(gotV.Location.String(), expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "loadLocation",
			expr: "grpc.federation.time.loadLocation('America/Los_Angeles')",
			cmp: func(got any) error {
				gotV, ok := got.(*cellib.Location)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected, err := time.LoadLocation("America/Los_Angeles")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotV.Location.String(), expected.String(), cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "location_string",
			expr: "grpc.federation.time.loadLocation('America/Los_Angeles').string()",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected, err := time.LoadLocation("America/Los_Angeles")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(string(gotV), expected.String(), cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "date",
			expr: "grpc.federation.time.date(2009, 11, 10, 23, 0, 0, 0, grpc.federation.time.UTC)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "now",
			expr: "grpc.federation.time.now()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now()
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse",
			expr: "grpc.federation.time.parse(grpc.federation.time.RFC3339, '2006-01-02T15:04:05Z')",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "fixedZone",
			expr: "grpc.federation.time.fixedZone('UTC-8', -8*60*60)",
			cmp: func(got any) error {
				gotV, ok := got.(*cellib.Location)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.FixedZone("UTC-8", -8*60*60).String()
				if diff := cmp.Diff(gotV.Location.String(), expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "addDate",
			expr: "grpc.federation.time.now().addDate(1, 1, 1)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().AddDate(1, 1, 1)
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "utc",
			expr: "grpc.federation.time.now().utc()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().UTC()
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "year",
			expr: "grpc.federation.time.now().year()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().Year()
				if diff := cmp.Diff(int(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "month",
			expr: "grpc.federation.time.now().month()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().Month()
				if diff := cmp.Diff(time.Month(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(cel.Lib(new(cellib.TimeLibrary)))
			if err != nil {
				t.Fatal(err)
			}
			ast, iss := env.Compile(test.expr)
			if iss.Err() != nil {
				t.Fatal(iss.Err())
			}
			program, err := env.Program(ast)
			if err != nil {
				t.Fatal(err)
			}
			out, _, err := program.Eval(test.args)
			if err != nil {
				t.Fatal(err)
			}
			if err := test.cmp(out); err != nil {
				t.Fatal(err)
			}
		})
	}
}
