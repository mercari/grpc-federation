package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mercari/grpc-federation/internal/testutil"
	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/types"
)

func TestCELIndexerDeterministic(t *testing.T) {
	t.Parallel()

	// The resolver orders the messages of a service by their short names only, and breaks ties
	// in an unspecified order. Two nested messages named Foo must still get the same indexes on every run.
	testdataDir := filepath.Join(testutil.RepoRoot(), "testdata")
	files := testutil.Compile(t, filepath.Join(testdataDir, "nested_same_name.proto"))

	table := func() string {
		result, err := resolver.New(files, resolver.ImportPathOption(testdataDir)).Resolve()
		if err != nil {
			t.Fatal(err)
		}
		file := result.Files[0]
		file.Name = filepath.Base(file.Name)
		var b strings.Builder
		for _, svc := range NewFile(file).Services() {
			entries, err := svc.CELPrecompileEntries()
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				fmt.Fprintf(&b, "%d:%s\n", entry.Index, entry.Expr)
			}
		}
		return b.String()
	}

	want := table()
	if !strings.Contains(want, "'a'") || !strings.Contains(want, "'b'") {
		t.Fatalf("unexpected table:\n%s", want)
	}
	for i := 0; i < 40; i++ {
		if got := table(); got != want {
			t.Fatalf("the indexes changed between runs:\n%s\nversus\n%s", want, got)
		}
	}
}

