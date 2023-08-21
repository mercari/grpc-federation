package validator_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

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
testdata/invalid_autobind.proto:23:3: "id" field found multiple times in the message specified by autobind. since it is not possible to determine one, please use "grpc.federation.field" to explicitly bind it. found message names are "a" name at messages and "b" name at messages
23:    string id = 1;
       ^
testdata/invalid_autobind.proto:23:3: "id" field in "org.federation.GetResponse" message needs to specify "grpc.federation.field" option
23:    string id = 1;
       ^
`},
		{file: "invalid_field_option.proto", expected: `
testdata/invalid_field_option.proto:30:50: "invalid" field does not exist in "Post" message
30:    Post post = 1 [(grpc.federation.field) = { by: "post.invalid" }];
                                                      ^
`},
		{file: "invalid_go_package.proto", expected: `
testdata/invalid_go_package.proto:9:21: go_package option "a;b;c;d" is invalid
9:  option go_package = "a;b;c;d";
                        ^
`},
		{file: "invalid_enum_alias_target.proto", expected: `
testdata/invalid_enum_alias_target.proto:47:41: required specify alias = "org.post.PostDataType" in grpc.federation.enum option for the "org.federation.PostType" type to automatically assign a value to the "PostData.type" field via autobind
47:    option (grpc.federation.enum).alias = "org.post.FakePostDataType";
                                             ^
testdata/invalid_enum_alias_target.proto:66:43: required specify alias = "org.post.PostContent.Category" in grpc.federation.enum option for the "org.federation.PostContent.Category" type to automatically assign a value to the "PostContent.category" field via autobind
66:      option (grpc.federation.enum).alias = "org.post.FakePostContent.FakeCategory";
                                               ^
`},
		{file: "invalid_message_alias_target.proto", expected: `
testdata/invalid_message_alias_target.proto:55:44: required specify alias = "org.post.PostData" in grpc.federation.message option for the "org.federation.PostData" type to automatically assign a value to the "Post.data" field via autobind
55:    option (grpc.federation.message).alias = "org.post.FakePostData";
                                                ^
`},
		{file: "invalid_message_alias.proto", expected: `
testdata/invalid_message_alias.proto:53:44: cannot find package from "invalid.Invalid"
53:    option (grpc.federation.message).alias = "invalid.Invalid";
                                                ^
testdata/invalid_message_alias.proto:53:3: root message does not exist in message rule dependency graph
53:    option (grpc.federation.message).alias = "invalid.Invalid";
       ^
testdata/invalid_message_alias.proto:43:3: required specify alias = "org.post.PostData" in grpc.federation.message option for the "org.federation.PostData" type to automatically assign a value to the "Post.data" field via autobind
43:    PostData data = 4;
       ^
testdata/invalid_message_alias.proto:55:3: "type" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
55:    PostType type = 1;
       ^
testdata/invalid_message_alias.proto:56:3: "title" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
56:    string title = 2;
       ^
testdata/invalid_message_alias.proto:57:3: "content" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
57:    PostContent content = 3;
       ^
`},
		{file: "invalid_method.proto", expected: `
testdata/invalid_method.proto:36:15: invalid method format. required format is "<package-name>.<service-name>/<method-name>" but specified ""
36:        method: ""
                   ^
testdata/invalid_method.proto:39:57: "invalid" name reference does not exist in this grpc.federation.message option
39:        { name: "user", message: "User", args: [{ inline: "invalid" }]}
                                                             ^
testdata/invalid_method.proto:42:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
42:    string id = 1;
       ^
testdata/invalid_method.proto:43:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
43:    string title = 2;
       ^
testdata/invalid_method.proto:44:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
44:    string content = 3;
       ^
`},
		{file: "invalid_retry.proto", expected: `
testdata/invalid_retry.proto:41:21: time: missing unit in duration "1"
41:            interval: "1"
                         ^
testdata/invalid_retry.proto:55:29: time: missing unit in duration "2"
55:            initial_interval: "2"
                                 ^
testdata/invalid_retry.proto:69:25: time: missing unit in duration "3"
69:            max_interval: "3"
                             ^
`},
		{file: "invalid_method_service_name.proto", expected: `
testdata/invalid_method_service_name.proto: "post.InvalidService" service does not exist
testdata/invalid_method_service_name.proto:36:15: cannot find "method" method because the service to which the method belongs does not exist
36:        method: "post.InvalidService/method"
                   ^
testdata/invalid_method_service_name.proto:42:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
42:    string id = 1;
       ^
testdata/invalid_method_service_name.proto:43:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
43:    string title = 2;
       ^
testdata/invalid_method_service_name.proto:44:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
44:    string content = 3;
       ^
`},
		{file: "invalid_method_name.proto", expected: `
testdata/invalid_method_name.proto:14:7: [WARN] "post.PostService" defined in "dependencies" of "grpc.federation.service" but it is not used
14:        { name: "post_service", service: "post.PostService" },
           ^
