package validator_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mercari/grpc-federation/source"
	"github.com/mercari/grpc-federation/validator"
)

func TestValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		file     string
		expected string
	}{
		{file: "conflict_service_variable.proto", expected: `
conflict_service_variable.proto:44:1: service variable "bar" has different types across services: BarService, FooService
44:  message GetResponse {
     ^
conflict_service_variable.proto:56:11: ERROR: <input>:1:20: undefined field 'baz'
 | grpc.federation.var.baz
 | ...................^
56:        by: "grpc.federation.var.baz"
               ^
`},
		{file: "empty_response_field.proto", expected: `
`},
		{file: "different_message_argument_type.proto", expected: `
different_message_argument_type.proto:28:14: "id" argument name is declared with a different type kind. found "string" and "int64" type
28:          args { name: "id" by: "1" }
                  ^
`},
		{file: "duplicated_variable_name.proto", expected: `
duplicated_variable_name.proto:27:25: found duplicated variable name "a"
27:              def { name: "a" by: "1" }
                             ^
duplicated_variable_name.proto:29:27: found duplicated variable name "a"
29:                def { name: "a" by: "2" }
                               ^
duplicated_variable_name.proto:38:25: found duplicated variable name "a"
38:              def { name: "a" by: "3" }
                             ^
duplicated_variable_name.proto:40:27: found duplicated variable name "a"
40:                def { name: "a" by: "4" }
                               ^
duplicated_variable_name.proto:50:19: found duplicated variable name "a"
50:        def { name: "a" by: "5" }
                       ^
`},
		{file: "invalid_autobind.proto", expected: `
invalid_autobind.proto:23:3: "id" field found multiple times in the message specified by autobind. since it is not possible to determine one, please use "grpc.federation.field" to explicitly bind it. found message names are "a" name at def and "b" name at def
23:    string id = 1;
       ^
invalid_autobind.proto:23:3: "id" field in "org.federation.GetResponse" message needs to specify "grpc.federation.field" option
23:    string id = 1;
       ^
`},
		{file: "invalid_call_error_handler.proto", expected: `
invalid_call_error_handler.proto:44:21: cannot set both "ignore" and "ignore_and_response"
44:              ignore: true
                         ^
invalid_call_error_handler.proto:45:34: cannot set both "ignore" and "ignore_and_response"
45:              ignore_and_response: "post.GetPostResponse{}"
                                      ^
invalid_call_error_handler.proto:49:19: "by" must always return a message value
49:                by: "1"
                       ^
invalid_call_error_handler.proto:53:34: value must be "post.GetPostResponse" type
53:              ignore_and_response: "10"
                                      ^
`},
		{file: "invalid_call_metadata.proto", expected: `
invalid_call_metadata.proto:26:19: gRPC Call metadata's value type must be map<string, repeated string> type
26:          metadata: "{'foo': 'bar'}"
                       ^
`},
		{file: "invalid_call_option.proto", expected: `
invalid_call_option.proto:29:19: "hdr" variable is not defined
29:            header: "hdr"
                       ^
invalid_call_option.proto:30:20: gRPC Call option trailer's value type must be map<string, repeated string> type
30:            trailer: "tlr"
                        ^
`},
		{file: "invalid_condition_type.proto", expected: `
invalid_condition_type.proto:38:13: return value of "if" must be bool type but got string type
38:          if: "$.id"
                 ^
`},
		{file: "invalid_field_option.proto", expected: `
invalid_field_option.proto:31:50: ERROR: <input>:1:5: undefined field 'invalid'
 | post.invalid
 | ....^
31:    Post post = 1 [(grpc.federation.field) = { by: "post.invalid" }];
                                                      ^
`},
		{file: "invalid_field_type.proto", expected: `
invalid_field_type.proto:18:3: cannot convert type automatically: field type is "string" but specified value type is "int64"
18:    string a = 1 [(grpc.federation.field).by = "1"];
       ^
`},
		{file: "invalid_go_package.proto", expected: `
invalid_go_package.proto:9:21: go_package option "a;b;c;d" is invalid
9:  option go_package = "a;b;c;d";
                        ^
`},
		{file: "invalid_enum_alias_target.proto", expected: `
invalid_enum_alias_target.proto:49:1: required specify alias = "org.post.PostDataType" in grpc.federation.enum option for the "org.federation.PostType" type to automatically assign a value to the "PostData.type" field via autobind
49:  enum PostType {
     ^
invalid_enum_alias_target.proto:68:3: required specify alias = "org.post.PostContent.Category" in grpc.federation.enum option for the "org.federation.PostContent.Category" type to automatically assign a value to the "PostContent.category" field via autobind
68:    enum Category {
       ^
`},
		{file: "invalid_enum_attribute.proto", expected: `
invalid_enum_attribute.proto:25:13: attribute name is required
25:        name: ""
                 ^
invalid_enum_attribute.proto:30:13: "xxx" attribute must be defined for all enum values, but it is only defined for 1/3 of them
30:        name: "xxx"
                 ^
invalid_enum_attribute.proto:36:13: "yyy" attribute must be defined for all enum values, but it is only defined for 1/3 of them
36:        name: "yyy"
                 ^
`},
		{file: "invalid_enum_conversion.proto", expected: `
invalid_enum_conversion.proto:27:13: required specify alias = "org.federation.PostType" in grpc.federation.enum option for the "org.post.PostContent.Category" type to automatically assign a value
27:          by: "org.post.PostContent.Category.value('CATEGORY_B')"
                 ^
invalid_enum_conversion.proto:33:13: enum must always return a enum value, but got "int64" type
33:          by: "1"
                 ^
invalid_enum_conversion.proto:49:15: required specify alias = "org.federation.PostType" in grpc.federation.enum option for the "org.post.PostContent.Category" type to automatically assign a value
49:            by: "typ"
                   ^
`},
		{file: "invalid_enum_selector.proto", expected: `
invalid_enum_selector.proto:22:15: ERROR: <input>:1:56: cannot specify an int type. if you are directly specifying an enum value, you need to explicitly use "pkg.EnumName.value('ENUM_VALUE')" function to use the enum type
 | grpc.federation.enum.select(true, org.post.PostDataType.POST_TYPE_B, 'foo')
 | .......................................................^
22:      def { by: "grpc.federation.enum.select(true, org.post.PostDataType.POST_TYPE_B, 'foo')" }
                   ^
invalid_enum_selector.proto:28:1: required specify alias = "org.post.PostContent.Category" in grpc.federation.enum option for the "org.federation.PostDataType" type to automatically assign a value to the "GetPostResponse.type" field via autobind
28:  enum PostDataType {
     ^
`},
		{file: "invalid_enum_value_noalias.proto", expected: `
invalid_enum_value_noalias.proto:52:77: "noalias" cannot be specified simultaneously with "default" or "alias"
52:    POST_TYPE_A = 0 [(grpc.federation.enum_value) = { default: true, noalias: true }];
                                                                                 ^
invalid_enum_value_noalias.proto:53:62: "noalias" cannot be specified simultaneously with "default" or "alias"
53:    POST_TYPE_B = 1 [(grpc.federation.enum_value) = { noalias: true, alias: ["POST_TYPE_B"] }];
                                                                  ^
`},
		{file: "invalid_env.proto", expected: `
invalid_env.proto:11:9: "message" and "var" cannot be used simultaneously
11:      env {
             ^
invalid_env.proto:24:16: "org.federation.Invalid" message does not exist
24:        message: "Invalid"
                    ^
`},
		{file: "invalid_multiple_env.proto", expected: `
invalid_multiple_env.proto:56:1: environment variable "ccc" has different types across services: InlineEnvService, RefEnvService
56:  message GetNameResponse {
     ^
invalid_multiple_env.proto:56:1: environment variable "eee" has different types across services: InlineEnvService, RefEnvService
56:  message GetNameResponse {
     ^
`},
		{file: "invalid_error_variable.proto", expected: `
invalid_error_variable.proto:20:17: "error" is the reserved keyword. this name is not available
20:      def { name: "error" by: "'foo'" }
                     ^
invalid_error_variable.proto:21:25: ERROR: <input>:1:1: undeclared reference to 'error' (in container 'org.federation')
 | error
 | ^
21:      def { name: "e" by: "error" }
                             ^
invalid_error_variable.proto:25:15: ERROR: <input>:1:1: undeclared reference to 'error' (in container 'org.federation')
 | error.code == 0
 | ^
25:            if: "error.code == 0"
                   ^
`},
		{file: "invalid_map_iterator_src_type.proto", expected: `
invalid_map_iterator_src_type.proto:43:18: map iterator's src value type must be repeated type
43:              src: "post_ids"
                      ^
invalid_map_iterator_src_type.proto:54:57: ERROR: <input>:1:1: undeclared reference to 'users' (in container 'org.federation')
 | users
 | ^
54:    repeated User users = 4 [(grpc.federation.field).by = "users"];
                                                             ^
invalid_map_iterator_src_type.proto:58:47: ERROR: <input>:1:8: undefined field 'user_id'
 | __ARG__.user_id
 | .......^
58:    string id = 1 [(grpc.federation.field).by = "$.user_id"];
                                                   ^
`},
		{file: "invalid_map_iterator_src.proto", expected: `
invalid_map_iterator_src.proto:39:18: "posts" variable is not defined
39:              src: "posts"
                      ^
invalid_map_iterator_src.proto:43:41: ERROR: <input>:1:1: undeclared reference to 'iter' (in container 'org.federation')
 | iter.id
 | ^
43:              args { name: "user_id", by: "iter.id" }
                                             ^
invalid_map_iterator_src.proto:54:47: ERROR: <input>:1:8: undefined field 'user_id'
 | __ARG__.user_id
 | .......^
54:    string id = 1 [(grpc.federation.field).by = "$.user_id"];
                                                   ^
`},
		{file: "invalid_message_alias_target.proto", expected: `
invalid_message_alias_target.proto:46:3: required specify alias = "org.post.PostData" in grpc.federation.message option for the "org.federation.PostData" type to automatically assign a value to the "Post.data" field via autobind
46:    PostData data = 4;
       ^
`},
		{file: "invalid_message_alias.proto", expected: `
invalid_message_alias.proto:46:3: required specify alias = "org.post.PostData" in grpc.federation.message option for the "org.federation.PostData" type to automatically assign a value to the "Post.data" field via autobind
46:    PostData data = 4;
       ^
invalid_message_alias.proto:56:44: cannot find package from "invalid.Invalid"
56:    option (grpc.federation.message).alias = "invalid.Invalid";
                                                ^
invalid_message_alias.proto:58:3: "type" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
58:    PostType type = 1;
       ^
invalid_message_alias.proto:59:3: "title" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
59:    string title = 2;
       ^
invalid_message_alias.proto:60:3: "content" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
60:    PostContent content = 3;
       ^
invalid_message_alias.proto:75:3: "org.federation.SomeUser" message does not exist
75:    option (grpc.federation.message).alias = "SomeUser";
       ^
invalid_message_alias.proto:77:3: "name" field in "org.federation.User" message needs to specify "grpc.federation.field" option
77:    string name = 1;
       ^
invalid_message_alias.proto:81:3: "google.protobuf.Comment" message does not exist
81:    option (grpc.federation.message).alias = "google.protobuf.Comment";
       ^
invalid_message_alias.proto:83:3: "body" field in "org.federation.Comment" message needs to specify "grpc.federation.field" option
83:    string body = 1;
       ^
`},
		{file: "invalid_nested_message_field.proto", expected: `
invalid_nested_message_field.proto:52:7: "body" field in "federation.A.B.C" message needs to specify "grpc.federation.field" option
52:        string body = 1;
           ^
`},
		{file: "invalid_method.proto", expected: `
invalid_method.proto:37:24: invalid method format. required format is "<package-name>.<service-name>/<method-name>" but specified ""
37:        { call { method: "" } },
                            ^
invalid_method.proto:42:26: ERROR: <input>:1:1: undeclared reference to 'invalid' (in container 'federation')
 | invalid
 | ^
42:            args { inline: "invalid" }
                              ^
invalid_method.proto:47:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
47:    string id = 1;
       ^
invalid_method.proto:48:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
48:    string title = 2;
       ^
invalid_method.proto:49:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
49:    string content = 3;
       ^
invalid_method.proto:59:36: ERROR: <input>:1:8: undefined field 'user_id'
 | __ARG__.user_id
 | .......^
59:          request { field: "id", by: "$.user_id" }
                                        ^
`},
		{file: "invalid_multi_alias.proto", expected: `
invalid_multi_alias.proto:56:3: if multiple aliases are specified, you must use grpc.federation.enum.select function to bind
56:    PostType post_type = 4 [(grpc.federation.field).by = "org.post.PostDataType.POST_TYPE_A"];
       ^
invalid_multi_alias.proto:65:3: "POST_TYPE_A" value must be present in all enums, but it is missing in "org.post.PostDataType", "org.post.v2.PostDataType" enum
65:    POST_TYPE_FOO = 1 [(grpc.federation.enum_value) = { alias: ["POST_TYPE_A"] }];
       ^
invalid_multi_alias.proto:66:3: "org.post.v2.PostDataType.POST_TYPE_B" value does not exist in "org.post.PostDataType", "org.post.v2.PostDataType" enum
66:    POST_TYPE_BAR = 2 [(grpc.federation.enum_value) = { alias: ["org.post.v2.PostDataType.POST_TYPE_B", "POST_TYPE_C"] }];
       ^
invalid_multi_alias.proto:66:3: "POST_TYPE_C" value must be present in all enums, but it is missing in "org.post.PostDataType", "org.post.v2.PostDataType" enum
66:    POST_TYPE_BAR = 2 [(grpc.federation.enum_value) = { alias: ["org.post.v2.PostDataType.POST_TYPE_B", "POST_TYPE_C"] }];
       ^
invalid_multi_alias.proto:75:3: The types of "org.federation.PostData"'s "title" field ("string") and "org.post.v2.PostData"'s field ("int64") are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself
75:    string title = 2;
       ^
invalid_multi_alias.proto:75:3: "title" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
75:    string title = 2;
       ^
invalid_multi_alias.proto:76:3: required specify alias = "org.post.v2.PostContent" in grpc.federation.message option for the "org.federation.PostContent" type to automatically assign a value to the "PostData.content" field via autobind
76:    PostContent content = 3;
       ^
invalid_multi_alias.proto:77:3: specified "alias" in grpc.federation.message option, but "dummy" field does not exist in "org.post.PostData" message
77:    int64 dummy = 4;
       ^
invalid_multi_alias.proto:77:3: "dummy" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
77:    int64 dummy = 4;
       ^
`},
		{file: "invalid_oneof_selection.proto", expected: `
invalid_oneof_selection.proto:26:47: "org.federation.UserSelection" type has "user" as oneof name, but "user" has a difference type and cannot be accessed directly, so "user" becomes an undefined field
ERROR: <input>:1:4: undefined field 'user'
 | sel.user
 | ...^
26:    User user = 1 [(grpc.federation.field).by = "sel.user"];
                                                   ^
`},
		{file: "invalid_oneof.proto", expected: `
invalid_oneof.proto:43:13: return value of "if" must be bool type but got int64 type
43:          if: "1"
                 ^
invalid_oneof.proto:56:39: "if" or "default" must be specified in "grpc.federation.field.oneof"
56:        (grpc.federation.field).oneof = {
                                           ^
invalid_oneof.proto:69:39: "by" must be specified in "grpc.federation.field.oneof"
69:        (grpc.federation.field).oneof = {
                                           ^
invalid_oneof.proto:83:18: "default" found multiple times in the "grpc.federation.field.oneof". "default" can only be specified once per oneof
83:          default: true
                      ^
invalid_oneof.proto:95:3: "oneof" feature can only be used for fields within oneof
95:    bool foo = 5 [(grpc.federation.field).oneof = {
       ^
invalid_oneof.proto:95:3: value must be specified
95:    bool foo = 5 [(grpc.federation.field).oneof = {
       ^
invalid_oneof.proto:95:3: "foo" field in "org.federation.UserSelection" message needs to specify "grpc.federation.field" option
95:    bool foo = 5 [(grpc.federation.field).oneof = {
       ^
invalid_oneof.proto:112:20: "foo" field is a oneof field, so you need to specify an "if" expression
112:            { field: "foo" by: "1" },
                         ^
invalid_oneof.proto:113:20: "bar" field is a oneof field, so you need to specify an "if" expression
113:            { field: "bar" by: "'hello'" }
                         ^
`},
		{file: "invalid_retry.proto", expected: `
invalid_retry.proto:38:15: ERROR: <input>:1:1: undeclared reference to 'foo' (in container 'org.federation')
 | foo
 | ^
38:            if: "foo"
                   ^
invalid_retry.proto:40:23: time: missing unit in duration "1"
40:              interval: "1"
                           ^
invalid_retry.proto:55:15: if must always return a boolean value
55:            if: "1"
                   ^
invalid_retry.proto:57:31: time: missing unit in duration "2"
57:              initial_interval: "2"
                                   ^
invalid_retry.proto:73:27: time: missing unit in duration "3"
73:              max_interval: "3"
                               ^
`},
		{file: "invalid_method_service_name.proto", expected: `
invalid_method_service_name.proto: "post.InvalidService" service does not exist
invalid_method_service_name.proto:37:24: cannot find "method" method because the service to which the method belongs does not exist
37:        { call { method: "post.InvalidService/method" } },
                            ^
invalid_method_service_name.proto:47:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
47:    string id = 1;
       ^
invalid_method_service_name.proto:48:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
48:    string title = 2;
       ^
invalid_method_service_name.proto:49:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
49:    string content = 3;
       ^
`},
		{file: "invalid_method_name.proto", expected: `
invalid_method_name.proto:37:24: "invalid" method does not exist in PostService service
37:        { call { method: "post.PostService/invalid" } },
                            ^
invalid_method_name.proto:47:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
47:    string id = 1;
       ^
invalid_method_name.proto:48:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
48:    string title = 2;
       ^
invalid_method_name.proto:49:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
49:    string content = 3;
       ^
`},
		{file: "invalid_method_timeout_format.proto", expected: `
invalid_method_timeout_format.proto:12:47: time: unknown unit "p" in duration "1p"
12:      option (grpc.federation.method).timeout = "1p";
                                                   ^
`},

		{file: "invalid_method_request.proto", expected: `
invalid_method_request.proto:43:28: "invalid" field does not exist in "post.GetPostRequest" message for method request
43:            request { field: "invalid", by: "$.invalid" }
                                ^
invalid_method_request.proto:43:43: ERROR: <input>:1:8: undefined field 'invalid'
 | __ARG__.invalid
 | .......^
43:            request { field: "invalid", by: "$.invalid" }
                                               ^
`},
		{file: "missing_field_option.proto", expected: `
missing_field_option.proto:57:3: "user" field in "federation.Post" message needs to specify "grpc.federation.field" option
57:    User user = 4;
       ^
`},
		{file: "missing_file_import.proto", expected: `
missing_file_import.proto:13:5: unknown.proto: no such file or directory
13:      "unknown.proto"
         ^
`},
		{file: "missing_map_iterator.proto", expected: `
missing_map_iterator.proto:36:13: map iterator name must be specified
36:          map {
                 ^
missing_map_iterator.proto:36:13: map iterator src must be specified
36:          map {
                 ^
`},
		{file: "missing_message_alias.proto", expected: `
missing_message_alias.proto:46:3: required specify alias = "org.post.PostData" in grpc.federation.message option for the "org.federation.PostData" type to automatically assign a value to the "Post.data" field via autobind
46:    PostData data = 4;
       ^
missing_message_alias.proto:56:3: "type" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
56:    PostType type = 1;
       ^
missing_message_alias.proto:57:3: "title" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
57:    string title = 2;
       ^
missing_message_alias.proto:58:3: use "alias" in "grpc.federation.field" option, but "alias" is not defined in "grpc.federation.message" option
58:    PostContent content = 3 [(grpc.federation.field).alias = "content"];
       ^
`},
		{file: "missing_enum_alias.proto", expected: `
missing_enum_alias.proto:49:1: required specify alias = "org.post.PostDataType" in grpc.federation.enum option for the "org.federation.PostType" type to automatically assign a value to the "PostData.type" field via autobind
49:  enum PostType {
     ^
missing_enum_alias.proto:51:62: use "alias" in "grpc.federation.enum_value" option, but "alias" is not defined in "grpc.federation.enum" option
51:    POST_TYPE_FOO = 1 [(grpc.federation.enum_value) = { alias: ["POST_TYPE_A"] }];
                                                                  ^
missing_enum_alias.proto:52:62: use "alias" in "grpc.federation.enum_value" option, but "alias" is not defined in "grpc.federation.enum" option
52:    POST_TYPE_BAR = 2 [(grpc.federation.enum_value) = { alias: ["POST_TYPE_B", "POST_TYPE_C"] }];
                                                                  ^
missing_enum_alias.proto:66:3: required specify alias = "org.post.PostContent.Category" in grpc.federation.enum option for the "org.federation.PostContent.Category" type to automatically assign a value to the "PostContent.category" field via autobind
66:    enum Category {
       ^
`},
		{file: "missing_enum_value.proto", expected: `
missing_enum_value.proto:52:3: specified "alias" in grpc.federation.enum option, but "FOO" value does not exist in "org.post.PostDataType" enum
52:    FOO = 0;
       ^
missing_enum_value.proto:69:5: specified "alias" in grpc.federation.enum option, but "CATEGORY_C" value does not exist in "org.post.PostContent.Category" enum
69:      CATEGORY_C = 2;
         ^
`},
		{file: "missing_enum_value_alias.proto", expected: `
missing_enum_value_alias.proto:52:3: specified "alias" in grpc.federation.enum option, but "POST_TYPE_UNKNOWN" value does not exist in "org.post.PostDataType" enum
52:    POST_TYPE_UNKNOWN = 0;
       ^
missing_enum_value_alias.proto:53:3: specified "alias" in grpc.federation.enum option, but "POST_TYPE_FOO" value does not exist in "org.post.PostDataType" enum
53:    POST_TYPE_FOO = 1;
       ^
missing_enum_value_alias.proto:54:3: specified "alias" in grpc.federation.enum option, but "POST_TYPE_BAR" value does not exist in "org.post.PostDataType" enum
54:    POST_TYPE_BAR = 2;
       ^
`},
		{file: "valid_enum_value_reference.proto", expected: ""},
		{file: "missing_message_field_alias.proto", expected: `
missing_message_field_alias.proto:78:3: specified "alias" in grpc.federation.message option, but "dup_body" field does not exist in "org.post.PostContent" message
78:    string dup_body = 4;
       ^
missing_message_field_alias.proto:78:3: "dup_body" field in "org.federation.PostContent" message needs to specify "grpc.federation.field" option
78:    string dup_body = 4;
       ^
`},
		{file: "missing_message_option.proto", expected: `
missing_message_option.proto:48:17: "federation.User" message does not specify "grpc.federation.message" option
48:            name: "User"
                     ^
`},
		{file: "missing_method_request_value.proto", expected: `
missing_method_request_value.proto:41:19: value must be specified
41:            request { field: "id" }
                       ^
`},
		{file: "missing_response_message_option.proto", expected: `
missing_response_message_option.proto:18:1: "federation.GetPostResponse" message needs to specify "grpc.federation.message" option
18:  message GetPostResponse {
     ^
`},
		{file: "missing_service_variable.proto", expected: `
missing_service_variable.proto:21:11: ERROR: <input>:1:1: undeclared reference to 'foo2' (in container 'org.federation')
 | foo2 + 1
 | ^
21:        by: "foo2 + 1"
               ^
missing_service_variable.proto:39:11: ERROR: <input>:1:20: undefined field 'unknown'
 | grpc.federation.var.unknown
 | ...................^
39:        by: "grpc.federation.var.unknown"
               ^
`},
		{file: "invalid_method_response_option.proto", expected: `
invalid_method_response_option.proto: "google.protobuf.Empty" message needs to specify "grpc.federation.message" option
invalid_method_response_option.proto:14:48: "federation.Invalid" message does not exist
14:      option (grpc.federation.method).response = "Invalid";
                                                    ^
invalid_method_response_option.proto:17:48: "federation.DeletePostResponse" message must contain fields with the same names and types as the "seconds", "nanos" fields in the "google.protobuf.Timestamp" message
17:      option (grpc.federation.method).response = "DeletePostResponse";
                                                    ^
`},
		{file: "invalid_method_response.proto", expected: `
invalid_method_response.proto:43:27: ERROR: <input>:1:1: undeclared reference to 'invalid' (in container 'federation')
 | invalid
 | ^
43:        { name: "post", by: "invalid", autobind: true  },
                               ^
invalid_method_response.proto:48:26: ERROR: <input>:1:1: undeclared reference to 'post' (in container 'federation')
 | post
 | ^
48:            args { inline: "post" }
                              ^
invalid_method_response.proto:53:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
53:    string id = 1;
       ^
invalid_method_response.proto:54:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
54:    string title = 2;
       ^
invalid_method_response.proto:55:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
55:    string content = 3;
       ^
invalid_method_response.proto:65:36: ERROR: <input>:1:8: undefined field 'user_id'
 | __ARG__.user_id
 | .......^
65:          request { field: "id", by: "$.user_id" }
                                        ^
`},
		{file: "invalid_message_name.proto", expected: `
invalid_message_name.proto:48:17: "federation.Invalid" message does not exist
48:          message {
                     ^
invalid_message_name.proto:49:17: undefined message specified
49:            name: "Invalid"
                     ^
invalid_message_name.proto:55:17: "post.Invalid" message does not exist
55:          message {
                     ^
invalid_message_name.proto:56:17: undefined message specified
56:            name: "post.Invalid"
                     ^
`},
		{file: "invalid_nested_message_name.proto", expected: `
invalid_nested_message_name.proto:36:31: "federation.Invalid1" message does not exist
36:          { name: "b1" message: { name: "Invalid1" } }
                                   ^
invalid_nested_message_name.proto:36:39: undefined message specified
36:          { name: "b1" message: { name: "Invalid1" } }
                                           ^
invalid_nested_message_name.proto:42:33: "federation.Invalid2" message does not exist
42:            { name: "c1" message: { name: "Invalid2" } }
                                     ^
invalid_nested_message_name.proto:42:41: undefined message specified
42:            { name: "c1" message: { name: "Invalid2" } }
                                             ^
invalid_nested_message_name.proto:45:7: cannot convert type automatically: field type is "string" but specified value type is "null"
45:        string c1 = 1 [(grpc.federation.field).by = "c1"];
           ^
invalid_nested_message_name.proto:47:5: cannot convert type automatically: field type is "string" but specified value type is "null"
47:      string b1 = 1 [(grpc.federation.field).by = "b1"];
         ^
`},
		{file: "invalid_message_argument.proto", expected: `
invalid_message_argument.proto:52:19: ERROR: <input>:1:11: type 'string' does not support field selection
 | __ARG__.id.invalid
 | ..........^
52:              { by: "$.id.invalid" },
                       ^
invalid_message_argument.proto:53:23: inline value is not message type
53:              { inline: "post.id" },
                           ^
invalid_message_argument.proto:54:19: ERROR: <input>:1:2: Syntax error: no viable alternative at input '..'
 | ....
 | .^
54:              { by: "...." },
                       ^
invalid_message_argument.proto:55:23: ERROR: <input>:1:2: Syntax error: no viable alternative at input '..'
 | ....
 | .^
55:              { inline: "...." }
                           ^
invalid_message_argument.proto:73:36: ERROR: <input>:1:8: undefined field 'user_id'
 | __ARG__.user_id
 | .......^
73:          request { field: "id", by: "$.user_id" }
                                        ^
invalid_message_argument.proto:88:14: "x" argument name is declared with a different type kind. found "string" and "bool" type
88:          args { name: "x" by: "true" }
                  ^
`},
		{file: "invalid_message_field_alias.proto", expected: `
invalid_message_field_alias.proto:61:3: The types of "org.federation.PostData"'s "title" field ("int64") and "org.post.PostData"'s field ("string") are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself
61:    int64 title = 2;
       ^
invalid_message_field_alias.proto:61:3: "title" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
61:    int64 title = 2;
       ^
invalid_message_field_alias.proto:71:3: cannot convert type automatically: field type is "enum" but specified value type is "int64"
71:    PostType type = 1 [(grpc.federation.field).by = "org.federation.PostType.POST_TYPE_FOO"];
       ^
invalid_message_field_alias.proto:88:3: The types of "org.federation.PostContent"'s "body" field ("int64") and "org.post.PostContent"'s field ("string") are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself
88:    int64 body = 3;
       ^
invalid_message_field_alias.proto:88:3: "body" field in "org.federation.PostContent" message needs to specify "grpc.federation.field" option
88:    int64 body = 3;
       ^
`},
		{file: "recursive_message_name.proto", expected: `
recursive_message_name.proto:35:1: found cyclic dependency in "federation.Post" message. dependency path: GetPostResponse => Post => Post
35:  message Post {
     ^
recursive_message_name.proto:46:39: recursive definition: "Post" is own message name
46:        { name: "self", message { name: "Post" } }
                                           ^
recursive_message_name.proto:49:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
49:    string id = 1;
       ^
recursive_message_name.proto:50:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
50:    string title = 2;
       ^
recursive_message_name.proto:51:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
51:    string content = 3;
       ^
`},
		{file: "message_cyclic_dependency.proto", expected: `
message_cyclic_dependency.proto:27:1: found cyclic dependency in "org.federation.A" message. dependency path: GetResponse => A => AA => AAA => A
27:  message A {
     ^
`},
		{file: "nested_message_cyclic_dependency.proto", expected: `
nested_message_cyclic_dependency.proto:50:5: found cyclic dependency in "federation.C" message. dependency path: GetAResponse => A => B => C => C
50:      message C {
         ^
nested_message_cyclic_dependency.proto:55:19: recursive definition: "C" is own message name
55:              name: "A.B.C"
                       ^
`},
		{file: "invalid_variable_name.proto", expected: `
invalid_variable_name.proto:22:15: "_def0" is invalid name. name should be in the following pattern: ^[a-zA-Z][a-zA-Z0-9_]*$
22:        { name: "_def0" by: "0" },
                   ^
invalid_variable_name.proto:26:25: "_def1" is invalid name. name should be in the following pattern: ^[a-zA-Z][a-zA-Z0-9_]*$
26:              def { name: "_def1" by: "1" }
                             ^
invalid_variable_name.proto:28:27: "_def2" is invalid name. name should be in the following pattern: ^[a-zA-Z][a-zA-Z0-9_]*$
28:                def { name: "_def2" by: "2" }
                               ^
invalid_variable_name.proto:37:25: "_def3" is invalid name. name should be in the following pattern: ^[a-zA-Z][a-zA-Z0-9_]*$
37:              def { name: "_def3" by: "3" }
                             ^
invalid_variable_name.proto:39:27: "_def4" is invalid name. name should be in the following pattern: ^[a-zA-Z][a-zA-Z0-9_]*$
39:                def { name: "_def4" by: "4" }
                               ^
invalid_variable_name.proto:49:19: "_def5" is invalid name. name should be in the following pattern: ^[a-zA-Z][a-zA-Z0-9_]*$
49:        def { name: "_def5" by: "5" }
                       ^
`},
		{file: "invalid_wrapper_type_conversion.proto", expected: `
invalid_wrapper_type_conversion.proto:20:3: cannot convert message to "double"
20:    double double_value = 1 [(grpc.federation.field).by = "google.protobuf.DoubleValue{value: 1.23}"];
       ^
invalid_wrapper_type_conversion.proto:21:3: cannot convert message to "float"
21:    float float_value = 2 [(grpc.federation.field).by = "google.protobuf.FloatValue{value: 3.45}"];
       ^
invalid_wrapper_type_conversion.proto:22:3: cannot convert message to "int64"
22:    int64 i64_value = 3 [(grpc.federation.field).by = "google.protobuf.Int64Value{value: 1}"];
       ^
invalid_wrapper_type_conversion.proto:23:3: cannot convert message to "uint64"
23:    uint64 u64_value = 4 [(grpc.federation.field).by = "google.protobuf.UInt64Value{value: uint(2)}"];
       ^
invalid_wrapper_type_conversion.proto:24:3: cannot convert message to "int32"
24:    int32 i32_value = 5 [(grpc.federation.field).by = "google.protobuf.Int32Value{value: 3}"];
       ^
invalid_wrapper_type_conversion.proto:25:3: cannot convert message to "uint32"
25:    uint32 u32_value = 6 [(grpc.federation.field).by = "google.protobuf.UInt32Value{value: uint(4)}"];
       ^
invalid_wrapper_type_conversion.proto:26:3: cannot convert message to "bool"
26:    bool bool_value = 7 [(grpc.federation.field).by = "google.protobuf.BoolValue{value: true}"];
       ^
invalid_wrapper_type_conversion.proto:27:3: cannot convert message to "string"
27:    string string_value = 8 [(grpc.federation.field).by = "google.protobuf.StringValue{value: 'hello'}"];
       ^
invalid_wrapper_type_conversion.proto:28:3: cannot convert message to "bytes"
28:    bytes bytes_value = 9 [(grpc.federation.field).by = "google.protobuf.BytesValue{value: bytes('world')}"];
       ^
invalid_wrapper_type_conversion.proto:40:3: cannot convert type automatically: field type is "message" but specified value type is "double"
40:    google.protobuf.DoubleValue double_wrapper_value2 = 19 [(grpc.federation.field).by = "1.23"];
       ^
invalid_wrapper_type_conversion.proto:41:3: cannot convert type automatically: field type is "message" but specified value type is "double"
41:    google.protobuf.FloatValue float_wrapper_value2 = 20 [(grpc.federation.field).by = "3.45"];
       ^
invalid_wrapper_type_conversion.proto:42:3: cannot convert type automatically: field type is "message" but specified value type is "int64"
42:    google.protobuf.Int64Value i64_wrapper_value2 = 21 [(grpc.federation.field).by = "1"];
       ^
invalid_wrapper_type_conversion.proto:43:3: cannot convert type automatically: field type is "message" but specified value type is "uint64"
43:    google.protobuf.UInt64Value u64_wrapper_value2 = 22 [(grpc.federation.field).by = "uint(2)"];
       ^
invalid_wrapper_type_conversion.proto:44:3: cannot convert type automatically: field type is "message" but specified value type is "int64"
44:    google.protobuf.Int32Value i32_wrapper_value2 = 23 [(grpc.federation.field).by = "3"];
       ^
invalid_wrapper_type_conversion.proto:45:3: cannot convert type automatically: field type is "message" but specified value type is "uint64"
45:    google.protobuf.UInt32Value u32_wrapper_value2 = 24 [(grpc.federation.field).by = "uint(4)"];
       ^
invalid_wrapper_type_conversion.proto:46:3: cannot convert type automatically: field type is "message" but specified value type is "bool"
46:    google.protobuf.BoolValue bool_wrapper_value2 = 25 [(grpc.federation.field).by = "true"];
       ^
invalid_wrapper_type_conversion.proto:47:3: cannot convert type automatically: field type is "message" but specified value type is "string"
47:    google.protobuf.StringValue string_wrapper_value2 = 26 [(grpc.federation.field).by = "'hello'"];
       ^
invalid_wrapper_type_conversion.proto:48:3: cannot convert type automatically: field type is "message" but specified value type is "bytes"
48:    google.protobuf.BytesValue bytes_wrapper_value2 = 27 [(grpc.federation.field).by = "bytes('world')"];
       ^
`},
		{file: "invalid_validation_return_type.proto", expected: `
invalid_validation_return_type.proto:50:17: if must always return a boolean value
50:              if: "post.id"
                     ^
`},
		{file: "invalid_validation_details_return_type.proto", expected: `
invalid_validation_details_return_type.proto:51:19: if must always return a boolean value
51:                if: "'string'"
                       ^
`},
		{file: "invalid_validation_message_argument.proto", expected: `
invalid_validation_message_argument.proto:71:52: ERROR: <input>:1:8: undefined field 'message'
 | __ARG__.message
 | .......^
71:    string message = 1 [(grpc.federation.field).by = "$.message"];
                                                        ^
`},
		{file: "invalid_validation_precondition_failure.proto", expected: `
invalid_validation_precondition_failure.proto:54:25: type must always return a string value
54:                    type: "1",
                             ^
invalid_validation_precondition_failure.proto:55:28: subject must always return a string value
55:                    subject: "2",
                                ^
invalid_validation_precondition_failure.proto:56:32: description must always return a string value
56:                    description: "3",
                                    ^
`},
		{file: "invalid_validation_bad_request.proto", expected: `
invalid_validation_bad_request.proto:54:26: field must always return a string value
54:                    field: "1",
                              ^
invalid_validation_bad_request.proto:55:32: description must always return a string value
55:                    description: "2",
                                    ^
`},
		{file: "invalid_validation_localized_message.proto", expected: `
invalid_validation_localized_message.proto:54:26: message must always return a string value
54:                  message: "1"
                              ^
`},
		{file: "invalid_validation_with_ignore.proto", expected: `
invalid_validation_with_ignore.proto:24:19: validation doesn't support "ignore" feature
24:            ignore: true
                       ^
invalid_validation_with_ignore.proto:32:32: validation doesn't support "ignore_and_response" feature
32:            ignore_and_response: "'foo'"
                                    ^
`},
		{file: "invalid_list_sort.proto", expected: `
invalid_list_sort.proto:55:59: ERROR: <input>:1:14: list(org.federation.User) is not comparable
 | users.sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)
 | .............^
ERROR: <input>:1:29: list(org.federation.User) is not comparable
 | users.sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)
 | ............................^
ERROR: <input>:1:49: list(org.federation.User) is not comparable
 | users.sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)
 | ................................................^
ERROR: <input>:1:70: list(org.federation.User) is not comparable
 | users.sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)
 | .....................................................................^
55:    repeated User invalid = 3 [(grpc.federation.field).by = "users.sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)"];
                                                               ^
`},
		{file: "invalid_message_map.proto", expected: `
invalid_message_map.proto:32:3: cannot convert type automatically: map key type is "int32" but specified map key type is "string"
32:    map<int32, int32> map_value = 1 [(grpc.federation.field).by = "map_value"];
       ^
invalid_message_map.proto:32:3: cannot convert type automatically: map value type is "int32" but specified map value type is "string"
32:    map<int32, int32> map_value = 1 [(grpc.federation.field).by = "map_value"];
       ^
invalid_message_map.proto:42:19: cannot convert type automatically: map key type is "string" but specified map key type is "int64"
42:            request { field: "ids", by: "$.ids" }
                       ^
invalid_message_map.proto:42:19: cannot convert type automatically: map value type is "string" but specified map value type is "int64"
42:            request { field: "ids", by: "$.ids" }
                       ^
`},
		{file: "invalid_message_map_alias.proto", expected: `
invalid_message_map_alias.proto:39:3: cannot convert type automatically: map key type is "string" but specified map key type is "int32"
39:    map<string, string> counts = 4;
       ^
invalid_message_map_alias.proto:39:3: cannot convert type automatically: map value type is "string" but specified map value type is "int32"
39:    map<string, string> counts = 4;
       ^
`},
		{file: "invalid_file_import.proto", expected: `
invalid_file_import.proto:6:8: [WARN] Import post.proto is used only for the definition of grpc federation. You can use grpc.federation.file.import instead.
6:  import "post.proto";
           ^
invalid_file_import.proto:12:5: [WARN] Import user.proto is unused for the definition of grpc federation.
12:      "user.proto"
         ^
`},
	}
	for _, test := range tests {
		test := test
		t.Run(test.file, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			v := validator.New()
			path := filepath.Join("testdata", test.file)
			f, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			file, err := source.NewFile(path, f)
			if err != nil {
				t.Fatal(err)
			}
			got := validator.Format(v.Validate(ctx, file, validator.ImportPathOption("testdata")))
			if test.expected == "" {
				if got != "" {
					t.Errorf("expected to receive no validation error but got: %s", got)
				}
				return
			}
			if diff := cmp.Diff("\n"+got, test.expected); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}
}
