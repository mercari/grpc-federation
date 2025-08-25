package generator

import (
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/mercari/grpc-federation/grpc/federation/generator/plugin"
	"github.com/mercari/grpc-federation/resolver"
)

type ActionType string

const (
	KeepAction   ActionType = "keep"
	CreateAction ActionType = "create"
	DeleteAction ActionType = "delete"
	UpdateAction ActionType = "update"
	ProtocAction ActionType = "protoc"
)

type CodeGeneratorRequestConfig struct {
	Type                 ActionType
	ProtoPath            string
	Files                []*plugin.ProtoCodeGeneratorResponse_File
	GRPCFederationFiles  []*resolver.File
	OutputFilePathConfig resolver.OutputFilePathConfig
}

type CodeGeneratorRequest struct {
	ProtoPath            string
	OutDir               string
	Files                []*plugin.ProtoCodeGeneratorResponse_File
	GRPCFederationFiles  []*resolver.File
	OutputFilePathConfig resolver.OutputFilePathConfig
}

type (
	CodeGeneratorResponse = pluginpb.CodeGeneratorResponse
	//nolint:stylecheck
	CodeGeneratorResponse_File = pluginpb.CodeGeneratorResponse_File
)
