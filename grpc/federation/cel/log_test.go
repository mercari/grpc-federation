package cel_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/go-cmp/cmp"

	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
	"github.com/mercari/grpc-federation/grpc/federation/log"
)

func TestLog(t *testing.T) {
	newTestLogger := func(w *bytes.Buffer) *slog.Logger {
		return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey && len(groups) == 0 {
					return slog.Attr{}
				}
				return a
			},
		}))
	}

	tests := []struct {
		name    string
		expr    string
		ctx     func(*bytes.Buffer) context.Context
		cmpDiff func(*bytes.Buffer) string
	}{
		{
			name: "debug",
			expr: "grpc.federation.log.debug('message')",
			ctx: func(w *bytes.Buffer) context.Context {
				logger := newTestLogger(w)
				return log.WithLogger(context.Background(), logger)
			},
			cmpDiff: func(got *bytes.Buffer) string {
				expected := `{"level":"DEBUG","msg":"message"}`
				return cmp.Diff(got.String(), fmt.Sprintln(expected))
			},
		},
		{
			name: "debug with args",
			expr: "grpc.federation.log.debug('message', {'a': 100})",
			ctx: func(w *bytes.Buffer) context.Context {
				logger := newTestLogger(w)
				return log.WithLogger(context.Background(), logger)
			},
			cmpDiff: func(got *bytes.Buffer) string {
				expected := `{"level":"DEBUG","msg":"message","a":100}`
				return cmp.Diff(got.String(), fmt.Sprintln(expected))
			},
		},
		{
			name: "info",
			expr: "grpc.federation.log.info('message')",
			ctx: func(w *bytes.Buffer) context.Context {
				logger := newTestLogger(w)
				return log.WithLogger(context.Background(), logger)
			},
			cmpDiff: func(got *bytes.Buffer) string {
				expected := `{"level":"INFO","msg":"message"}`
				return cmp.Diff(got.String(), fmt.Sprintln(expected))
			},
		},
		{
			name: "info with args",
			expr: "grpc.federation.log.info('message', {'a': 100})",
			ctx: func(w *bytes.Buffer) context.Context {
				logger := newTestLogger(w)
				return log.WithLogger(context.Background(), logger)
			},
			cmpDiff: func(got *bytes.Buffer) string {
				expected := `{"level":"INFO","msg":"message","a":100}`
				return cmp.Diff(got.String(), fmt.Sprintln(expected))
			},
		},
		{
			name: "warn",
			expr: "grpc.federation.log.warn('message')",
			ctx: func(w *bytes.Buffer) context.Context {
				logger := newTestLogger(w)
				return log.WithLogger(context.Background(), logger)
			},
			cmpDiff: func(got *bytes.Buffer) string {
				expected := `{"level":"WARN","msg":"message"}`
				return cmp.Diff(got.String(), fmt.Sprintln(expected))
			},
		},
		{
			name: "warn with args",
			expr: "grpc.federation.log.warn('message', {'a': 100})",
			ctx: func(w *bytes.Buffer) context.Context {
				logger := newTestLogger(w)
				return log.WithLogger(context.Background(), logger)
			},
			cmpDiff: func(got *bytes.Buffer) string {
				expected := `{"level":"WARN","msg":"message","a":100}`
				return cmp.Diff(got.String(), fmt.Sprintln(expected))
			},
		},
		{
			name: "error",
			expr: "grpc.federation.log.error('message')",
			ctx: func(w *bytes.Buffer) context.Context {
				logger := newTestLogger(w)
				return log.WithLogger(context.Background(), logger)
			},
			cmpDiff: func(got *bytes.Buffer) string {
				expected := `{"level":"ERROR","msg":"message"}`
				return cmp.Diff(got.String(), fmt.Sprintln(expected))
			},
		},
		{
			name: "error with args",
			expr: "grpc.federation.log.error('message', {'a': 100})",
			ctx: func(w *bytes.Buffer) context.Context {
				logger := newTestLogger(w)
				return log.WithLogger(context.Background(), logger)
			},
			cmpDiff: func(got *bytes.Buffer) string {
				expected := `{"level":"ERROR","msg":"message","a":100}`
				return cmp.Diff(got.String(), fmt.Sprintln(expected))
			},
		},
		{
			name: "add",
			expr: "grpc.federation.log.add({'a': 100}) && grpc.federation.log.info('message')",
			ctx: func(w *bytes.Buffer) context.Context {
				logger := newTestLogger(w)
				return log.WithLogger(context.Background(), logger)
			},
			cmpDiff: func(got *bytes.Buffer) string {
				expected := `{"level":"INFO","msg":"message","a":100}`
				return cmp.Diff(got.String(), fmt.Sprintln(expected))
			},
		},
		{
			name: "expand value",
			expr: `grpc.federation.log.info('message', {
	'a': {
		'b': [
			google.protobuf.Timestamp{seconds: 1},
			google.protobuf.Timestamp{seconds: 2}
		]
	}
})`,
			ctx: func(w *bytes.Buffer) context.Context {
				logger := newTestLogger(w)
				return log.WithLogger(context.Background(), logger)
			},
			cmpDiff: func(got *bytes.Buffer) string {
				expected := `{"level":"INFO","msg":"message","a":{"b":["1970-01-01T00:00:01Z","1970-01-01T00:00:02Z"]}}`
				return cmp.Diff(got.String(), fmt.Sprintln(expected))
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			lib := cellib.NewLogLibrary()
			lib.Initialize(test.ctx(&output))
			env, err := cel.NewEnv(cel.Lib(lib))
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
			_, _, err = program.Eval(map[string]any{})
			if err != nil {
				t.Fatal(err)
			}
			if d := test.cmpDiff(&output); d != "" {
				t.Fatalf("(-got, +want)\n%s", d)
			}
		})
	}
}
