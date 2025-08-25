package server

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mercari/grpc-federation/source"
)

func TestCompletion(t *testing.T) {
	path := filepath.Join("testdata", "service.proto")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	srcFile, err := source.NewFile(path, content)
	if err != nil {
		t.Fatal(err)
	}
	file := newFile(path, srcFile, nil)
	completer := NewCompleter(logger)
	t.Run("method", func(t *testing.T) {
		// resolver.method value position of Post in service.proto file
		_, candidates, err := completer.Completion(ctx, file, []string{"testdata"}, source.Position{
			Line: 40,
			Col:  19,
		})
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(candidates, []string{
			"post.PostService/GetPost",
			"post.PostService/GetPosts",
			"user.UserService/GetUser",
			"user.UserService/GetUsers",
		}); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("request.field", func(t *testing.T) {
		// resolver.request.field value position of Post in service.proto file
		_, candidates, err := completer.Completion(ctx, file, []string{"testdata"}, source.Position{
			Line: 41,
			Col:  28,
		})
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(candidates, []string{
			"id",
		}); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("request.by", func(t *testing.T) {
		// resolver.request.by value position os Post in service.proto file
		_, candidates, err := completer.Completion(ctx, file, []string{"testdata"}, source.Position{
			Line: 41,
			Col:  38,
		})
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(candidates, []string{
			"$.id",
		}); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("filter response", func(t *testing.T) {
		// resolver.response.field value position of Post in service.proto file
		_, candidates, err := completer.Completion(ctx, file, []string{"testdata"}, source.Position{
			Line: 44,
			Col:  28,
		})
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(candidates, []string{
			"res",
			"$.id",
		}); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("message", func(t *testing.T) {
		// def[2].message value position of Post in service.proto file
		_, candidates, err := completer.Completion(ctx, file, []string{"testdata"}, source.Position{
			Line: 48,
			Col:  17,
		})
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(candidates, []string{
			// federation package messages.
			"GetPostRequest",
			"GetPostResponse",
			"User",

			// google messages
			"google.protobuf.Any",
			"google.protobuf.DescriptorProto",
			"google.protobuf.DescriptorProto.ExtensionRange",
			"google.protobuf.DescriptorProto.ReservedRange",
			"google.protobuf.Duration",
			"google.protobuf.EnumDescriptorProto",
			"google.protobuf.EnumDescriptorProto.EnumReservedRange",
			"google.protobuf.EnumOptions",
			"google.protobuf.EnumValueDescriptorProto",
			"google.protobuf.EnumValueOptions",
			"google.protobuf.ExtensionRangeOptions",
			"google.protobuf.ExtensionRangeOptions.Declaration",
			"google.protobuf.FeatureSet",
			"google.protobuf.FeatureSetDefaults",
			"google.protobuf.FeatureSetDefaults.FeatureSetEditionDefault",
			"google.protobuf.FieldDescriptorProto",
			"google.protobuf.FieldOptions",
			"google.protobuf.FieldOptions.EditionDefault",
			"google.protobuf.FieldOptions.FeatureSupport",
			"google.protobuf.FileDescriptorProto",
			"google.protobuf.FileDescriptorSet",
			"google.protobuf.FileOptions",
			"google.protobuf.GeneratedCodeInfo",
			"google.protobuf.GeneratedCodeInfo.Annotation",
			"google.protobuf.MessageOptions",
			"google.protobuf.MethodDescriptorProto",
			"google.protobuf.MethodOptions",
			"google.protobuf.OneofDescriptorProto",
			"google.protobuf.OneofOptions",
			"google.protobuf.ServiceDescriptorProto",
			"google.protobuf.ServiceOptions",
			"google.protobuf.SourceCodeInfo",
			"google.protobuf.SourceCodeInfo.Location",
			"google.protobuf.Timestamp",
			"google.protobuf.UninterpretedOption",
			"google.protobuf.UninterpretedOption.NamePart",
			"google.rpc.BadRequest",
			"google.rpc.BadRequest.FieldViolation",
			"google.rpc.DebugInfo",
			"google.rpc.ErrorInfo",
			"google.rpc.ErrorInfo.MetadataEntry",
			"google.rpc.Help",
			"google.rpc.Help.Link",
			"google.rpc.LocalizedMessage",
			"google.rpc.PreconditionFailure",
			"google.rpc.PreconditionFailure.Violation",
			"google.rpc.QuotaFailure",
			"google.rpc.QuotaFailure.Violation",
			"google.rpc.RequestInfo",
			"google.rpc.ResourceInfo",
			"google.rpc.RetryInfo",

			// post package messages.
			"post.GetPostRequest",
			"post.GetPostResponse",
			"post.GetPostsRequest",
			"post.GetPostsResponse",
			"post.Post",

			// user package messages.
			"user.GetUserRequest",
			"user.GetUserResponse",
			"user.GetUsersRequest",
			"user.GetUsersResponse",
			"user.User",
		}); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
}
