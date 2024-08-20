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
		{file: "empty_response_field.proto", expected: `
`},
		{file: "different_message_argument_type.proto", expected: `
testdata/different_message_argument_type.proto:28:14: "id" argument name is declared with a different type kind. found "string" and "int64" type
28:          args { name: "id" by: "1" }
                  ^
`},
		{file: "duplicated_variable_name.proto", expected: `
testdata/duplicated_variable_name.proto:25:25: found duplicated variable name "a"
25:              def { name: "a" by: "1" }
                             ^
testdata/duplicated_variable_name.proto:27:27: found duplicated variable name "a"
27:                def { name: "a" by: "2" }
                               ^
testdata/duplicated_variable_name.proto:36:25: found duplicated variable name "a"
36:              def { name: "a" by: "3" }
                             ^
testdata/duplicated_variable_name.proto:38:27: found duplicated variable name "a"
38:                def { name: "a" by: "4" }
                               ^
testdata/duplicated_variable_name.proto:48:19: found duplicated variable name "a"
48:        def { name: "a" by: "5" }
                       ^
`},
		{file: "invalid_autobind.proto", expected: `
testdata/invalid_autobind.proto:23:3: "id" field found multiple times in the message specified by autobind. since it is not possible to determine one, please use "grpc.federation.field" to explicitly bind it. found message names are "a" name at def and "b" name at def
23:    string id = 1;
       ^
testdata/invalid_autobind.proto:23:3: "id" field in "org.federation.GetResponse" message needs to specify "grpc.federation.field" option
23:    string id = 1;
       ^
`},
		{file: "invalid_call_error_handler.proto", expected: `
testdata/invalid_call_error_handler.proto:42:21: cannot set both "ignore" and "ignore_and_response"
42:              ignore: true
                         ^
testdata/invalid_call_error_handler.proto:43:34: cannot set both "ignore" and "ignore_and_response"
43:              ignore_and_response: "post.GetPostResponse{}"
                                      ^
testdata/invalid_call_error_handler.proto:47:19: "by" must always return a message value
47:                by: "1"
                       ^
testdata/invalid_call_error_handler.proto:51:34: value must be "post.GetPostResponse" type
51:              ignore_and_response: "10"
                                      ^
`},
		{file: "invalid_condition_type.proto", expected: `
testdata/invalid_condition_type.proto:36:13: return value of "if" must be bool type but got string type
36:          if: "$.id"
                 ^
`},
		{file: "invalid_field_option.proto", expected: `
testdata/invalid_field_option.proto:30:50: ERROR: <input>:1:5: undefined field 'invalid'
 | post.invalid
 | ....^
30:    Post post = 1 [(grpc.federation.field) = { by: "post.invalid" }];
                                                      ^
`},
		{file: "invalid_field_type.proto", expected: `
testdata/invalid_field_type.proto:18:3: cannot convert type automatically: field type is "string" but specified value type is "int64"
18:    string a = 1 [(grpc.federation.field).by = "1"];
       ^
`},
		{file: "invalid_go_package.proto", expected: `
testdata/invalid_go_package.proto:9:21: go_package option "a;b;c;d" is invalid
9:  option go_package = "a;b;c;d";
                        ^
`},
		{file: "invalid_enum_alias_target.proto", expected: `
testdata/invalid_enum_alias_target.proto:47:1: required specify alias = "org.post.PostDataType" in grpc.federation.enum option for the "org.federation.PostType" type to automatically assign a value to the "PostData.type" field via autobind
47:  enum PostType {
     ^
testdata/invalid_enum_alias_target.proto:66:3: required specify alias = "org.post.PostContent.Category" in grpc.federation.enum option for the "org.federation.PostContent.Category" type to automatically assign a value to the "PostContent.category" field via autobind
66:    enum Category {
       ^
`},
		{file: "invalid_enum_selector.proto", expected: `
testdata/invalid_enum_selector.proto:20:15: ERROR: <input>:1:56: cannot specify an int type. if you are directly specifying an enum value, you need to explicitly use "pkg.EnumName.value('ENUM_VALUE')" function to use the enum type
 | grpc.federation.enum.select(true, org.post.PostDataType.POST_TYPE_B, 'foo')
 | .......................................................^
20:      def { by: "grpc.federation.enum.select(true, org.post.PostDataType.POST_TYPE_B, 'foo')" }
                   ^
`},
		{file: "invalid_env.proto", expected: `
testdata/invalid_env.proto:11:9: "message" and "var" cannot be used simultaneously
11:      env {
             ^
testdata/invalid_env.proto:24:16: "org.federation.Invalid" message does not exist
24:        message: "Invalid"
                    ^
`},
		{file: "invalid_multiple_env.proto", expected: `
testdata/invalid_multiple_env.proto:56:1: environment variable "ccc" has different types across services: InlineEnvService, RefEnvService
56:  message GetNameResponse {
     ^
testdata/invalid_multiple_env.proto:56:1: environment variable "eee" has different types across services: InlineEnvService, RefEnvService
56:  message GetNameResponse {
     ^
`},
		{file: "invalid_error_variable.proto", expected: `
testdata/invalid_error_variable.proto:20:17: "error" is the reserved keyword. this name is not available
20:      def { name: "error" by: "'foo'" }
                     ^
testdata/invalid_error_variable.proto:21:25: ERROR: <input>:1:1: undeclared reference to 'error' (in container 'org.federation')
 | error
 | ^
21:      def { name: "e" by: "error" }
                             ^
testdata/invalid_error_variable.proto:25:15: ERROR: <input>:1:1: undeclared reference to 'error' (in container 'org.federation')
 | error.code == 0
 | ^
25:            if: "error.code == 0"
                   ^
`},
		{file: "invalid_map_iterator_src_type.proto", expected: `
testdata/invalid_map_iterator_src_type.proto:40:13: map iterator's src value type must be repeated type
40:          map {
                 ^
testdata/invalid_map_iterator_src_type.proto:54:57: ERROR: <input>:1:1: undeclared reference to 'users' (in container 'org.federation')
 | users
 | ^
54:    repeated User users = 4 [(grpc.federation.field).by = "users"];
                                                             ^
testdata/invalid_map_iterator_src_type.proto:58:47: ERROR: <input>:1:8: undefined field 'user_id'
 | __ARG__.user_id
 | .......^
58:    string id = 1 [(grpc.federation.field).by = "$.user_id"];
                                                   ^
`},
		{file: "invalid_map_iterator_src.proto", expected: `
testdata/invalid_map_iterator_src.proto:36:13: "posts" variable is not defined
36:          map {
                 ^
testdata/invalid_map_iterator_src.proto:36:13: ERROR: <input>:1:1: undeclared reference to 'iter' (in container 'org.federation')
 | iter.id
 | ^
36:          map {
                 ^
testdata/invalid_map_iterator_src.proto:54:47: ERROR: <input>:1:8: undefined field 'user_id'
 | __ARG__.user_id
 | .......^
54:    string id = 1 [(grpc.federation.field).by = "$.user_id"];
                                                   ^
`},
		{file: "invalid_message_alias_target.proto", expected: `
testdata/invalid_message_alias_target.proto:44:3: required specify alias = "org.post.PostData" in grpc.federation.message option for the "org.federation.PostData" type to automatically assign a value to the "Post.data" field via autobind
44:    PostData data = 4;
       ^
`},
		{file: "invalid_message_alias.proto", expected: `
testdata/invalid_message_alias.proto:44:3: required specify alias = "org.post.PostData" in grpc.federation.message option for the "org.federation.PostData" type to automatically assign a value to the "Post.data" field via autobind
44:    PostData data = 4;
       ^
testdata/invalid_message_alias.proto:54:44: cannot find package from "invalid.Invalid"
54:    option (grpc.federation.message).alias = "invalid.Invalid";
                                                ^
testdata/invalid_message_alias.proto:56:3: "type" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
56:    PostType type = 1;
       ^
testdata/invalid_message_alias.proto:57:3: "title" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
57:    string title = 2;
       ^
testdata/invalid_message_alias.proto:58:3: "content" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
58:    PostContent content = 3;
       ^
testdata/invalid_message_alias.proto:73:3: "org.federation.SomeUser" message does not exist
73:    option (grpc.federation.message).alias = "SomeUser";
       ^
testdata/invalid_message_alias.proto:75:3: "name" field in "org.federation.User" message needs to specify "grpc.federation.field" option
75:    string name = 1;
       ^
testdata/invalid_message_alias.proto:79:3: "google.protobuf.Comment" message does not exist
79:    option (grpc.federation.message).alias = "google.protobuf.Comment";
       ^
testdata/invalid_message_alias.proto:81:3: "body" field in "org.federation.Comment" message needs to specify "grpc.federation.field" option
81:    string body = 1;
       ^
`},
		{file: "invalid_nested_message_field.proto", expected: `
testdata/invalid_nested_message_field.proto:52:7: "body" field in "federation.A.B.C" message needs to specify "grpc.federation.field" option
52:        string body = 1;
           ^
`},
		{file: "invalid_method.proto", expected: `
testdata/invalid_method.proto:36:24: invalid method format. required format is "<package-name>.<service-name>/<method-name>" but specified ""
36:        { call { method: "" } },
                            ^
testdata/invalid_method.proto:41:26: ERROR: <input>:1:1: undeclared reference to 'invalid' (in container 'federation')
 | invalid
 | ^
41:            args { inline: "invalid" }
                              ^
testdata/invalid_method.proto:46:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
46:    string id = 1;
       ^
testdata/invalid_method.proto:47:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
47:    string title = 2;
       ^
testdata/invalid_method.proto:48:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
48:    string content = 3;
       ^
testdata/invalid_method.proto:58:36: ERROR: <input>:1:8: undefined field 'user_id'
 | __ARG__.user_id
 | .......^
58:          request { field: "id", by: "$.user_id" }
                                        ^
`},
		{file: "invalid_multi_alias.proto", expected: `
testdata/invalid_multi_alias.proto:55:3: if multiple aliases are specified, you must use grpc.federation.enum.select function to bind
55:    PostType post_type = 4 [(grpc.federation.field).by = "org.post.PostDataType.POST_TYPE_A"];
       ^
testdata/invalid_multi_alias.proto:64:3: "POST_TYPE_A" value must be present in all enums, but it is missing in "org.post.PostDataType", "org.post.v2.PostDataType" enum
64:    POST_TYPE_FOO = 1 [(grpc.federation.enum_value) = { alias: ["POST_TYPE_A"] }];
       ^
testdata/invalid_multi_alias.proto:65:3: "org.post.v2.PostDataType.POST_TYPE_B" value does not exist in "org.post.PostDataType", "org.post.v2.PostDataType" enum
65:    POST_TYPE_BAR = 2 [(grpc.federation.enum_value) = { alias: ["org.post.v2.PostDataType.POST_TYPE_B", "POST_TYPE_C"] }];
       ^
testdata/invalid_multi_alias.proto:65:3: "POST_TYPE_C" value must be present in all enums, but it is missing in "org.post.PostDataType", "org.post.v2.PostDataType" enum
65:    POST_TYPE_BAR = 2 [(grpc.federation.enum_value) = { alias: ["org.post.v2.PostDataType.POST_TYPE_B", "POST_TYPE_C"] }];
       ^
testdata/invalid_multi_alias.proto:74:3: The types of "org.federation.PostData"'s "title" field ("string") and "org.post.v2.PostData"'s field ("int64") are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself
74:    string title = 2;
       ^
testdata/invalid_multi_alias.proto:74:3: "title" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
74:    string title = 2;
       ^
testdata/invalid_multi_alias.proto:75:3: required specify alias = "org.post.v2.PostContent" in grpc.federation.message option for the "org.federation.PostContent" type to automatically assign a value to the "PostData.content" field via autobind
75:    PostContent content = 3;
       ^
testdata/invalid_multi_alias.proto:76:3: specified "alias" in grpc.federation.message option, but "dummy" field does not exist in "org.post.PostData" message
76:    int64 dummy = 4;
       ^
testdata/invalid_multi_alias.proto:76:3: "dummy" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
76:    int64 dummy = 4;
       ^
`},
		{file: "invalid_oneof_selection.proto", expected: `
testdata/invalid_oneof_selection.proto:26:47: "org.federation.UserSelection" type has "user" as oneof name, but "user" has a difference type and cannot be accessed directly, so "user" becomes an undefined field
ERROR: <input>:1:4: undefined field 'user'
 | sel.user
 | ...^
26:    User user = 1 [(grpc.federation.field).by = "sel.user"];
                                                   ^
`},
		{file: "invalid_oneof.proto", expected: `
testdata/invalid_oneof.proto:40:13: return value of "if" must be bool type but got int64 type
40:          if: "1"
                 ^
testdata/invalid_oneof.proto:53:39: "if" or "default" must be specified in "grpc.federation.field.oneof"
53:        (grpc.federation.field).oneof = {
                                           ^
testdata/invalid_oneof.proto:66:39: "by" must be specified in "grpc.federation.field.oneof"
66:        (grpc.federation.field).oneof = {
                                           ^
testdata/invalid_oneof.proto:80:18: "default" found multiple times in the "grpc.federation.field.oneof". "default" can only be specified once per oneof
80:          default: true
                      ^
testdata/invalid_oneof.proto:92:3: "oneof" feature can only be used for fields within oneof
92:    bool foo = 5 [(grpc.federation.field).oneof = {
       ^
testdata/invalid_oneof.proto:92:3: value must be specified
92:    bool foo = 5 [(grpc.federation.field).oneof = {
       ^
testdata/invalid_oneof.proto:92:3: "foo" field in "org.federation.UserSelection" message needs to specify "grpc.federation.field" option
92:    bool foo = 5 [(grpc.federation.field).oneof = {
       ^
testdata/invalid_oneof.proto:109:20: "foo" field is a oneof field, so you need to specify an "if" expression
109:            { field: "foo" by: "1" },
                         ^
testdata/invalid_oneof.proto:110:20: "bar" field is a oneof field, so you need to specify an "if" expression
110:            { field: "bar" by: "'hello'" }
                         ^
`},
		{file: "invalid_retry.proto", expected: `
testdata/invalid_retry.proto:37:15: ERROR: <input>:1:1: undeclared reference to 'foo' (in container 'org.federation')
 | foo
 | ^
37:            if: "foo"
                   ^
testdata/invalid_retry.proto:39:23: time: missing unit in duration "1"
39:              interval: "1"
                           ^
testdata/invalid_retry.proto:54:15: if must always return a boolean value
54:            if: "1"
                   ^
testdata/invalid_retry.proto:56:31: time: missing unit in duration "2"
56:              initial_interval: "2"
                                   ^
testdata/invalid_retry.proto:72:27: time: missing unit in duration "3"
72:              max_interval: "3"
                               ^
`},
		{file: "invalid_method_service_name.proto", expected: `
testdata/invalid_method_service_name.proto: "post.InvalidService" service does not exist
testdata/invalid_method_service_name.proto:36:24: cannot find "method" method because the service to which the method belongs does not exist
36:        { call { method: "post.InvalidService/method" } },
                            ^
testdata/invalid_method_service_name.proto:46:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
46:    string id = 1;
       ^
testdata/invalid_method_service_name.proto:47:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
47:    string title = 2;
       ^
testdata/invalid_method_service_name.proto:48:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
48:    string content = 3;
       ^
`},
		{file: "invalid_method_name.proto", expected: `
testdata/invalid_method_name.proto:36:24: "invalid" method does not exist in PostService service
36:        { call { method: "post.PostService/invalid" } },
                            ^
testdata/invalid_method_name.proto:46:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
46:    string id = 1;
       ^
testdata/invalid_method_name.proto:47:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
47:    string title = 2;
       ^
testdata/invalid_method_name.proto:48:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
48:    string content = 3;
       ^
`},
		{file: "invalid_method_timeout_format.proto", expected: `
testdata/invalid_method_timeout_format.proto:12:47: time: unknown unit "p" in duration "1p"
12:      option (grpc.federation.method).timeout = "1p";
                                                   ^
`},

		{file: "invalid_method_request.proto", expected: `
testdata/invalid_method_request.proto:40:28: "invalid" field does not exist in "post.GetPostRequest" message for method request
40:            request { field: "invalid", by: "$.invalid" }
                                ^
testdata/invalid_method_request.proto:40:43: ERROR: <input>:1:8: undefined field 'invalid'
 | __ARG__.invalid
 | .......^
40:            request { field: "invalid", by: "$.invalid" }
                                               ^
`},
		{file: "missing_field_option.proto", expected: `
testdata/missing_field_option.proto:56:3: "user" field in "federation.Post" message needs to specify "grpc.federation.field" option
56:    User user = 4;
       ^
`},
		{file: "missing_map_iterator.proto", expected: `
testdata/missing_map_iterator.proto:36:13: map iterator name must be specified
36:          map {
                 ^
testdata/missing_map_iterator.proto:36:13: map iterator src must be specified
36:          map {
                 ^
`},
		{file: "missing_message_alias.proto", expected: `
testdata/missing_message_alias.proto:44:3: required specify alias = "org.post.PostData" in grpc.federation.message option for the "org.federation.PostData" type to automatically assign a value to the "Post.data" field via autobind
44:    PostData data = 4;
       ^
testdata/missing_message_alias.proto:54:3: "type" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
54:    PostType type = 1;
       ^
testdata/missing_message_alias.proto:55:3: "title" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
55:    string title = 2;
       ^
testdata/missing_message_alias.proto:56:3: use "alias" in "grpc.federation.field" option, but "alias" is not defined in "grpc.federation.message" option
56:    PostContent content = 3 [(grpc.federation.field).alias = "content"];
       ^
`},
		{file: "missing_enum_alias.proto", expected: `
testdata/missing_enum_alias.proto:47:1: required specify alias = "org.post.PostDataType" in grpc.federation.enum option for the "org.federation.PostType" type to automatically assign a value to the "PostData.type" field via autobind
47:  enum PostType {
     ^
testdata/missing_enum_alias.proto:49:62: use "alias" in "grpc.federation.enum_value" option, but "alias" is not defined in "grpc.federation.enum" option
49:    POST_TYPE_FOO = 1 [(grpc.federation.enum_value) = { alias: ["POST_TYPE_A"] }];
                                                                  ^
testdata/missing_enum_alias.proto:50:62: use "alias" in "grpc.federation.enum_value" option, but "alias" is not defined in "grpc.federation.enum" option
50:    POST_TYPE_BAR = 2 [(grpc.federation.enum_value) = { alias: ["POST_TYPE_B", "POST_TYPE_C"] }];
                                                                  ^
testdata/missing_enum_alias.proto:64:3: required specify alias = "org.post.PostContent.Category" in grpc.federation.enum option for the "org.federation.PostContent.Category" type to automatically assign a value to the "PostContent.category" field via autobind
64:    enum Category {
       ^
`},
		{file: "missing_enum_value.proto", expected: `
testdata/missing_enum_value.proto:50:3: specified "alias" in grpc.federation.enum option, but "FOO" value does not exist in "org.post.PostDataType" enum
50:    FOO = 0;
       ^
testdata/missing_enum_value.proto:67:5: specified "alias" in grpc.federation.enum option, but "CATEGORY_C" value does not exist in "org.post.PostContent.Category" enum
67:      CATEGORY_C = 2;
         ^
`},
		{file: "missing_enum_value_alias.proto", expected: `
testdata/missing_enum_value_alias.proto:50:3: specified "alias" in grpc.federation.enum option, but "POST_TYPE_UNKNOWN" value does not exist in "org.post.PostDataType" enum
50:    POST_TYPE_UNKNOWN = 0;
       ^
testdata/missing_enum_value_alias.proto:51:3: specified "alias" in grpc.federation.enum option, but "POST_TYPE_FOO" value does not exist in "org.post.PostDataType" enum
51:    POST_TYPE_FOO = 1;
       ^
testdata/missing_enum_value_alias.proto:52:3: specified "alias" in grpc.federation.enum option, but "POST_TYPE_BAR" value does not exist in "org.post.PostDataType" enum
52:    POST_TYPE_BAR = 2;
       ^
`},
		{file: "valid_enum_value_reference.proto", expected: ""},
		{file: "missing_message_field_alias.proto", expected: `
testdata/missing_message_field_alias.proto:76:3: specified "alias" in grpc.federation.message option, but "dup_body" field does not exist in "org.post.PostContent" message
76:    string dup_body = 4;
       ^
testdata/missing_message_field_alias.proto:76:3: "dup_body" field in "org.federation.PostContent" message needs to specify "grpc.federation.field" option
76:    string dup_body = 4;
       ^
`},
		{file: "missing_message_option.proto", expected: `
testdata/missing_message_option.proto:47:17: "federation.User" message does not specify "grpc.federation.message" option
47:            name: "User"
                     ^
`},
		{file: "missing_method_request_value.proto", expected: `
testdata/missing_method_request_value.proto:40:19: value must be specified
40:            request { field: "id" }
                       ^
`},
		{file: "missing_response_message_option.proto", expected: `
testdata/missing_response_message_option.proto:18:1: "federation.GetPostResponse" message needs to specify "grpc.federation.message" option
18:  message GetPostResponse {
     ^
`},
		{file: "invalid_method_response.proto", expected: `
testdata/invalid_method_response.proto:42:27: ERROR: <input>:1:1: undeclared reference to 'invalid' (in container 'federation')
 | invalid
 | ^
42:        { name: "post", by: "invalid", autobind: true  },
                               ^
testdata/invalid_method_response.proto:47:26: ERROR: <input>:1:1: undeclared reference to 'post' (in container 'federation')
 | post
 | ^
47:            args { inline: "post" }
                              ^
testdata/invalid_method_response.proto:52:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
52:    string id = 1;
       ^
testdata/invalid_method_response.proto:53:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
53:    string title = 2;
       ^
testdata/invalid_method_response.proto:54:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
54:    string content = 3;
       ^
testdata/invalid_method_response.proto:64:36: ERROR: <input>:1:8: undefined field 'user_id'
 | __ARG__.user_id
 | .......^
64:          request { field: "id", by: "$.user_id" }
                                        ^
`},
		{file: "invalid_message_name.proto", expected: `
testdata/invalid_message_name.proto:46:17: "federation.Invalid" message does not exist
46:          message {
                     ^
testdata/invalid_message_name.proto:47:17: undefined message specified
47:            name: "Invalid"
                     ^
testdata/invalid_message_name.proto:53:17: "post.Invalid" message does not exist
53:          message {
                     ^
testdata/invalid_message_name.proto:54:17: undefined message specified
54:            name: "post.Invalid"
                     ^
`},
		{file: "invalid_nested_message_name.proto", expected: `
testdata/invalid_nested_message_name.proto:36:31: "federation.Invalid1" message does not exist
36:          { name: "b1" message: { name: "Invalid1" } }
                                   ^
testdata/invalid_nested_message_name.proto:36:39: undefined message specified
36:          { name: "b1" message: { name: "Invalid1" } }
                                           ^
testdata/invalid_nested_message_name.proto:42:33: "federation.Invalid2" message does not exist
42:            { name: "c1" message: { name: "Invalid2" } }
                                     ^
testdata/invalid_nested_message_name.proto:42:41: undefined message specified
42:            { name: "c1" message: { name: "Invalid2" } }
                                             ^
testdata/invalid_nested_message_name.proto:45:7: cannot convert type automatically: field type is "string" but specified value type is "null"
45:        string c1 = 1 [(grpc.federation.field).by = "c1"];
           ^
testdata/invalid_nested_message_name.proto:47:5: cannot convert type automatically: field type is "string" but specified value type is "null"
47:      string b1 = 1 [(grpc.federation.field).by = "b1"];
         ^
`},
		{file: "invalid_message_argument.proto", expected: `
testdata/invalid_message_argument.proto:51:19: ERROR: <input>:1:11: type 'string' does not support field selection
 | __ARG__.id.invalid
 | ..........^
51:              { by: "$.id.invalid" },
                       ^
testdata/invalid_message_argument.proto:52:23: inline value is not message type
52:              { inline: "post.id" },
                           ^
testdata/invalid_message_argument.proto:53:19: ERROR: <input>:1:2: Syntax error: no viable alternative at input '..'
 | ....
 | .^
53:              { by: "...." },
                       ^
testdata/invalid_message_argument.proto:54:23: ERROR: <input>:1:2: Syntax error: no viable alternative at input '..'
 | ....
 | .^
54:              { inline: "...." }
                           ^
testdata/invalid_message_argument.proto:72:36: ERROR: <input>:1:8: undefined field 'user_id'
 | __ARG__.user_id
 | .......^
72:          request { field: "id", by: "$.user_id" }
                                        ^
testdata/invalid_message_argument.proto:87:14: "x" argument name is declared with a different type kind. found "string" and "bool" type
87:          args { name: "x" by: "true" }
                  ^
`},
		{file: "invalid_message_field_alias.proto", expected: `
testdata/invalid_message_field_alias.proto:59:3: The types of "org.federation.PostData"'s "title" field ("int64") and "org.post.PostData"'s field ("string") are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself
59:    int64 title = 2;
       ^
testdata/invalid_message_field_alias.proto:59:3: "title" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
59:    int64 title = 2;
       ^
testdata/invalid_message_field_alias.proto:86:3: The types of "org.federation.PostContent"'s "body" field ("int64") and "org.post.PostContent"'s field ("string") are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself
86:    int64 body = 3;
       ^
testdata/invalid_message_field_alias.proto:86:3: "body" field in "org.federation.PostContent" message needs to specify "grpc.federation.field" option
86:    int64 body = 3;
       ^
`},
		{file: "recursive_message_name.proto", expected: `
testdata/recursive_message_name.proto:33:1: found cyclic dependency in "federation.Post" message. dependency path: GetPostResponse => Post => Post
33:  message Post {
     ^
testdata/recursive_message_name.proto:44:39: recursive definition: "Post" is own message name
44:        { name: "self", message { name: "Post" } }
                                           ^
testdata/recursive_message_name.proto:47:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
47:    string id = 1;
       ^
testdata/recursive_message_name.proto:48:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
48:    string title = 2;
       ^
testdata/recursive_message_name.proto:49:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
49:    string content = 3;
       ^
`},
		{file: "message_cyclic_dependency.proto", expected: `
testdata/message_cyclic_dependency.proto:27:1: found cyclic dependency in "org.federation.A" message. dependency path: GetResponse => A => AA => AAA => A
27:  message A {
     ^
`},
		{file: "nested_message_cyclic_dependency.proto", expected: `
testdata/nested_message_cyclic_dependency.proto:50:5: found cyclic dependency in "federation.C" message. dependency path: GetAResponse => A => B => C => C
50:      message C {
         ^
testdata/nested_message_cyclic_dependency.proto:55:19: recursive definition: "C" is own message name
55:              name: "A.B.C"
                       ^
`},
		{file: "invalid_variable_name.proto", expected: `
testdata/invalid_variable_name.proto:20:15: "_def0" is invalid name. name should be in the following pattern: ^[a-zA-Z][a-zA-Z0-9_]*$
20:        { name: "_def0" by: "0" },
                   ^
testdata/invalid_variable_name.proto:24:25: "_def1" is invalid name. name should be in the following pattern: ^[a-zA-Z][a-zA-Z0-9_]*$
24:              def { name: "_def1" by: "1" }
                             ^
testdata/invalid_variable_name.proto:26:27: "_def2" is invalid name. name should be in the following pattern: ^[a-zA-Z][a-zA-Z0-9_]*$
26:                def { name: "_def2" by: "2" }
                               ^
testdata/invalid_variable_name.proto:35:25: "_def3" is invalid name. name should be in the following pattern: ^[a-zA-Z][a-zA-Z0-9_]*$
35:              def { name: "_def3" by: "3" }
                             ^
testdata/invalid_variable_name.proto:37:27: "_def4" is invalid name. name should be in the following pattern: ^[a-zA-Z][a-zA-Z0-9_]*$
37:                def { name: "_def4" by: "4" }
                               ^
testdata/invalid_variable_name.proto:47:19: "_def5" is invalid name. name should be in the following pattern: ^[a-zA-Z][a-zA-Z0-9_]*$
47:        def { name: "_def5" by: "5" }
                       ^
`},
		{file: "invalid_wrapper_type_conversion.proto", expected: `
testdata/invalid_wrapper_type_conversion.proto:20:3: cannot convert message to "double"
20:    double double_value = 1 [(grpc.federation.field).by = "google.protobuf.DoubleValue{value: 1.23}"];
       ^
testdata/invalid_wrapper_type_conversion.proto:21:3: cannot convert message to "float"
21:    float float_value = 2 [(grpc.federation.field).by = "google.protobuf.FloatValue{value: 3.45}"];
       ^
testdata/invalid_wrapper_type_conversion.proto:22:3: cannot convert message to "int64"
22:    int64 i64_value = 3 [(grpc.federation.field).by = "google.protobuf.Int64Value{value: 1}"];
       ^
testdata/invalid_wrapper_type_conversion.proto:23:3: cannot convert message to "uint64"
23:    uint64 u64_value = 4 [(grpc.federation.field).by = "google.protobuf.UInt64Value{value: uint(2)}"];
       ^
testdata/invalid_wrapper_type_conversion.proto:24:3: cannot convert message to "int32"
24:    int32 i32_value = 5 [(grpc.federation.field).by = "google.protobuf.Int32Value{value: 3}"];
       ^
testdata/invalid_wrapper_type_conversion.proto:25:3: cannot convert message to "uint32"
25:    uint32 u32_value = 6 [(grpc.federation.field).by = "google.protobuf.UInt32Value{value: uint(4)}"];
       ^
testdata/invalid_wrapper_type_conversion.proto:26:3: cannot convert message to "bool"
26:    bool bool_value = 7 [(grpc.federation.field).by = "google.protobuf.BoolValue{value: true}"];
       ^
testdata/invalid_wrapper_type_conversion.proto:27:3: cannot convert message to "string"
27:    string string_value = 8 [(grpc.federation.field).by = "google.protobuf.StringValue{value: 'hello'}"];
       ^
testdata/invalid_wrapper_type_conversion.proto:28:3: cannot convert message to "bytes"
28:    bytes bytes_value = 9 [(grpc.federation.field).by = "google.protobuf.BytesValue{value: bytes('world')}"];
       ^
testdata/invalid_wrapper_type_conversion.proto:40:3: cannot convert type automatically: field type is "message" but specified value type is "double"
40:    google.protobuf.DoubleValue double_wrapper_value2 = 19 [(grpc.federation.field).by = "1.23"];
       ^
testdata/invalid_wrapper_type_conversion.proto:41:3: cannot convert type automatically: field type is "message" but specified value type is "double"
41:    google.protobuf.FloatValue float_wrapper_value2 = 20 [(grpc.federation.field).by = "3.45"];
       ^
testdata/invalid_wrapper_type_conversion.proto:42:3: cannot convert type automatically: field type is "message" but specified value type is "int64"
42:    google.protobuf.Int64Value i64_wrapper_value2 = 21 [(grpc.federation.field).by = "1"];
       ^
testdata/invalid_wrapper_type_conversion.proto:43:3: cannot convert type automatically: field type is "message" but specified value type is "uint64"
43:    google.protobuf.UInt64Value u64_wrapper_value2 = 22 [(grpc.federation.field).by = "uint(2)"];
       ^
testdata/invalid_wrapper_type_conversion.proto:44:3: cannot convert type automatically: field type is "message" but specified value type is "int64"
44:    google.protobuf.Int32Value i32_wrapper_value2 = 23 [(grpc.federation.field).by = "3"];
       ^
testdata/invalid_wrapper_type_conversion.proto:45:3: cannot convert type automatically: field type is "message" but specified value type is "uint64"
45:    google.protobuf.UInt32Value u32_wrapper_value2 = 24 [(grpc.federation.field).by = "uint(4)"];
       ^
testdata/invalid_wrapper_type_conversion.proto:46:3: cannot convert type automatically: field type is "message" but specified value type is "bool"
46:    google.protobuf.BoolValue bool_wrapper_value2 = 25 [(grpc.federation.field).by = "true"];
       ^
testdata/invalid_wrapper_type_conversion.proto:47:3: cannot convert type automatically: field type is "message" but specified value type is "string"
47:    google.protobuf.StringValue string_wrapper_value2 = 26 [(grpc.federation.field).by = "'hello'"];
       ^
testdata/invalid_wrapper_type_conversion.proto:48:3: cannot convert type automatically: field type is "message" but specified value type is "bytes"
48:    google.protobuf.BytesValue bytes_wrapper_value2 = 27 [(grpc.federation.field).by = "bytes('world')"];
       ^
`},
		{file: "invalid_validation_return_type.proto", expected: `
testdata/invalid_validation_return_type.proto:48:17: if must always return a boolean value
48:              if: "post.id"
                     ^
`},
		{file: "invalid_validation_details_return_type.proto", expected: `
testdata/invalid_validation_details_return_type.proto:49:19: if must always return a boolean value
49:                if: "'string'"
                       ^
`},
		{file: "invalid_validation_message_argument.proto", expected: `
testdata/invalid_validation_message_argument.proto:69:52: ERROR: <input>:1:8: undefined field 'message'
 | __ARG__.message
 | .......^
69:    string message = 1 [(grpc.federation.field).by = "$.message"];
                                                        ^
`},
		{file: "invalid_validation_precondition_failure.proto", expected: `
testdata/invalid_validation_precondition_failure.proto:52:25: type must always return a string value
52:                    type: "1",
                             ^
testdata/invalid_validation_precondition_failure.proto:53:28: subject must always return a string value
53:                    subject: "2",
                                ^
testdata/invalid_validation_precondition_failure.proto:54:32: description must always return a string value
54:                    description: "3",
                                    ^
`},
		{file: "invalid_validation_bad_request.proto", expected: `
testdata/invalid_validation_bad_request.proto:52:26: field must always return a string value
52:                    field: "1",
                              ^
testdata/invalid_validation_bad_request.proto:53:32: description must always return a string value
53:                    description: "2",
                                    ^
`},
		{file: "invalid_validation_localized_message.proto", expected: `
testdata/invalid_validation_localized_message.proto:52:26: message must always return a string value
52:                  message: "1"
                              ^
`},
		{file: "invalid_list_sort.proto", expected: `
testdata/invalid_list_sort.proto:55:59: ERROR: <input>:1:14: list(org.federation.User) is not comparable
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
testdata/invalid_message_map.proto:30:3: cannot convert type automatically: map key type is "int32" but specified map key type is "string"
30:    map<int32, int32> map_value = 1 [(grpc.federation.field).by = "map_value"];
       ^
testdata/invalid_message_map.proto:30:3: cannot convert type automatically: map value type is "int32" but specified map value type is "string"
30:    map<int32, int32> map_value = 1 [(grpc.federation.field).by = "map_value"];
       ^
testdata/invalid_message_map.proto:40:19: cannot convert type automatically: map key type is "string" but specified map key type is "int64"
40:            request { field: "ids", by: "$.ids" }
                       ^
testdata/invalid_message_map.proto:40:19: cannot convert type automatically: map value type is "string" but specified map value type is "int64"
40:            request { field: "ids", by: "$.ids" }
                       ^
`},
		{file: "invalid_message_map_alias.proto", expected: `
testdata/invalid_message_map_alias.proto:37:3: cannot convert type automatically: map key type is "string" but specified map key type is "int32"
37:    map<string, string> counts = 4;
       ^
testdata/invalid_message_map_alias.proto:37:3: cannot convert type automatically: map value type is "string" but specified map value type is "int32"
37:    map<string, string> counts = 4;
       ^
`},
		{file: "invalid_file_import.proto", expected: `
testdata/invalid_file_import.proto:10:12: unknown.proto: no such file or directory
10:    import: ["unknown.proto"]
                ^
testdata/invalid_file_import.proto:40:17: "post" package does not exist
40:          method: "post.PostService/GetPost"
                     ^
testdata/invalid_file_import.proto:44:29: ERROR: <input>:1:1: undeclared reference to 'res' (in container 'federation')
 | res.post
 | ^
44:      def { name: "post", by: "res.post", autobind: true  }
                                 ^
testdata/invalid_file_import.proto:46:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
46:    string id = 1;
       ^
testdata/invalid_file_import.proto:47:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
47:    string title = 2;
       ^
testdata/invalid_file_import.proto:48:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
48:    string content = 3;
       ^
testdata/invalid_file_import.proto:49:3: "user_id" field in "federation.Post" message needs to specify "grpc.federation.field" option
49:    string user_id = 4;
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
			got := validator.Format(v.Validate(ctx, file))
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
