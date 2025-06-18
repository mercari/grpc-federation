# How to use gRPC Federation

## Understanding gRPC Federation

First, read https://deepwiki.com/mercari/grpc-federation to understand gRPC Federation.

The `grpc/federation/federation.proto` is defined at https://github.com/mercari/grpc-federation/blob/main/proto/grpc/federation/federation.proto.

The schema defined in this proto represents the options configurable in gRPC Federation.

Detailed explanations of these options can be found at https://github.com/mercari/grpc-federation/blob/main/docs/references.md.

When describing gRPC Federation options, implement them based on this information.

## Practices for Describing gRPC Federation Options

- If you want to validate or process errors returned from dependent microservices, use the `(grpc.federation.message).def.call.error` option.
- Avoid creating messages unrelated to the response structure of the microservice you are building.
    - If the same value is always returned, use Service Variable. Otherwise, create a message only when the same process is used in multiple places and the process itself is complex.
      If the process is not complex, even if it is the same process, it is preferable to write the process directly in the existing message. Actively use the alias functionality of messages or enums when simple mapping suffices.
- To avoid redundant `grpc.federation.field` option descriptions, actively use the `autobind` feature.
- You can associate enum values with strings using `(grpc.federation.enum_value).attr`. Use this feature actively if necessary.
- gRPC Federation provides a wide range of standard libraries starting with `grpc.federation.*`. Consider whether these features can be utilized as needed.
- When the package is the same, you can omit the package prefix when referencing other messages or enums using `(grpc.federation.message).def.message` or `(grpc.federation.message).def.enum`. Otherwise, always include the package prefix.
- The CEL specification is available at https://github.com/google/cel-spec/blob/master/doc/langdef.md.
    - Additionally, in gRPC Federation, you can use the optional keyword (`?`). Always consider using optional when the existence of the target value cannot be guaranteed during field access or index access (the optional feature is also explained at https://pkg.go.dev/github.com/google/cel-go/cel#hdr-Syntax_Changes-OptionalTypes).
- Use `custom_resolver` only as a last resort. Use it only when it is impossible to express something in proto.

## Examples of gRPC Federation

To learn more about how to describe gRPC Federation options, refer to the examples at https://github.com/mercari/grpc-federation/tree/main/_examples.
Each example includes Go language files (`*_grpc_federation.pb.go`) and test codes generated automatically from the proto descriptions. Read these codes as needed to understand the meaning of the options.

## When Changing gRPC Federation Options

When changing gRPC Federation options, always compile the target proto file and repeat the process until it succeeds to verify the correctness of the description.

Follow the steps below to compile the proto file:

1. Extract the list of import files from the proto file to be compiled. Use the `get_import_proto_list` tool for extraction.
2. Create the absolute path of the import path required to compile the obtained import file list. Analyze the current repository structure to use the correct information.
3. Repeat steps 1 to 2 for the obtained import paths to create a unique and complete list of import paths required for compilation.
4. Use the `compile_proto` tool with the extracted import path list to compile the target proto file.
