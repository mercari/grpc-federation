package generator

import (
	"testing"

	"github.com/mercari/grpc-federation/resolver"
)

func TestServiceCELCacheIndex(t *testing.T) {
	t.Parallel()

	newTestService := func() *Service {
		return newService(&resolver.Service{Name: "TestService"}, nil)
	}

	t.Run("records every issued index", func(t *testing.T) {
		t.Parallel()
		svc := newTestService()
		for i, expr := range []string{`1`, `2`} {
			idx, err := svc.CELCacheIndex(&resolver.CELValue{Expr: expr})
			if err != nil {
				t.Fatal(err)
			}
			if idx != i+1 {
				t.Fatalf("unexpected index %d", idx)
			}
		}
		entries := svc.CELPrecompileEntries()
		if len(entries) != 2 || entries[0].Index != 1 || entries[1].Index != 2 {
			t.Fatalf("unexpected entries: %+v", entries)
		}
	})

	t.Run("nil CEL value is an error", func(t *testing.T) {
		t.Parallel()
		if _, err := newTestService().CELCacheIndex(nil); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("index issued after the table was rendered is an error", func(t *testing.T) {
		t.Parallel()
		svc := newTestService()
		svc.CELPrecompileEntries()
		if _, err := svc.CELCacheIndex(&resolver.CELValue{Expr: `1`}); err == nil {
			t.Fatal("expected error")
		}
	})
}
