package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/cel-go/cel"
	exprv1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"

	grpcfed "github.com/mercari/grpc-federation/grpc/federation"

	"github.com/mercari/grpc-federation/internal/testutil"
	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/types"
)

func TestTemplateFuncsAreBoundToAFile(t *testing.T) {
	t.Parallel()

	testdataDir := filepath.Join(testutil.RepoRoot(), "testdata")
	files := testutil.Compile(t, filepath.Join(testdataDir, "simple_aggregation.proto"))
	result, err := resolver.New(files, resolver.ImportPathOption(testdataDir)).Resolve()
	if err != nil {
		t.Fatal(err)
	}
	file := result.Files[0]
	file.Name = filepath.Base(file.Name)

	// Executing the template without binding the functions must fail instead of rendering a cache index of 0.
	tmpl, err := loadTemplate()
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	err = tmpl.Execute(&b, NewFile(file))
	if err == nil || !strings.Contains(err.Error(), "is not bound to a file") {
		t.Fatalf("expected an error about the unbound function, got %v", err)
	}
}

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

// evalNativeType evaluates the Go expression that toCELNativeType renders into the CEL type it declares.
func evalNativeType(t *testing.T, src string) *cel.Type {
	t.Helper()

	expr, err := parser.ParseExpr(src)
	if err != nil {
		t.Fatalf("failed to parse %q: %v", src, err)
	}
	var eval func(ast.Expr) *cel.Type
	eval = func(e ast.Expr) *cel.Type {
		switch e := e.(type) {
		case *ast.SelectorExpr:
			if pkg, ok := e.X.(*ast.Ident); !ok || pkg.Name != "grpcfed" {
				t.Fatalf("unexpected selector in %q", src)
			}
			switch e.Sel.Name {
			case "CELDoubleType":
				return grpcfed.CELDoubleType
			case "CELIntType":
				return grpcfed.CELIntType
			case "CELUintType":
				return grpcfed.CELUintType
			case "CELBoolType":
				return grpcfed.CELBoolType
			case "CELStringType":
				return grpcfed.CELStringType
			case "CELBytesType":
				return grpcfed.CELBytesType
			case "CELNullType":
				return grpcfed.CELNullType
			}
		case *ast.CallExpr:
			sel, ok := e.Fun.(*ast.SelectorExpr)
			if !ok {
				break
			}
			switch sel.Sel.Name {
			case "CELListType", "NewCELListType":
				return cel.ListType(eval(e.Args[0]))
			case "NewCELMapType":
				return cel.MapType(eval(e.Args[0]), eval(e.Args[1]))
			case "CELObjectType", "NewCELObjectType":
				name, err := strconv.Unquote(e.Args[0].(*ast.BasicLit).Value)
				if err != nil {
					t.Fatal(err)
				}
				return cel.ObjectType(name)
			}
		}
		t.Fatalf("cannot evaluate %q", src)
		return nil
	}
	return eval(expr)
}

