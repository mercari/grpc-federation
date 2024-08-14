package federation

import (
	grpcfedcel "github.com/mercari/grpc-federation/grpc/federation/cel"
)

const (
	PrivatePackageName          = "grpc.federation.private"
	MessageArgumentVariableName = "__ARG__"
	ContextVariableName         = grpcfedcel.ContextVariableName
	ContextTypeName             = grpcfedcel.ContextTypeName
)