testdata/invalid_method_name.proto:37:15: "invalid" method does not exist in PostService service
37:        method: "post.PostService/invalid"
                   ^
testdata/invalid_method_name.proto:43:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
43:    string id = 1;
       ^
testdata/invalid_method_name.proto:44:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
44:    string title = 2;
       ^
testdata/invalid_method_name.proto:45:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
45:    string content = 3;
       ^
`},
		{file: "invalid_method_timeout_format.proto", expected: `
testdata/invalid_method_timeout_format.proto:12:47: time: unknown unit "p" in duration "1p"
12:      option (grpc.federation.method).timeout = "1p";
                                                   ^
`},

		{file: "invalid_method_request.proto", expected: `
testdata/invalid_method_request.proto:39:18: "invalid" field does not exist in "post.GetPostRequest" message for method request
39:          { field: "invalid", by: "$.invalid" }
                      ^
testdata/invalid_method_request.proto:39:33: "invalid" field does not exist in "PostArgument" message
39:          { field: "invalid", by: "$.invalid" }
                                     ^
`},
		{file: "missing_field_option.proto", expected: `
testdata/missing_field_option.proto:50:3: "user" field in "federation.Post" message needs to specify "grpc.federation.field" option
50:    User user = 4;
       ^
`},
		{file: "missing_message_alias.proto", expected: `
testdata/missing_message_alias.proto:55:3: use "alias" in "grpc.federation.field" option, but "alias" is not defined in "grpc.federation.message" option
55:    PostContent content = 3 [(grpc.federation.field).alias = "content"];
       ^
testdata/missing_message_alias.proto:43:3: required specify alias = "org.post.PostData" in grpc.federation.message option for the "org.federation.PostData" type to automatically assign a value to the "Post.data" field via autobind
43:    PostData data = 4;
       ^
`},
		{file: "missing_enum_alias.proto", expected: `
testdata/missing_enum_alias.proto:48:62: use "alias" in "grpc.federation.enum_value" option, but "alias" is not defined in "grpc.federation.enum" option
48:    POST_TYPE_FOO = 1 [(grpc.federation.enum_value) = { alias: ["POST_TYPE_A"] }];
                                                                  ^
testdata/missing_enum_alias.proto:49:62: use "alias" in "grpc.federation.enum_value" option, but "alias" is not defined in "grpc.federation.enum" option
49:    POST_TYPE_BAR = 2 [(grpc.federation.enum_value) = { alias: ["POST_TYPE_B", "POST_TYPE_C"] }];
                                                                  ^