func TestCELNativeType(t *testing.T) {
	t.Parallel()

	message := &resolver.Message{
		Name: "Foo",
		File: &resolver.File{Package: &resolver.Package{Name: "org.federation"}},
	}
	wellKnown := func(name string) *resolver.Message {
		return &resolver.Message{
			Name: name,
			File: &resolver.File{Package: &resolver.Package{Name: "google.protobuf"}},
		}
	}
	scalar := func(kind types.Kind) *resolver.Type { return &resolver.Type{Kind: kind} }
	repeated := func(t *resolver.Type) *resolver.Type {
		t = t.Clone()
		t.Repeated = true
		return t
	}

	// The variables a precompiled expression declares must be the types the lazy path declares,
	// and both are what toCELNativeType renders. The compiler declares the same variables
	// with resolver.ToCELType when it checks the expression.
	for _, test := range []struct {
		name string
		typ  *resolver.Type
		// differsFromCompiler is set for the types the runtime declares differently from the compiler,
		// and want is then the type of the runtime.
		differsFromCompiler bool
		want                *cel.Type
	}{
		{name: "double", typ: scalar(types.Double)},
		{name: "float", typ: scalar(types.Float)},
		{name: "int32", typ: scalar(types.Int32)},
		{name: "int64", typ: scalar(types.Int64)},
		{name: "sint32", typ: scalar(types.Sint32)},
		{name: "sint64", typ: scalar(types.Sint64)},
		{name: "sfixed32", typ: scalar(types.Sfixed32)},
		{name: "sfixed64", typ: scalar(types.Sfixed64)},
		{name: "uint32", typ: scalar(types.Uint32)},
		{name: "uint64", typ: scalar(types.Uint64)},
		{name: "fixed32", typ: scalar(types.Fixed32)},
		{name: "fixed64", typ: scalar(types.Fixed64)},
		{name: "bool", typ: scalar(types.Bool)},
		{name: "string", typ: scalar(types.String)},
		{name: "bytes", typ: scalar(types.Bytes)},
		{name: "message", typ: resolver.NewMessageType(message, false)},
		{name: "repeated message", typ: resolver.NewMessageType(message, true)},
		{name: "repeated string", typ: repeated(scalar(types.String))},
		{name: "list of lists", typ: repeated(repeated(scalar(types.Int64)))},
		// No example uses these, and CEL treats most of them specially.
		{name: "duration", typ: resolver.NewMessageType(wellKnown("Duration"), false)},
		{name: "timestamp", typ: resolver.NewMessageType(wellKnown("Timestamp"), false)},
		{name: "any", typ: resolver.NewMessageType(wellKnown("Any"), false)},
		{name: "wrapper", typ: resolver.NewMessageType(wellKnown("Int64Value"), false)},
		{name: "repeated duration", typ: resolver.NewMessageType(wellKnown("Duration"), true)},
		{name: "map", typ: resolver.NewMapType(scalar(types.String), scalar(types.Int64))},
		{name: "map of messages", typ: resolver.NewMapType(scalar(types.String), resolver.NewMessageType(message, false))},
		{name: "map of lists", typ: resolver.NewMapType(scalar(types.String), repeated(scalar(types.String)))},
		{
			// The compiler declares an enum as an opaque type named after it, and the runtime as an int.
			name:                "enum",
			typ:                 &resolver.Type{Kind: types.Enum, Enum: &resolver.Enum{Name: "E", File: message.File}},
			differsFromCompiler: true,
			want:                cel.IntType,
		},
		{
			name:                "repeated enum",
			typ:                 &resolver.Type{Kind: types.Enum, Repeated: true, Enum: &resolver.Enum{Name: "E", File: message.File}},
			differsFromCompiler: true,
			want:                cel.ListType(cel.IntType),
		},
		{
			name:                "null",
			typ:                 &resolver.Type{IsNull: true},
			differsFromCompiler: false,
			want:                cel.NullType,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := evalNativeType(t, toCELNativeType(test.typ))
			want := resolver.ToCELType(test.typ)
			if test.differsFromCompiler {
				if got.IsExactType(want) {
					t.Fatalf("%s no longer differs from the compiler; update the test", test.name)
				}
				want = test.want
			}
			if !got.IsExactType(want) {
				t.Fatalf("the runtime declares %s but expected %s", got, want)
			}
		})
	}
}

func TestCELWalkerSkipsMissingNodes(t *testing.T) {
	t.Parallel()

	indexer, err := newCELIndexer(&resolver.File{})
	if err != nil {
		t.Fatal(err)
	}
	w := &celWalker{indexer: indexer, seen: make(map[int]struct{})}
	// An expression field that is not set is not an error.
	w.value(nil)
	w.name(nil)
	w.name(&resolver.VariableDefinition{})
	w.serviceVariableExpr(nil)
	w.message(&resolver.Message{Fields: []*resolver.Field{{}}})
	w.defSet(nil)
	w.def(nil)
	w.def(&resolver.VariableDefinition{Name: "no_expr"})
	w.mapExpr(nil)
	w.callExpr(nil)
	w.messageExpr(nil)
	w.enumExpr(nil)
	w.switchExpr(nil)
	w.grpcError(nil)
	if w.err != nil || len(w.entries) != 0 {
		t.Fatalf("expected no entries and no error, got %d entries and %v", len(w.entries), w.err)
	}
}

func TestCELWalkerVisitsAValueOnce(t *testing.T) {
	t.Parallel()

	indexer, err := newCELIndexer(&resolver.File{})
	if err != nil {
		t.Fatal(err)
	}
	w := &celWalker{indexer: indexer, seen: make(map[int]struct{})}
	value := &resolver.CELValue{Expr: "1", CheckedExpr: &exprv1.CheckedExpr{}}
	// A CELValue that is reachable by two paths, like the ones a service variable shares with its definition,
	// has one index and one entry.
	w.value(value)
	w.value(value)
	if len(w.entries) != 1 || indexer.values[value] != w.entries[0].Index {
		t.Fatalf("expected one entry for the value, got %d", len(w.entries))
	}
}

