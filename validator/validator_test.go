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
	tests := []struct {
		file     string
		expected string
	}{
		{file: "empty_response_field.proto", expected: `
`},
		{file: "invalid_autobind.proto", expected: `
testdata/invalid_autobind.proto:23:3: "id" field found multiple times in the message specified by autobind. since it is not possible to determine one, please use "grpc.federation.field" to explicitly bind it. found message names are "a" name at def and "b" name at def
23:    string id = 1;
       ^
testdata/invalid_autobind.proto:23:3: "id" field in "org.federation.GetResponse" message needs to specify "grpc.federation.field" option
23:    string id = 1;
       ^
`},
		{file: "invalid_condition_type.proto", expected: `
testdata/invalid_condition_type.proto:36:13: return value of "if" must be bool type but got string type
36:          if: "$.id"
                 ^
`},
		{file: "invalid_field_option.proto", expected: `
testdata/invalid_field_option.proto:34:50: ERROR: <input>:1:5: undefined field 'invalid'
 | post.invalid
 | ....^
34:    Post post = 1 [(grpc.federation.field) = { by: "post.invalid" }];
                                                      ^
`},
		{file: "invalid_go_package.proto", expected: `
testdata/invalid_go_package.proto:9:21: go_package option "a;b;c;d" is invalid
9:  option go_package = "a;b;c;d";
                        ^
`},
		{file: "invalid_enum_alias_target.proto", expected: `
testdata/invalid_enum_alias_target.proto:52:41: required specify alias = "org.post.PostDataType" in grpc.federation.enum option for the "org.federation.PostType" type to automatically assign a value to the "PostData.type" field via autobind
52:    option (grpc.federation.enum).alias = "org.post.FakePostDataType";
                                             ^
testdata/invalid_enum_alias_target.proto:71:43: required specify alias = "org.post.PostContent.Category" in grpc.federation.enum option for the "org.federation.PostContent.Category" type to automatically assign a value to the "PostContent.category" field via autobind
71:      option (grpc.federation.enum).alias = "org.post.FakePostContent.FakeCategory";
                                               ^
`},
		{file: "invalid_error_variable.proto", expected: `
testdata/invalid_error_variable.proto:20:17: "error" is the reserved keyword. this name is not available
20:      def { name: "error" by: "'foo'" }
                     ^
testdata/invalid_error_variable.proto:21:25: ERROR: <input>:1:1: undeclared reference to 'error' (in container '')
 | error
 | ^
21:      def { name: "e" by: "error" }
                             ^
testdata/invalid_error_variable.proto:25:15: ERROR: <input>:1:1: undeclared reference to 'error' (in container '')
 | error.code == 0
 | ^
25:            if: "error.code == 0"
                   ^
`},
		{file: "invalid_map_iterator_src_type.proto", expected: `
testdata/invalid_map_iterator_src_type.proto:40:13: map iterator's src value type must be repeated type
40:          map {
                 ^
testdata/invalid_map_iterator_src_type.proto:54:57: ERROR: <input>:1:1: undeclared reference to 'users' (in container '')
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
testdata/invalid_map_iterator_src.proto:36:13: ERROR: <input>:1:1: undeclared reference to 'iter' (in container '')
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
testdata/invalid_message_alias_target.proto:60:44: required specify alias = "org.post.PostData" in grpc.federation.message option for the "org.federation.PostData" type to automatically assign a value to the "Post.data" field via autobind
60:    option (grpc.federation.message).alias = "org.post.FakePostData";
                                                ^
`},
		{file: "invalid_message_alias.proto", expected: `
testdata/invalid_message_alias.proto:48:3: required specify alias = "org.post.PostData" in grpc.federation.message option for the "org.federation.PostData" type to automatically assign a value to the "Post.data" field via autobind
48:    PostData data = 4;
       ^
testdata/invalid_message_alias.proto:58:44: cannot find package from "invalid.Invalid"
58:    option (grpc.federation.message).alias = "invalid.Invalid";
                                                ^
testdata/invalid_message_alias.proto:60:3: "type" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
60:    PostType type = 1;
       ^
testdata/invalid_message_alias.proto:61:3: "title" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
61:    string title = 2;
       ^
testdata/invalid_message_alias.proto:62:3: "content" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
62:    PostContent content = 3;
       ^
`},
		{file: "invalid_method.proto", expected: `
testdata/invalid_method.proto:40:24: invalid method format. required format is "<package-name>.<service-name>/<method-name>" but specified ""
40:        { call { method: "" } },
                            ^
testdata/invalid_method.proto:45:26: ERROR: <input>:1:1: undeclared reference to 'invalid' (in container '')
 | invalid
 | ^
45:            args { inline: "invalid" }
                              ^
testdata/invalid_method.proto:50:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
50:    string id = 1;
       ^
testdata/invalid_method.proto:51:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
51:    string title = 2;
       ^
testdata/invalid_method.proto:52:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
52:    string content = 3;
       ^
testdata/invalid_method.proto:62:36: ERROR: <input>:1:8: undefined field 'user_id'
 | __ARG__.user_id
 | .......^
62:          request { field: "id", by: "$.user_id" }
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
testdata/invalid_oneof.proto:39:13: return value of "if" must be bool type but got int64 type
39:          if: "1"
                 ^
testdata/invalid_oneof.proto:52:39: "if" or "default" must be specified in "grpc.federation.field.oneof"
52:        (grpc.federation.field).oneof = {
                                           ^
testdata/invalid_oneof.proto:65:39: "by" must be specified in "grpc.federation.field.oneof"
65:        (grpc.federation.field).oneof = {
                                           ^
testdata/invalid_oneof.proto:79:18: "default" found multiple times in the "grpc.federation.field.oneof". "default" can only be specified once per oneof
79:          default: true
                      ^
testdata/invalid_oneof.proto:91:3: "oneof" feature can only be used for fields within oneof
91:    bool foo = 5 [(grpc.federation.field).oneof = {
       ^
testdata/invalid_oneof.proto:91:3: value must be specified
91:    bool foo = 5 [(grpc.federation.field).oneof = {
       ^
testdata/invalid_oneof.proto:91:3: "foo" field in "org.federation.UserSelection" message needs to specify "grpc.federation.field" option
91:    bool foo = 5 [(grpc.federation.field).oneof = {
       ^
`},
		{file: "invalid_retry.proto", expected: `
testdata/invalid_retry.proto:42:23: time: missing unit in duration "1"
42:              interval: "1"
                           ^
testdata/invalid_retry.proto:58:31: time: missing unit in duration "2"
58:              initial_interval: "2"
                                   ^
testdata/invalid_retry.proto:74:27: time: missing unit in duration "3"
74:              max_interval: "3"
                               ^
`},
		{file: "invalid_method_service_name.proto", expected: `
testdata/invalid_method_service_name.proto: "post.InvalidService" service does not exist
testdata/invalid_method_service_name.proto:40:24: cannot find "method" method because the service to which the method belongs does not exist
40:        { call { method: "post.InvalidService/method" } },
                            ^
testdata/invalid_method_service_name.proto:50:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
50:    string id = 1;
       ^
testdata/invalid_method_service_name.proto:51:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
51:    string title = 2;
       ^
testdata/invalid_method_service_name.proto:52:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
52:    string content = 3;
       ^
`},
		{file: "invalid_method_name.proto", expected: `
testdata/invalid_method_name.proto:14:7: [WARN] "post.PostService" defined in "dependencies" of "grpc.federation.service" but it is not used
14:        { name: "post_service", service: "post.PostService" },
           ^
testdata/invalid_method_name.proto:41:24: "invalid" method does not exist in PostService service
41:        { call { method: "post.PostService/invalid" } },
                            ^
testdata/invalid_method_name.proto:51:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
51:    string id = 1;
       ^
testdata/invalid_method_name.proto:52:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
52:    string title = 2;
       ^
testdata/invalid_method_name.proto:53:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
53:    string content = 3;
       ^
`},
		{file: "invalid_method_timeout_format.proto", expected: `
testdata/invalid_method_timeout_format.proto:12:47: time: unknown unit "p" in duration "1p"
12:      option (grpc.federation.method).timeout = "1p";
                                                   ^
`},

		{file: "invalid_method_request.proto", expected: `
testdata/invalid_method_request.proto:45:28: "invalid" field does not exist in "post.GetPostRequest" message for method request
45:            request { field: "invalid", by: "$.invalid" }
                                ^
testdata/invalid_method_request.proto:45:43: ERROR: <input>:1:8: undefined field 'invalid'
 | __ARG__.invalid
 | .......^
45:            request { field: "invalid", by: "$.invalid" }
                                               ^
`},
		{file: "missing_field_option.proto", expected: `
testdata/missing_field_option.proto:61:3: "user" field in "federation.Post" message needs to specify "grpc.federation.field" option
61:    User user = 4;
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
testdata/missing_message_alias.proto:48:3: required specify alias = "org.post.PostData" in grpc.federation.message option for the "org.federation.PostData" type to automatically assign a value to the "Post.data" field via autobind
48:    PostData data = 4;
       ^
testdata/missing_message_alias.proto:58:3: "type" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
58:    PostType type = 1;
       ^
testdata/missing_message_alias.proto:59:3: "title" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
59:    string title = 2;
       ^
testdata/missing_message_alias.proto:60:3: use "alias" in "grpc.federation.field" option, but "alias" is not defined in "grpc.federation.message" option
60:    PostContent content = 3 [(grpc.federation.field).alias = "content"];
       ^
`},
		{file: "missing_enum_alias.proto", expected: `
testdata/missing_enum_alias.proto:51:1: required specify alias = "org.post.PostDataType" in grpc.federation.enum option for the "org.federation.PostType" type to automatically assign a value to the "PostData.type" field via autobind
51:  enum PostType {
     ^
testdata/missing_enum_alias.proto:53:62: use "alias" in "grpc.federation.enum_value" option, but "alias" is not defined in "grpc.federation.enum" option
53:    POST_TYPE_FOO = 1 [(grpc.federation.enum_value) = { alias: ["POST_TYPE_A"] }];
                                                                  ^
testdata/missing_enum_alias.proto:54:62: use "alias" in "grpc.federation.enum_value" option, but "alias" is not defined in "grpc.federation.enum" option
54:    POST_TYPE_BAR = 2 [(grpc.federation.enum_value) = { alias: ["POST_TYPE_B", "POST_TYPE_C"] }];
                                                                  ^
testdata/missing_enum_alias.proto:68:3: required specify alias = "org.post.PostContent.Category" in grpc.federation.enum option for the "org.federation.PostContent.Category" type to automatically assign a value to the "PostContent.category" field via autobind
68:    enum Category {
       ^
`},
		{file: "missing_enum_value.proto", expected: `
testdata/missing_enum_value.proto:54:3: specified "alias" in grpc.federation.enum option, but "FOO" value does not exist in "org.post.PostDataType" enum
54:    FOO = 0;
       ^
testdata/missing_enum_value.proto:71:5: specified "alias" in grpc.federation.enum option, but "CATEGORY_C" value does not exist in "org.post.PostContent.Category" enum
71:      CATEGORY_C = 2;
         ^
`},
		{file: "missing_enum_value_alias.proto", expected: `
testdata/missing_enum_value_alias.proto:54:3: specified "alias" in grpc.federation.enum option, but "POST_TYPE_UNKNOWN" value does not exist in "org.post.PostDataType" enum
54:    POST_TYPE_UNKNOWN = 0;
       ^
testdata/missing_enum_value_alias.proto:55:3: specified "alias" in grpc.federation.enum option, but "POST_TYPE_FOO" value does not exist in "org.post.PostDataType" enum
55:    POST_TYPE_FOO = 1;
       ^
testdata/missing_enum_value_alias.proto:56:3: specified "alias" in grpc.federation.enum option, but "POST_TYPE_BAR" value does not exist in "org.post.PostDataType" enum
56:    POST_TYPE_BAR = 2;
       ^
`},
		{file: "valid_enum_value_reference.proto", expected: ""},
		{file: "missing_message_field_alias.proto", expected: `
testdata/missing_message_field_alias.proto:80:3: specified "alias" in grpc.federation.message option, but "dup_body" field does not exist in "org.post.PostContent" message
80:    string dup_body = 4;
       ^
testdata/missing_message_field_alias.proto:80:3: "dup_body" field in "org.federation.PostContent" message needs to specify "grpc.federation.field" option
80:    string dup_body = 4;
       ^
`},
		{file: "missing_message_option.proto", expected: `
testdata/missing_message_option.proto:51:17: "federation.User" message does not specify "grpc.federation.message" option
51:            name: "User"
                     ^
`},
		{file: "missing_method_request_value.proto", expected: `
testdata/missing_method_request_value.proto:45:19: value must be specified
45:            request { field: "id" }
                       ^
`},
		{file: "missing_response_message_option.proto", expected: `
testdata/missing_response_message_option.proto:18:1: "federation.GetPostResponse" message needs to specify "grpc.federation.message" option
18:  message GetPostResponse {
     ^
`},
		{file: "invalid_method_response.proto", expected: `
testdata/invalid_method_response.proto:47:27: ERROR: <input>:1:1: undeclared reference to 'invalid' (in container '')
 | invalid
 | ^
47:        { name: "post", by: "invalid", autobind: true  },
                               ^
testdata/invalid_method_response.proto:52:26: ERROR: <input>:1:1: undeclared reference to 'post' (in container '')
 | post
 | ^
52:            args { inline: "post" }
                              ^
testdata/invalid_method_response.proto:57:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
57:    string id = 1;
       ^
testdata/invalid_method_response.proto:58:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
58:    string title = 2;
       ^
testdata/invalid_method_response.proto:59:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
59:    string content = 3;
       ^
testdata/invalid_method_response.proto:69:36: ERROR: <input>:1:8: undefined field 'user_id'
 | __ARG__.user_id
 | .......^
69:          request { field: "id", by: "$.user_id" }
                                        ^
`},
		{file: "invalid_message_name.proto", expected: `
testdata/invalid_message_name.proto: "federation.Invalid" message does not exist
testdata/invalid_message_name.proto:15:7: [WARN] "user.UserService" defined in "dependencies" of "grpc.federation.service" but it is not used
15:        { name: "user_service", service: "user.UserService" }
           ^
testdata/invalid_message_name.proto:52:17: undefined message specified
52:            name: "Invalid"
                     ^
testdata/invalid_message_name.proto:61:47: unknown type null_type is required
61:    User user = 4 [(grpc.federation.field).by = "user"];
                                                   ^
testdata/invalid_message_name.proto:70:36: ERROR: <input>:1:8: undefined field 'user_id'
 | __ARG__.user_id
 | .......^
70:          request { field: "id", by: "$.user_id" }
                                        ^
`},
		{file: "invalid_message_argument.proto", expected: `
testdata/invalid_message_argument.proto:54:19: ERROR: <input>:1:11: type 'string' does not support field selection
 | __ARG__.id.invalid
 | ..........^
54:              { by: "$.id.invalid" },
                       ^
testdata/invalid_message_argument.proto:55:23: inline value is not message type
55:              { inline: "post.id" },
                           ^
testdata/invalid_message_argument.proto:56:19: ERROR: <input>:1:2: Syntax error: no viable alternative at input '..'
 | ....
 | .^
56:              { by: "...." },
                       ^
testdata/invalid_message_argument.proto:57:23: ERROR: <input>:1:2: Syntax error: no viable alternative at input '..'
 | ....
 | .^
57:              { inline: "...." }
                           ^
testdata/invalid_message_argument.proto:75:36: ERROR: <input>:1:8: undefined field 'user_id'
 | __ARG__.user_id
 | .......^
75:          request { field: "id", by: "$.user_id" }
                                        ^
`},
		{file: "invalid_message_field_alias.proto", expected: `
testdata/invalid_message_field_alias.proto:63:3: The types of "org.federation.PostData"'s "title" field ("int64") and "org.post.PostData"'s field ("string") are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself
63:    int64 title = 2;
       ^
testdata/invalid_message_field_alias.proto:63:3: "title" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
63:    int64 title = 2;
       ^
testdata/invalid_message_field_alias.proto:79:3: The types of "org.federation.PostContent"'s "body" field ("int64") and "org.post.PostContent"'s field ("string") are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself
79:    int64 body = 3;
       ^
testdata/invalid_message_field_alias.proto:79:3: "body" field in "org.federation.PostContent" message needs to specify "grpc.federation.field" option
79:    int64 body = 3;
       ^
`},
		{file: "duplicate_service_dependency_name.proto", expected: `
testdata/duplicate_service_dependency_name.proto:15:17: "dup_name" name duplicated
15:          { name: "dup_name", service: "user.UserService" }
                     ^
`},
		{file: "invalid_service_dependency_package.proto", expected: `
testdata/invalid_service_dependency_package.proto:14:42: cannot find package from "invalid.PostService"
14:          { name: "post_service", service: "invalid.PostService" },
                                              ^
testdata/invalid_service_dependency_package.proto:15:42: cannot find package from "invalid.UserService"
15:          { name: "user_service", service: "invalid.UserService" }
                                              ^
`},
		{file: "invalid_service_dependency_service.proto", expected: `
testdata/invalid_service_dependency_service.proto: "post.InvalidService" service does not exist
testdata/invalid_service_dependency_service.proto: "user.InvalidService" service does not exist
testdata/invalid_service_dependency_service.proto:14:42: "post.InvalidService" does not exist
14:          { name: "post_service", service: "post.InvalidService" },
                                              ^
testdata/invalid_service_dependency_service.proto:15:42: "user.InvalidService" does not exist
15:          { name: "user_service", service: "user.InvalidService" }
                                              ^
`},
		{file: "missing_service_dependency_service.proto", expected: `
testdata/missing_service_dependency_service.proto:14:9: "service" must be specified
14:          { name: "post_service" },
             ^
testdata/missing_service_dependency_service.proto:15:9: "service" must be specified
15:          { name: "user_service" }
             ^
`},
		{file: "recursive_message_name.proto", expected: `
testdata/recursive_message_name.proto:37:1: found cyclic dependency in "federation.Post" message. dependency path: GetPostResponse => Post => Post
37:  message Post {
     ^
testdata/recursive_message_name.proto:48:39: recursive definition: "Post" is own message name
48:        { name: "self", message { name: "Post" } }
                                           ^
testdata/recursive_message_name.proto:51:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
51:    string id = 1;
       ^
testdata/recursive_message_name.proto:52:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
52:    string title = 2;
       ^
testdata/recursive_message_name.proto:53:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
53:    string content = 3;
       ^
`},
		{file: "message_cyclic_dependency.proto", expected: `
testdata/message_cyclic_dependency.proto:27:1: found cyclic dependency in "org.federation.A" message. dependency path: GetResponse => A => AA => AAA => A
27:  message A {
     ^
`},
		{file: "invalid_validation_return_type.proto", expected: `
testdata/invalid_validation_return_type.proto:50:17: if must always return a boolean value
50:              if: "post.id"
                     ^
`},
		{file: "invalid_validation_details_return_type.proto", expected: `
testdata/invalid_validation_details_return_type.proto:51:19: if must always return a boolean value
51:                if: "'string'"
                       ^
`},
		{file: "invalid_validation_message_argument.proto", expected: `
testdata/invalid_validation_message_argument.proto:71:52: ERROR: <input>:1:8: undefined field 'message'
 | __ARG__.message
 | .......^
71:    string message = 1 [(grpc.federation.field).by = "$.message"];
                                                        ^
`},
		{file: "invalid_validation_precondition_failure.proto", expected: `
testdata/invalid_validation_precondition_failure.proto:54:25: type must always return a string value
54:                    type: "1",
                             ^
testdata/invalid_validation_precondition_failure.proto:55:28: subject must always return a string value
55:                    subject: "2",
                                ^
testdata/invalid_validation_precondition_failure.proto:56:32: description must always return a string value
56:                    description: "3",
                                    ^
`},
		{file: "invalid_validation_bad_request.proto", expected: `
testdata/invalid_validation_bad_request.proto:54:26: field must always return a string value
54:                    field: "1",
                              ^
testdata/invalid_validation_bad_request.proto:55:32: description must always return a string value
55:                    description: "2",
                                    ^
`},
		{file: "invalid_validation_localized_message.proto", expected: `
testdata/invalid_validation_localized_message.proto:54:26: message must always return a string value
54:                  message: "1"
                              ^
`},
	}
	ctx := context.Background()
	v := validator.New()
	for _, test := range tests {
		t.Run(test.file, func(t *testing.T) {
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