func TestCELIndexerMissedExpression(t *testing.T) {
	t.Parallel()

	testdataDir := filepath.Join(testutil.RepoRoot(), "testdata")
	files := testutil.Compile(t, filepath.Join(testdataDir, "simple_aggregation.proto"))
	result, err := resolver.New(files, resolver.ImportPathOption(testdataDir)).Resolve()
	if err != nil {
		t.Fatal(err)
	}
	file := result.Files[0]
	file.Name = filepath.Base(file.Name)

	generate := func(f *File) error {
		tmpl, err := loadTemplate()
		if err != nil {
			t.Fatal(err)
		}
		_, err = generateGoContent(tmpl, f)
		return err
	}

	t.Run("every expression is reached", func(t *testing.T) {
		t.Parallel()
		if err := generate(NewFile(file)); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("an expression the walker did not reach fails generation", func(t *testing.T) {
		t.Parallel()
		f := NewFile(file)
		indexer, err := f.celIndexer()
		if err != nil {
			t.Fatal(err)
		}
		// Forget the first expression, as if the walker had not visited it.
		var missed *resolver.CELValue
		for value, idx := range indexer.values {
			if idx == 1 {
				missed = value
			}
		}
		if missed == nil {
			t.Fatal("no expression was indexed")
		}
		delete(indexer.values, missed)

		err = generate(f)
		if err == nil {
			t.Fatal("expected generation to fail")
		}
		if !strings.Contains(err.Error(), "did not visit") || !strings.Contains(err.Error(), missed.Expr) {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestCELIndexerUnresolvedExpression(t *testing.T) {
	t.Parallel()

	testdataDir := filepath.Join(testutil.RepoRoot(), "testdata")
	files := testutil.Compile(t, filepath.Join(testdataDir, "simple_aggregation.proto"))
	result, err := resolver.New(files, resolver.ImportPathOption(testdataDir)).Resolve()
	if err != nil {
		t.Fatal(err)
	}
	file := result.Files[0]
	file.Name = filepath.Base(file.Name)

	indexer, err := NewFile(file).celIndexer()
	if err != nil {
		t.Fatal(err)
	}
	// Make an expression look as if the compiler had not resolved it.
	var unresolved *resolver.CELValue
	for value, idx := range indexer.values {
		if idx == 1 {
			unresolved = value
		}
	}
	if unresolved == nil {
		t.Fatal("no expression was indexed")
	}
	unresolved.CheckedExpr = nil

	_, err = NewFile(file).celIndexer()
	if err == nil {
		t.Fatal("expected indexing to fail")
	}
	if !strings.Contains(err.Error(), "not resolved") || !strings.Contains(err.Error(), unresolved.Expr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCELIndexerNames(t *testing.T) {
	t.Parallel()

	def := func(name string, kind types.Kind) *resolver.VariableDefinition {
		return &resolver.VariableDefinition{
			Name: name,
			Expr: &resolver.VariableExpr{Type: &resolver.Type{Kind: kind}},
		}
	}
	register := func(t *testing.T, defs ...*resolver.VariableDefinition) (*celIndexer, []*CELPrecompileEntry) {
		t.Helper()
		indexer, err := newCELIndexer(&resolver.File{})
		if err != nil {
			t.Fatal(err)
		}
		w := &celWalker{indexer: indexer, seen: make(map[int]struct{})}
		for _, d := range defs {
			w.name(d)
		}
		if w.err != nil {
			t.Fatal(w.err)
		}
		return indexer, w.entries
	}
	indexes := func(t *testing.T, indexer *celIndexer, defs ...*resolver.VariableDefinition) []int {
		t.Helper()
		ret := make([]int, 0, len(defs))
		for _, d := range defs {
			idx, err := indexer.nameIndex(d)
			if err != nil {
				t.Fatal(err)
			}
			ret = append(ret, idx)
		}
		return ret
	}

	t.Run("same name and type share an index and an entry", func(t *testing.T) {
		t.Parallel()
		a, b := def("_def0_err_detail0_msg0", types.String), def("_def0_err_detail0_msg0", types.String)
		indexer, entries := register(t, a, b)
		got := indexes(t, indexer, a, b)
		if got[0] != got[1] {
			t.Fatalf("expected the same index, got %v", got)
		}
		if len(entries) != 1 {
			t.Fatalf("expected one entry, got %d", len(entries))
		}
		if entries[0].Expr != "_def0_err_detail0_msg0" || len(entries[0].Variables) != 1 ||
			entries[0].Variables[0].Name != "_def0_err_detail0_msg0" || entries[0].Variables[0].CELType != "grpcfed.CELStringType" {
			t.Fatalf("unexpected entry: %+v", entries[0])
		}
	})

	t.Run("same name and different type get different indexes and entries", func(t *testing.T) {
		t.Parallel()
		a, b := def("_def0_err_detail0_msg0", types.String), def("_def0_err_detail0_msg0", types.Bool)
		indexer, entries := register(t, a, b)
		got := indexes(t, indexer, a, b)
		if got[0] == got[1] {
			t.Fatalf("expected different indexes, got %v", got)
		}
		if len(entries) != 2 {
			t.Fatalf("expected two entries, got %d", len(entries))
		}
	})

	t.Run("different name and same type get different indexes and entries", func(t *testing.T) {
		t.Parallel()
		a, b := def("_def0_err_detail0_msg0", types.String), def("_def0_err_detail0_msg1", types.String)
		indexer, entries := register(t, a, b)
		got := indexes(t, indexer, a, b)
		if got[0] == got[1] {
			t.Fatalf("expected different indexes, got %v", got)
		}
		if len(entries) != 2 {
			t.Fatalf("expected two entries, got %d", len(entries))
		}
	})

	t.Run("types with the same declaration share an index", func(t *testing.T) {
		t.Parallel()
		// An enum is declared as an int by toCELTypeDeclare, so the programs are identical.
		a, b := def("v", types.Enum), def("v", types.Int64)
		indexer, entries := register(t, a, b)
		got := indexes(t, indexer, a, b)
		if got[0] != got[1] || len(entries) != 1 {
			t.Fatalf("expected one shared index and entry, got %v and %d entries", got, len(entries))
		}
	})

	t.Run("an unregistered variable is an error", func(t *testing.T) {
		t.Parallel()
		indexer, _ := register(t, def("a", types.String))
		if _, err := indexer.nameIndex(def("b", types.String)); err == nil || !strings.Contains(err.Error(), `"b"`) {
			t.Fatalf("expected an error naming the variable, got %v", err)
		}
		if _, err := indexer.nameIndex(def("a", types.Bool)); err == nil {
			t.Fatal("expected an error for a registered name with a different type")
		}
	})
}
