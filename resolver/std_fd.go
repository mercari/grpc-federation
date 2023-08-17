package resolver

import (
	"github.com/bufbuild/protocompile/protoutil"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	// link in packages that include the standard protos included with protoc.
	_ "google.golang.org/protobuf/types/known/anypb"
	_ "google.golang.org/protobuf/types/known/apipb"
	_ "google.golang.org/protobuf/types/known/durationpb"
	_ "google.golang.org/protobuf/types/known/emptypb"
	_ "google.golang.org/protobuf/types/known/fieldmaskpb"
	_ "google.golang.org/protobuf/types/known/sourcecontextpb"
	_ "google.golang.org/protobuf/types/known/structpb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	_ "google.golang.org/protobuf/types/known/typepb"
	_ "google.golang.org/protobuf/types/known/wrapperspb"
	_ "google.golang.org/protobuf/types/pluginpb"
)

func stdFileDescriptors() []*descriptorpb.FileDescriptorProto {
	stdFileNames := []string{
		"google/protobuf/any.proto",
		"google/protobuf/api.proto",
		"google/protobuf/compiler/plugin.proto",
		"google/protobuf/descriptor.proto",
		"google/protobuf/duration.proto",
		"google/protobuf/empty.proto",
		"google/protobuf/field_mask.proto",
		"google/protobuf/source_context.proto",
		"google/protobuf/struct.proto",
		"google/protobuf/timestamp.proto",
		"google/protobuf/type.proto",
		"google/protobuf/wrappers.proto",
	}
	stdFDs := make([]*descriptorpb.FileDescriptorProto, 0, len(stdFileNames))
	for _, fn := range stdFileNames {
		fd, err := protoregistry.GlobalFiles.FindFileByPath(fn)
		if err != nil {
			// ignore if error occurred.
			continue
		}
		stdFDs = append(stdFDs, protoutil.ProtoFromFileDescriptor(fd))
	}
	return stdFDs
}