testdata/missing_enum_alias.proto:46:1: required specify alias = "org.post.PostDataType" in grpc.federation.enum option for the "org.federation.PostType" type to automatically assign a value to the "PostData.type" field via autobind
46:  enum PostType {
     ^
testdata/missing_enum_alias.proto:63:3: required specify alias = "org.post.PostContent.Category" in grpc.federation.enum option for the "org.federation.PostContent.Category" type to automatically assign a value to the "PostContent.category" field via autobind
63:    enum Category {
       ^
`},
		{file: "missing_enum_value.proto", expected: `
testdata/missing_enum_value.proto:66:5: specified "alias" in grpc.federation.enum option, but "CATEGORY_C" value does not exist in "org.post.PostContent.Category" enum
66:      CATEGORY_C = 2;
         ^
testdata/missing_enum_value.proto:49:3: specified "alias" in grpc.federation.enum option, but "FOO" value does not exist in "org.post.PostDataType" enum
49:    FOO = 0;
       ^
`},
		{file: "missing_enum_value_alias.proto", expected: `
testdata/missing_enum_value_alias.proto:49:3: specified "alias" in grpc.federation.enum option, but "POST_TYPE_UNKNOWN" value does not exist in "org.post.PostDataType" enum
49:    POST_TYPE_UNKNOWN = 0;
       ^
testdata/missing_enum_value_alias.proto:50:3: specified "alias" in grpc.federation.enum option, but "POST_TYPE_FOO" value does not exist in "org.post.PostDataType" enum
50:    POST_TYPE_FOO = 1;
       ^
testdata/missing_enum_value_alias.proto:51:3: specified "alias" in grpc.federation.enum option, but "POST_TYPE_BAR" value does not exist in "org.post.PostDataType" enum
51:    POST_TYPE_BAR = 2;
       ^
`},
		{file: "missing_message_field_alias.proto", expected: `
testdata/missing_message_field_alias.proto:75:3: specified "alias" in grpc.federation.message option, but "dup_body" field does not exist in "org.post.PostContent" message
75:    string dup_body = 4;
       ^
testdata/missing_message_field_alias.proto:75:3: "dup_body" field in "org.federation.PostContent" message needs to specify "grpc.federation.field" option
75:    string dup_body = 4;
       ^
`},
		{file: "missing_message_option.proto", expected: `
testdata/missing_message_option.proto:43:32: "federation.User" message does not specify "grpc.federation.message" option
43:        { name: "user", message: "User", args: [{ inline: "post" }]}
                                    ^
`},
		{file: "missing_method_request_value.proto", expected: `
testdata/missing_method_request_value.proto:39:9: "by" or literal value must be specified
39:          { field: "id" }
             ^
`},
		{file: "missing_response_message_option.proto", expected: `
testdata/missing_response_message_option.proto:18:1: "federation.GetPostResponse" message needs to specify "grpc.federation.message" option
18:  message GetPostResponse {
     ^
`},
		{file: "invalid_method_response.proto", expected: `
testdata/invalid_method_response.proto:41:42: "invalid" field does not exist in "post.GetPostResponse" message for method response
41:        response: [ { name: "post", field: "invalid", autobind: true  } ]
                                              ^
testdata/invalid_method_response.proto:44:57: "post" name reference does not exist in this grpc.federation.message option
44:        { name: "user", message: "User", args: [{ inline: "post" }]}
                                                             ^
testdata/invalid_method_response.proto:47:3: "id" field in "federation.Post" message needs to specify "grpc.federation.field" option
47:    string id = 1;
       ^
testdata/invalid_method_response.proto:48:3: "title" field in "federation.Post" message needs to specify "grpc.federation.field" option
48:    string title = 2;
       ^
testdata/invalid_method_response.proto:49:3: "content" field in "federation.Post" message needs to specify "grpc.federation.field" option
49:    string content = 3;
       ^
`},
		{file: "invalid_message_name.proto", expected: `
testdata/invalid_message_name.proto:15:7: [WARN] "user.UserService" defined in "dependencies" of "grpc.federation.service" but it is not used
15:        { name: "user_service", service: "user.UserService" }
           ^
testdata/invalid_message_name.proto: "federation.Invalid" message does not exist
testdata/invalid_message_name.proto:44:32: undefined message specified "grpc.federation.message" option
44:        { name: "user", message: "Invalid", args: [{ inline: "post" }]}
                                    ^
`},
		{file: "invalid_message_argument.proto", expected: `
testdata/invalid_message_argument.proto:44:100: invalid path format. "." cannot be used after dot character
44:        { name: "user", message: "User", args: [{ by: "$.id.invalid" }, { inline: "post.id" }, { by: "...." }, { inline: "...." }]}
                                                                                                        ^
testdata/invalid_message_argument.proto:44:120: invalid path format. "." cannot be used after dot character
44:        { name: "user", message: "User", args: [{ by: "$.id.invalid" }, { inline: "post.id" }, { by: "...." }, { inline: "...." }]}
                                                                                                                            ^
testdata/invalid_message_argument.proto:44:81: inline keyword must refer to a message type. but "post" is not a message type
44:        { name: "user", message: "User", args: [{ by: "$.id.invalid" }, { inline: "post.id" }, { by: "...." }, { inline: "...." }]}
                                                                                     ^
testdata/invalid_message_argument.proto:44:53: "invalid" path selector must refer to a message type. but it is not a message type
44:        { name: "user", message: "User", args: [{ by: "$.id.invalid" }, { inline: "post.id" }, { by: "...." }, { inline: "...." }]}
                                                         ^
`},
		{file: "invalid_message_field_alias.proto", expected: `
testdata/invalid_message_field_alias.proto:58:3: The types of "org.federation.PostData"'s "title" field ("int64") and "org.post.PostData"'s field ("string") are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself
58:    int64 title = 2;
       ^
testdata/invalid_message_field_alias.proto:74:3: The types of "org.federation.PostContent"'s "body" field ("int64") and "org.post.PostContent"'s field ("string") are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself
74:    int64 body = 3;
       ^
testdata/invalid_message_field_alias.proto:58:3: "title" field in "org.federation.PostData" message needs to specify "grpc.federation.field" option
58:    int64 title = 2;
       ^
testdata/invalid_message_field_alias.proto:74:3: "body" field in "org.federation.PostContent" message needs to specify "grpc.federation.field" option
74:    int64 body = 3;
       ^
`},
		{file: "duplicate_service_dependency_name.proto", expected: `
testdata/duplicate_service_dependency_name.proto:15:17: "%s" name duplicated
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
testdata/invalid_service_dependency_service.proto:14:42: "post.InvalidService" does not exist
14:          { name: "post_service", service: "post.InvalidService" },
                                              ^
testdata/invalid_service_dependency_service.proto: "user.InvalidService" service does not exist
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
testdata/recursive_message_name.proto:43:32: recursive definition: "Post" is own message name
43:        { name: "self", message: "Post" }
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
			actual := "\n" + validator.Format(v.Validate(ctx, file))
			if test.expected != actual {
				t.Fatalf("expected error %s\n but got %s", test.expected, actual)
			}
		})
	}
}