func TestCELWalkerSkipsIncompleteVariables(t *testing.T) {
	t.Parallel()

	indexer, err := newCELIndexer(&resolver.File{})
	if err != nil {
		t.Fatal(err)
	}
	w := &celWalker{indexer: indexer, seen: make(map[int]struct{})}
	w.value(&resolver.CELValue{
		Expr:        "ok",
		CheckedExpr: &exprv1.CheckedExpr{},
		Variables: []*resolver.CELVariable{
			nil,
			{Type: &resolver.Type{Kind: types.String}},
			{Name: "no_type"},
			{Name: "ok", Type: &resolver.Type{Kind: types.String}},
		},
	})
	if len(w.entries) != 1 || len(w.entries[0].Variables) != 1 || w.entries[0].Variables[0].Name != "ok" {
		t.Fatalf("expected one entry with the variable ok, got %+v", w.entries)
	}
}

func TestCELIndexerErrorsAreReturned(t *testing.T) {
	t.Parallel()

	// An expression the compiler did not resolve makes the indexer fail, and each function that needs it returns that.
	unresolved := &resolver.CELValue{Expr: "unresolved"}
	svc := &resolver.Service{
		Rule: &resolver.ServiceRule{
			Vars: []*resolver.ServiceVariable{{Name: "v", Expr: &resolver.ServiceVariableExpr{By: unresolved}}},
		},
	}
	file := NewFile(&resolver.File{Services: []*resolver.Service{svc}})

	if _, err := file.cacheIndex(unresolved); err == nil || !strings.Contains(err.Error(), "not resolved") {
		t.Fatalf("cacheIndex: got %v", err)
	}
	def := &resolver.VariableDefinition{Name: "v", Expr: &resolver.VariableExpr{Type: &resolver.Type{Kind: types.String}}}
	if _, err := file.nameCacheIndex(def); err == nil || !strings.Contains(err.Error(), "not resolved") {
		t.Fatalf("nameCacheIndex: got %v", err)
	}
	if _, err := file.Services()[0].CELPrecompileEntries(); err == nil || !strings.Contains(err.Error(), "not resolved") {
		t.Fatalf("CELPrecompileEntries: got %v", err)
	}
}

func TestNameCacheIndexArguments(t *testing.T) {
	t.Parallel()

	// A definition of the message block of an error detail is evaluated by its name.
	def := &resolver.VariableDefinition{
		Name: "v",
		Expr: &resolver.VariableExpr{
			Type: &resolver.Type{Kind: types.String},
			By:   &resolver.CELValue{Expr: "'v'", CheckedExpr: &exprv1.CheckedExpr{}},
		},
	}
	outer := &resolver.VariableDefinition{
		Name: "outer",
		Expr: &resolver.VariableExpr{
			Validation: &resolver.ValidationExpr{
				Error: &resolver.GRPCError{
					Details: resolver.GRPCErrorDetails{
						{Messages: &resolver.VariableDefinitionSet{Defs: resolver.VariableDefinitions{def}}},
					},
				},
			},
		},
	}
	svc := &resolver.Service{
		Messages: []*resolver.Message{{
			Name: "M",
			Rule: &resolver.MessageRule{
				DefSet: &resolver.VariableDefinitionSet{Defs: resolver.VariableDefinitions{outer}},
			},
		}},
	}
	file := NewFile(&resolver.File{Services: []*resolver.Service{svc}})

	// The template functions accept the definition both of the resolver and of the generator.
	fromResolver, err := file.nameCacheIndex(def)
	if err != nil {
		t.Fatal(err)
	}
	fromGenerator, err := file.nameCacheIndex(&VariableDefinition{VariableDefinition: def})
	if err != nil {
		t.Fatal(err)
	}
	if fromResolver != fromGenerator || fromResolver == 0 {
		t.Fatalf("got %d and %d", fromResolver, fromGenerator)
	}
	if _, err := file.nameCacheIndex("v"); err == nil || !strings.Contains(err.Error(), "unexpected variable definition type") {
		t.Fatalf("got %v", err)
	}
	// A definition without a type cannot be evaluated by its name.
	if _, err := file.nameCacheIndex(&resolver.VariableDefinition{Name: "typeless", Expr: &resolver.VariableExpr{}}); err == nil {
		t.Fatal("expected an error for a definition without a type")
	}
}
