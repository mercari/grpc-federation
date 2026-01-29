package generator_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"

	"github.com/mercari/grpc-federation/grpc/federation/generator"
	"github.com/mercari/grpc-federation/internal/testutil"
	"github.com/mercari/grpc-federation/resolver"
)

func TestRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []string{
		"simple_aggregation",
		"minimum",
		"create_post",
		"custom_resolver",
		"async",
		"alias",
		"autobind",
		"multi_user",
		"resolver_overlaps",
		"oneof",
		"validation",
		"map",
		"condition",
		"switch",
	}
	for _, test := range tests {
		test := test
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			testdataDir := filepath.Join(testutil.RepoRoot(), "testdata")
			files := testutil.Compile(t, filepath.Join(testdataDir, fmt.Sprintf("%s.proto", test)))
			r := resolver.New(files, resolver.ImportPathOption(testdataDir))
			result, err := r.Resolve()
			if err != nil {
				t.Fatal(err)
			}
			svcToDepNum := make(map[string]int)
			for _, file := range result.Files {
				for _, svc := range file.Services {
					svcToDepNum[svc.Name] = len(svc.ServiceDependencies())
				}
			}
			genReq := generator.CreateCodeGeneratorRequest(&generator.CodeGeneratorRequestConfig{
				GRPCFederationFiles: result.Files,
			})
			genReqBytes, err := proto.Marshal(genReq)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := generator.ToCodeGeneratorRequest(bytes.NewBuffer(genReqBytes))
			if err != nil {
				t.Fatal(err)
			}
			gotSvcToDepNum := make(map[string]int)
			for _, file := range decoded.GRPCFederationFiles {
				for _, svc := range file.Services {
					gotSvcToDepNum[svc.Name] = len(svc.ServiceDependencies())
				}
			}
			if diff := cmp.Diff(svcToDepNum, gotSvcToDepNum); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
			if diff := cmp.Diff(
				decoded.GRPCFederationFiles,
				result.Files,
				testutil.ResolverCmpOpts()...,
			); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}
}
