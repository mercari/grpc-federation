package validator

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bufbuild/protocompile/ast"

	"github.com/mercari/grpc-federation/compiler"
	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/source"
)

type Validator struct {
	compiler      *compiler.Compiler
	importPaths   []string
	manualImport  bool
	pathToFileMap map[string]*source.File
}

func New() *Validator {
	return &Validator{
		compiler:      compiler.New(),
		pathToFileMap: map[string]*source.File{},
	}
}

type ValidationOutput struct {
	IsWarning  bool
	Path       string
	Start      ast.SourcePos
	End        ast.SourcePos
	Message    string
	SourceLine string
}

type ValidatorOption func(v *Validator)

func ImportPathOption(path ...string) ValidatorOption {
	return func(v *Validator) {
		v.importPaths = append(v.importPaths, path...)
	}
}

func ManualImportOption() ValidatorOption {
	return func(v *Validator) {
		v.manualImport = true
	}
}

func (v *Validator) Validate(ctx context.Context, file *source.File, opts ...ValidatorOption) []*ValidationOutput {
	v.pathToFileMap = map[string]*source.File{}
	copied := make([]string, len(v.importPaths))
	copy(copied, v.importPaths)
	v.importPaths = copied
	for _, opt := range opts {
		opt(v)
	}
	var compilerOpts []compiler.Option
	compilerOpts = append(compilerOpts, compiler.ImportPathOption(v.importPaths...))
	if v.manualImport {
		compilerOpts = append(compilerOpts, compiler.ManualImportOption())
	}
	protos, err := v.compiler.Compile(ctx, file, compilerOpts...)
	if err != nil {
		compilerErr, ok := err.(*compiler.CompilerError)
		if !ok {
			// unknown compile error
			return nil
		}
		return v.compilerErrorToValidationOutputs(compilerErr)
	}
	r := resolver.New(protos)
	result, err := r.Resolve()
	var outs []*ValidationOutput
	if result != nil {
		outs = v.toValidationOutputByWarnings(result.Warnings)
	}
	for _, e := range resolver.ExtractIndividualErrors(err) {
		outs = append(outs, v.toValidationOutputByError(e))
	}
	sort.SliceStable(outs, func(i, j int) bool {
		return outs[i].Start.Col < outs[j].Start.Col
	})
	sort.SliceStable(outs, func(i, j int) bool {
		return outs[i].Start.Line < outs[j].Start.Line
	})
	return outs
}

func (v *Validator) compilerErrorToValidationOutputs(err *compiler.CompilerError) []*ValidationOutput {
	if len(err.ErrWithPos) == 0 {
		return []*ValidationOutput{{Message: err.Error()}}
	}
	var outs []*ValidationOutput
	for _, e := range err.ErrWithPos {
		msg := e.Error()
		start := e.GetPosition()
		outs = append(outs, &ValidationOutput{
			Start:   start,
			End:     start,
			Message: msg,
		})
	}
	return outs
}

func (v *Validator) toValidationOutputByWarnings(warnings []*resolver.Warning) []*ValidationOutput {
	outs := make([]*ValidationOutput, 0, len(warnings))
	for _, warn := range warnings {
		out := v.toValidationOutput(warn.Location, warn.Message)
		out.IsWarning = true
		outs = append(outs, out)
	}
	return outs
}

func (v *Validator) toValidationOutputByError(err error) *ValidationOutput {
	locErr := resolver.ToLocationError(err)
	if locErr == nil {
		return &ValidationOutput{Message: err.Error()}
	}
	return v.toValidationOutput(locErr.Location, locErr.Message)
}

func (v *Validator) toValidationOutput(loc *source.Location, msg string) *ValidationOutput {
	path := loc.FileName
	file := v.getFile(path)
	if file == nil {
		return &ValidationOutput{
			Path:    path,
			Message: msg,
		}
	}
	nodeInfo := file.NodeInfoByLocation(loc)
	if nodeInfo == nil {
		return &ValidationOutput{
			Path:    path,
			Message: msg,
		}
	}
	var sourceLine string
	startLine := nodeInfo.Start().Line
	contents := strings.Split(string(file.Content()), "\n")
	if len(contents) > startLine && startLine > 0 {
		sourceLine = contents[startLine-1]
	}
	return &ValidationOutput{
		Path:       path,
		Start:      nodeInfo.Start(),
		End:        nodeInfo.End(),
		Message:    msg,
		SourceLine: sourceLine,
	}
}

func (v *Validator) getFile(path string) *source.File {
	file, exists := v.pathToFileMap[path]
	if exists {
		return file
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	f, err := source.NewFile(path, content)
	if err != nil {
		return nil
	}
	v.pathToFileMap[path] = f
	return f
}

func ExistsError(outs []*ValidationOutput) bool {
	for _, out := range outs {
		if out.IsWarning {
			continue
		}
		return true
	}
	return false
}

func Format(outs []*ValidationOutput) string {
	var b strings.Builder
	for _, out := range outs {
		var msg string
		if out.IsWarning {
			msg = fmt.Sprintf("[WARN] %s", out.Message)
		} else {
			msg = out.Message
		}
		if out.Path == "" {
			fmt.Fprintf(&b, "%s\n", msg)
			continue
		}
		startPos := out.Start
		if startPos.Line == 0 {
			fmt.Fprintf(&b, "%s: %s\n", out.Path, msg)
		} else {
			fmt.Fprintf(&b, "%s:%d:%d: %s\n", out.Path, startPos.Line, startPos.Col, msg)
		}
		if out.SourceLine != "" {
			header := fmt.Sprintf("%d: ", startPos.Line)
			fmt.Fprintf(&b, "%s %s\n", header, out.SourceLine)
			fmt.Fprintf(&b, "%s^\n", strings.Repeat(" ", len(header)+startPos.Col))
		}
	}
	return b.String()
}
