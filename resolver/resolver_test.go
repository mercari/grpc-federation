package resolver_test

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mercari/grpc-federation/internal/testutil"
	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/source"
)

func TestSimpleAggregation(t *testing.T) {
	fb := testutil.NewFileBuilder("simple_aggregation.proto")
	ref := testutil.NewBuilderReferenceManager(getUserProtoBuilder(t), getPostProtoBuilder(t), fb)

	fb.SetPackage("org.federation").
		SetGoPackage("example/federation", "federation").
		AddMessage(
			testutil.NewMessageBuilder("ZArgument").
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("Z").
				AddFieldWithRule(
					"foo",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(nil).SetMessageCustomResolver(true).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "ZArgument")).
						SetCustomResolver(true).
						SetDependencyGraph(testutil.NewDependencyGraphBuilder().Build(t)).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("MArgument").
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("M").
				AddFieldWithRule(
					"foo",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(resolver.NewStringValue("foo")).Build(t),
				).
				AddFieldWithRule(
					"bar",
					resolver.Int64Type,
					testutil.NewFieldRuleBuilder(resolver.NewInt64Value(1)).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "MArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("PostArgument").
				AddField("id", resolver.StringType).
				AddField("double_value", resolver.DoubleType).
				AddField("float_value", resolver.FloatType).
				AddField("int32_value", resolver.Int32Type).
				AddField("int64_value", resolver.Int64Type).
				AddField("uint32_value", resolver.Uint32Type).
				AddField("uint64_value", resolver.Uint64Type).
				AddField("sint32_value", resolver.Sint32Type).
				AddField("sint64_value", resolver.Sint64Type).
				AddField("fixed32_value", resolver.Fixed32Type).
				AddField("fixed64_value", resolver.Fixed64Type).
				AddField("sfixed32_value", resolver.Sfixed32Type).
				AddField("sfixed64_value", resolver.Sfixed64Type).
				AddField("bool_value", resolver.BoolType).
				AddField("string_value", resolver.StringType).
				AddField("bytes_value", resolver.BytesType).
				AddField("message_value", ref.Type(t, "org.post", "Post")).
				AddField("double_list_value", resolver.DoubleRepeatedType).
				AddField("float_list_value", resolver.FloatRepeatedType).
				AddField("int32_list_value", resolver.Int32RepeatedType).
				AddField("int64_list_value", resolver.Int64RepeatedType).
				AddField("uint32_list_value", resolver.Uint32RepeatedType).
				AddField("uint64_list_value", resolver.Uint64RepeatedType).
				AddField("sint32_list_value", resolver.Sint32RepeatedType).
				AddField("sint64_list_value", resolver.Sint64RepeatedType).
				AddField("fixed32_list_value", resolver.Fixed32RepeatedType).
				AddField("fixed64_list_value", resolver.Fixed64RepeatedType).
				AddField("sfixed32_list_value", resolver.Sfixed32RepeatedType).
				AddField("sfixed64_list_value", resolver.Sfixed64RepeatedType).
				AddField("bool_list_value", resolver.BoolRepeatedType).
				AddField("string_list_value", resolver.StringRepeatedType).
				AddField("bytes_list_value", resolver.BytesRepeatedType).
				AddField("message_list_value", ref.RepeatedType(t, "org.post", "Post")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("UserArgument").
				AddField("id", resolver.StringType).
				AddField("title", resolver.StringType).
				AddField("content", resolver.StringType).
				AddField("user_id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("Item").
				AddEnum(
					testutil.NewEnumBuilder("ItemType").
						WithAlias(ref.Enum(t, "org.user", "Item.ItemType")).
						AddValueWithAlias("ITEM_TYPE_1", ref.EnumValue(t, "org.user", "Item.ItemType", "ITEM_TYPE_1")).
						AddValueWithAlias("ITEM_TYPE_2", ref.EnumValue(t, "org.user", "Item.ItemType", "ITEM_TYPE_2")).
						AddValueWithAlias("ITEM_TYPE_3", ref.EnumValue(t, "org.user", "Item.ItemType", "ITEM_TYPE_3")).
						Build(t),
				).
				AddFieldWithAlias("name", resolver.StringType, ref.Field(t, "org.user", "Item", "name")).
				AddFieldWithTypeNameAndAlias(t, "type", "ItemType", false, ref.Field(t, "org.user", "Item", "type")).
				AddFieldWithAlias("value", resolver.Int64Type, ref.Field(t, "org.user", "Item", "value")).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetAlias(ref.Message(t, "org.user", "Item")).
						SetDependencyGraph(testutil.NewDependencyGraphBuilder().Build(t)).
						Build(t),
				).
				Build(t),
		).
		AddEnum(
			testutil.NewEnumBuilder("UserType").
				WithAlias(ref.Enum(t, "org.user", "UserType")).
				AddValueWithAlias("USER_TYPE_1", ref.EnumValue(t, "org.user", "UserType", "USER_TYPE_1")).
				AddValueWithAlias("USER_TYPE_2", ref.EnumValue(t, "org.user", "UserType", "USER_TYPE_2")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("User").
				AddMessage(
					testutil.NewMessageBuilder("ProfileEntry").
						SetIsMapEntry(true).
						AddField("key", resolver.StringType).
						AddField("value", resolver.AnyType).
						Build(t),
				).
				AddFieldWithAutoBind("id", resolver.StringType, ref.Field(t, "org.user", "User", "id")).
				AddFieldWithAutoBind("type", ref.Type(t, "org.federation", "UserType"), ref.Field(t, "org.user", "User", "type")).
				AddFieldWithAutoBind("name", resolver.StringType, ref.Field(t, "org.user", "User", "name")).
				AddFieldWithRule("age", resolver.Uint64Type, testutil.NewFieldRuleBuilder(nil).SetCustomResolver(true).Build(t)).
				AddFieldWithAutoBind("desc", resolver.StringRepeatedType, ref.Field(t, "org.user", "User", "desc")).
				AddFieldWithAutoBind("main_item", ref.Type(t, "org.federation", "Item"), ref.Field(t, "org.user", "User", "main_item")).
				AddFieldWithAutoBind("items", ref.RepeatedType(t, "org.federation", "Item"), ref.Field(t, "org.user", "User", "items")).
				AddFieldWithTypeNameAndAutoBind(t, "profile", "ProfileEntry", true, ref.Field(t, "org.user", "User", "profile")).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMethodCall(
							testutil.NewMethodCallBuilder(ref.Method(t, "org.user", "UserService", "GetUser")).
								SetRequest(
									testutil.NewRequestBuilder().
										AddField("id", resolver.StringType, testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "user_id").Build(t)).
										Build(t),
								).
								SetResponse(
									testutil.NewResponseBuilder().AddField("user", "user", ref.Type(t, "org.user", "User"), true, true).Build(t),
								).
								SetTimeout("20s").
								SetRetryPolicyExponential(
									testutil.NewRetryPolicyExponentialBuilder().
										SetInitialInterval("1s").
										SetRandomizationFactor(0.7).
										SetMultiplier(1.7).
										SetMaxInterval("30s").
										SetMaxRetries(3).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "UserArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.user", "GetUserResponse")).
								Build(t),
						).
						AddResolver(testutil.NewMessageResolverGroupByName("GetUser")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("Post").
				AddFieldWithAutoBind("id", resolver.StringType, ref.Field(t, "org.post", "Post", "id")).
				AddFieldWithAutoBind("title", resolver.StringType, ref.Field(t, "org.post", "Post", "title")).
				AddFieldWithAutoBind("content", resolver.StringType, ref.Field(t, "org.post", "Post", "content")).
				AddFieldWithRule(
					"user",
					ref.Type(t, "org.federation", "User"),
					testutil.NewFieldRuleBuilder(
						testutil.NewNameReferenceValueBuilder(
							ref.Type(t, "org.federation", "User"),
							ref.Type(t, "org.federation", "User"),
							"user",
						).Build(t),
					).Build(t),
				).
				AddFieldWithAutoBind("foo", resolver.StringType, ref.Field(t, "org.federation", "M", "foo")).
				AddFieldWithAutoBind("bar", resolver.Int64Type, ref.Field(t, "org.federation", "M", "bar")).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMethodCall(
							testutil.NewMethodCallBuilder(ref.Method(t, "org.post", "PostService", "GetPost")).
								SetRequest(
									testutil.NewRequestBuilder().
										AddField("id", resolver.StringType, testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
										Build(t),
								).
								SetResponse(
									testutil.NewResponseBuilder().
										AddField("post", "post", ref.Type(t, "org.post", "Post"), true, true).
										Build(t),
								).
								SetTimeout("10s").
								SetRetryPolicyConstant(
									testutil.NewRetryPolicyConstantBuilder().
										SetInterval("2s").
										SetMaxRetries(3).
										Build(t),
								).
								Build(t),
						).
						AddMessageDependency(
							"user",
							ref.Message(t, "org.federation", "User"),
							testutil.NewMessageDependencyArgumentBuilder().
								Inline(testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.post", "GetPostResponse"), ref.Type(t, "org.post", "Post"), "post").Build(t)).
								Build(t),
							false,
							true,
						).
						AddMessageDependency(
							"z",
							ref.Message(t, "org.federation", "Z"),
							nil,
							false,
							false,
						).
						AddMessageDependency(
							"m",
							ref.Message(t, "org.federation", "M"),
							nil,
							true,
							true,
						).
						SetMessageArgument(ref.Message(t, "org.federation", "PostArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.post", "GetPostResponse"), ref.Message(t, "org.federation", "User")).
								Add(ref.Message(t, "org.federation", "M")).
								Add(ref.Message(t, "org.federation", "Z")).
								Build(t),
						).
						AddResolver(testutil.NewMessageResolverGroupByName("m")).
						AddResolver(
							testutil.NewMessageResolverGroupBuilder().
								AddStart(testutil.NewMessageResolverGroupByName("GetPost")).
								SetEnd(testutil.NewMessageResolver("user")).
								Build(t),
						).
						AddResolver(testutil.NewMessageResolverGroupByName("z")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostRequest").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostResponseArgument").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostResponse").
				AddFieldWithRule(
					"post",
					ref.Type(t, "org.federation", "Post"),
					testutil.NewFieldRuleBuilder(
						testutil.NewNameReferenceValueBuilder(
							ref.Type(t, "org.federation", "Post"),
							ref.Type(t, "org.federation", "Post"),
							"post",
						).Build(t),
					).Build(t),
				).
				AddFieldWithRule("literal", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("foo")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddMessageDependency(
							"post",
							ref.Message(t, "org.federation", "Post"),
							testutil.NewMessageDependencyArgumentBuilder().
								Add("id", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
								Add("double_value", resolver.NewDoubleValue(1.23)).
								Add("float_value", resolver.NewFloatValue(4.56)).
								Add("int32_value", resolver.NewInt32Value(1)).
								Add("int64_value", resolver.NewInt64Value(2)).
								Add("uint32_value", resolver.NewUint32Value(3)).
								Add("uint64_value", resolver.NewUint64Value(4)).
								Add("sint32_value", resolver.NewSint32Value(-1)).
								Add("sint64_value", resolver.NewSint64Value(-2)).
								Add("fixed32_value", resolver.NewFixed32Value(5)).
								Add("fixed64_value", resolver.NewFixed64Value(6)).
								Add("sfixed32_value", resolver.NewSfixed32Value(7)).
								Add("sfixed64_value", resolver.NewSfixed64Value(8)).
								Add("bool_value", resolver.NewBoolValue(true)).
								Add("string_value", resolver.NewStringValue("hello")).
								Add("bytes_value", resolver.NewBytesValue([]byte("hello"))).
								Add("message_value", resolver.NewMessageValue(ref.Type(t, "org.post", "Post"), map[string]*resolver.Value{"content": resolver.NewStringValue("xxxyyyzzz")})).
								Add("double_list_value", resolver.NewDoubleListValue(1.23, 4.56)).
								Add("float_list_value", resolver.NewFloatListValue(7.89, 1.23)).
								Add("int32_list_value", resolver.NewInt32ListValue(1, 2, 3)).
								Add("int64_list_value", resolver.NewInt64ListValue(4, 5, 6)).
								Add("uint32_list_value", resolver.NewUint32ListValue(7, 8, 9)).
								Add("uint64_list_value", resolver.NewUint64ListValue(10, 11, 12)).
								Add("sint32_list_value", resolver.NewSint32ListValue(-1, -2, -3)).
								Add("sint64_list_value", resolver.NewSint64ListValue(-4, -5, -6)).
								Add("fixed32_list_value", resolver.NewFixed32ListValue(11, 12, 13)).
								Add("fixed64_list_value", resolver.NewFixed64ListValue(14, 15, 16)).
								Add("sfixed32_list_value", resolver.NewSfixed32ListValue(-1, -2, -3)).
								Add("sfixed64_list_value", resolver.NewSfixed64ListValue(-4, -5, -6)).
								Add("bool_list_value", resolver.NewBoolListValue(false, true)).
								Add("string_list_value", resolver.NewStringListValue("hello", "world")).
								Add("bytes_list_value", resolver.NewBytesListValue([]byte("hello"), []byte("world"))).
								Add("message_list_value", resolver.NewMessageListValue(ref.RepeatedType(t, "org.post", "Post"), map[string]*resolver.Value{"content": resolver.NewStringValue("aaabbbccc")})).
								Build(t),
							false,
							true,
						).
						SetMessageArgument(ref.Message(t, "org.federation", "GetPostResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "Post")).
								Build(t),
						).
						AddResolver(testutil.NewMessageResolverGroupByName("post")).
						Build(t),
				).
				Build(t),
		).
		AddService(
			testutil.NewServiceBuilder("FederationService").
				AddMethod(
					"GetPost",
					ref.Message(t, "org.federation", "GetPostRequest"),
					ref.Message(t, "org.federation", "GetPostResponse"),
					testutil.NewMethodRuleBuilder().Timeout("1m").Build(t),
				).
				SetRule(
					testutil.NewServiceRuleBuilder().
						AddDependency("", ref.Service(t, "org.post", "PostService")).
						AddDependency("", ref.Service(t, "org.user", "UserService")).
						Build(t),
				).
				AddMessage(ref.Message(t, "org.federation", "GetPostResponse"), ref.Message(t, "org.federation", "GetPostResponseArgument")).
				AddMessage(ref.Message(t, "org.federation", "Post"), ref.Message(t, "org.federation", "PostArgument")).
				AddMessage(ref.Message(t, "org.federation", "User"), ref.Message(t, "org.federation", "UserArgument")).
				AddMessage(ref.Message(t, "org.federation", "Z"), ref.Message(t, "org.federation", "ZArgument")).
				AddMessage(ref.Message(t, "org.federation", "M"), ref.Message(t, "org.federation", "MArgument")).
				Build(t),
		)

	federationFile := fb.Build(t)
	federationService := federationFile.Services[0]

	r := resolver.New(testutil.Compile(t, filepath.Join(testutil.RepoRoot(), "testdata", "simple_aggregation.proto")))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Services))
	}
	if diff := cmp.Diff(result.Services[0], federationService, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}

	t.Run("candidates", func(t *testing.T) {
		candidates := r.Candidates(&source.Location{
			FileName: "simple_aggregation.proto",
			Message: &source.Message{
				Name: "Post",
				Option: &source.MessageOption{
					Resolver: &source.ResolverOption{
						Method: true,
					},
				},
			},
		})
		if diff := cmp.Diff(
			candidates, []string{
				"org.post.PostService/CreatePost",
				"org.post.PostService/GetPost",
				"org.post.PostService/GetPosts",
				"org.user.UserService/GetUser",
				"org.user.UserService/GetUsers",
			},
		); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
}

func TestCreatePost(t *testing.T) {
	fb := testutil.NewFileBuilder("create_post.proto")
	ref := testutil.NewBuilderReferenceManager(getUserProtoBuilder(t), getPostProtoBuilder(t), fb)

	fb.SetPackage("org.federation").
		SetGoPackage("example/federation", "federation").
		AddMessage(
			testutil.NewMessageBuilder("Post").
				AddFieldWithAlias("id", resolver.StringType, ref.Field(t, "org.post", "Post", "id")).
				AddFieldWithAlias("title", resolver.StringType, ref.Field(t, "org.post", "Post", "title")).
				AddFieldWithAlias("content", resolver.StringType, ref.Field(t, "org.post", "Post", "content")).
				AddFieldWithAlias("user_id", resolver.StringType, ref.Field(t, "org.post", "Post", "user_id")).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetAlias(ref.Message(t, "org.post", "Post")).
						SetDependencyGraph(testutil.NewDependencyGraphBuilder().Build(t)).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("CreatePostArgument").
				AddField("title", resolver.StringType).
				AddField("content", resolver.StringType).
				AddField("user_id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("CreatePost").
				AddFieldWithRule(
					"title",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "title").Build(t),
					).Build(t),
				).
				AddFieldWithRule(
					"content",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "content").Build(t),
					).Build(t),
				).
				AddFieldWithRule(
					"user_id",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "user_id").Build(t),
					).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "CreatePostArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("CreatePostRequest").
				AddField("title", resolver.StringType).
				AddField("content", resolver.StringType).
				AddField("user_id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("CreatePostResponseArgument").
				AddField("title", resolver.StringType).
				AddField("content", resolver.StringType).
				AddField("user_id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("CreatePostResponse").
				AddFieldWithRule(
					"post",
					ref.Type(t, "org.federation", "Post"),
					testutil.NewFieldRuleBuilder(
						testutil.NewNameReferenceValueBuilder(
							// Ref type is not org.post.Post.
							// To use the original response type when creating the dependency graph,
							// leave Ref as it is and use Filtered to calculate the name reference.
							ref.Type(t, "org.post", "CreatePostResponse"),
							ref.Type(t, "org.post", "Post"),
							"p",
						).Build(t),
					).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddMessageDependency(
							"cp",
							ref.Message(t, "org.federation", "CreatePost"),
							testutil.NewMessageDependencyArgumentBuilder().
								Add("title", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "title").Build(t)).
								Add("content", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "content").Build(t)).
								Add("user_id", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "user_id").Build(t)).
								Build(t),
							false,
							true,
						).
						SetMethodCall(
							testutil.NewMethodCallBuilder(ref.Method(t, "org.post", "PostService", "CreatePost")).
								SetRequest(
									testutil.NewRequestBuilder().
										AddField(
											"post",
											ref.Type(t, "org.post", "CreatePost"),
											testutil.NewNameReferenceValueBuilder(
												ref.Type(t, "org.federation", "CreatePost"),
												ref.Type(t, "org.federation", "CreatePost"),
												"cp",
											).Build(t),
										).
										Build(t),
								).
								SetResponse(
									testutil.NewResponseBuilder().
										AddField("p", "post", ref.Type(t, "org.post", "Post"), false, true).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "CreatePostResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "CreatePost")).
								Build(t),
						).
						AddResolver(
							testutil.NewMessageResolverGroupBuilder().
								AddStart(testutil.NewMessageResolverGroupByName("cp")).
								SetEnd(testutil.NewMessageResolver("CreatePost")).
								Build(t),
						).
						Build(t),
				).
				Build(t),
		).
		AddService(
			testutil.NewServiceBuilder("FederationService").
				AddMethod("CreatePost", ref.Message(t, "org.federation", "CreatePostRequest"), ref.Message(t, "org.federation", "CreatePostResponse"), nil).
				SetRule(
					testutil.NewServiceRuleBuilder().
						AddDependency("post_service", ref.Service(t, "org.post", "PostService")).
						Build(t),
				).
				AddMessage(ref.Message(t, "org.federation", "CreatePostResponse"), ref.Message(t, "org.federation", "CreatePostResponseArgument")).
				AddMessage(ref.Message(t, "org.federation", "CreatePost"), ref.Message(t, "org.federation", "CreatePostArgument")).
				Build(t),
		)

	federationFile := fb.Build(t)
	federationService := federationFile.Services[0]

	r := resolver.New(testutil.Compile(t, filepath.Join(testutil.RepoRoot(), "testdata", "create_post.proto")))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Services))
	}
	if diff := cmp.Diff(result.Services[0], federationService, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestMinimum(t *testing.T) {
	r := resolver.New(testutil.Compile(t, filepath.Join(testutil.RepoRoot(), "testdata", "minimum.proto")))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Services))
	}

	fb := testutil.NewFileBuilder("minimum.proto")
	ref := testutil.NewBuilderReferenceManager(fb)

	fb.SetPackage("org.federation").
		SetGoPackage("example/federation", "federation").
		AddMessage(
			testutil.NewMessageBuilder("User").
				AddField("id", resolver.StringType).
				AddField("name", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("Post").
				AddField("id", resolver.StringType).
				AddField("title", resolver.StringType).
				AddField("content", resolver.StringType).
				AddField("user", ref.Type(t, "org.federation", "User")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostRequest").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostResponseArgument").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostResponse").
				AddFieldWithRule(
					"post",
					ref.Type(t, "org.federation", "Post"),
					testutil.NewFieldRuleBuilder(nil).SetMessageCustomResolver(true).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "GetPostResponseArgument")).
						SetCustomResolver(true).
						SetDependencyGraph(testutil.NewDependencyGraphBuilder().Build(t)).
						Build(t),
				).
				Build(t),
		).
		AddService(
			testutil.NewServiceBuilder("FederationService").
				AddMethod("GetPost", ref.Message(t, "org.federation", "GetPostRequest"), ref.Message(t, "org.federation", "GetPostResponse"), nil).
				SetRule(testutil.NewServiceRuleBuilder().Build(t)).
				AddMessage(ref.Message(t, "org.federation", "GetPostResponse"), ref.Message(t, "org.federation", "GetPostResponseArgument")).
				Build(t),
		)
	federationFile := fb.Build(t)
	service := federationFile.Services[0]
	if diff := cmp.Diff(result.Services[0], service, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestCustomResolver(t *testing.T) {
	r := resolver.New(testutil.Compile(t, filepath.Join(testutil.RepoRoot(), "testdata", "custom_resolver.proto")))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Services))
	}

	fb := testutil.NewFileBuilder("custom_resolver.proto")
	ref := testutil.NewBuilderReferenceManager(getUserProtoBuilder(t), getPostProtoBuilder(t), fb)

	fb.SetPackage("org.federation").
		SetGoPackage("example/federation", "federation").
		AddMessage(
			testutil.NewMessageBuilder("PostArgument").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("UserArgument").
				AddField("id", resolver.StringType).
				AddField("title", resolver.StringType).
				AddField("content", resolver.StringType).
				AddField("user_id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("User").
				AddFieldWithRule(
					"id",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(nil).SetMessageCustomResolver(true).Build(t),
				).
				AddFieldWithRule(
					"name",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(nil).SetCustomResolver(true).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMethodCall(
							testutil.NewMethodCallBuilder(ref.Method(t, "org.user", "UserService", "GetUser")).
								SetRequest(
									testutil.NewRequestBuilder().
										AddField("id", resolver.StringType, testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "user_id").Build(t)).
										Build(t),
								).
								SetResponse(
									testutil.NewResponseBuilder().AddField("u", "user", ref.Type(t, "org.user", "User"), false, true).Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "UserArgument")).
						SetCustomResolver(true).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.user", "GetUserResponse")).
								Build(t),
						).
						AddResolver(testutil.NewMessageResolverGroupByName("GetUser")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("Post").
				AddFieldWithAutoBind("id", resolver.StringType, ref.Field(t, "org.post", "Post", "id")).
				AddFieldWithAutoBind("title", resolver.StringType, ref.Field(t, "org.post", "Post", "title")).
				AddFieldWithAutoBind("content", resolver.StringType, ref.Field(t, "org.post", "Post", "content")).
				AddFieldWithRule(
					"user",
					ref.Type(t, "org.federation", "User"),
					testutil.NewFieldRuleBuilder(nil).SetCustomResolver(true).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMethodCall(
							testutil.NewMethodCallBuilder(ref.Method(t, "org.post", "PostService", "GetPost")).
								SetRequest(
									testutil.NewRequestBuilder().
										AddField("id", resolver.StringType, testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
										Build(t),
								).
								SetResponse(
									testutil.NewResponseBuilder().
										AddField("post", "post", ref.Type(t, "org.post", "Post"), true, true).
										Build(t),
								).
								Build(t),
						).
						AddMessageDependency(
							"user",
							ref.Message(t, "org.federation", "User"),
							testutil.NewMessageDependencyArgumentBuilder().
								Inline(testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.post", "GetPostResponse"), ref.Type(t, "org.post", "Post"), "post").Build(t)).
								Build(t),
							false,
							true,
						).
						SetMessageArgument(ref.Message(t, "org.federation", "PostArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.post", "GetPostResponse"), ref.Message(t, "org.federation", "User")).
								Build(t),
						).
						AddResolver(
							testutil.NewMessageResolverGroupBuilder().
								AddStart(testutil.NewMessageResolverGroupByName("GetPost")).
								SetEnd(testutil.NewMessageResolver("user")).
								Build(t),
						).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostRequest").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostResponseArgument").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostResponse").
				AddFieldWithRule(
					"post",
					ref.Type(t, "org.federation", "Post"),
					testutil.NewFieldRuleBuilder(
						testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "Post"), ref.Type(t, "org.federation", "Post"), "post").
							Build(t),
					).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddMessageDependency(
							"post",
							ref.Message(t, "org.federation", "Post"),
							testutil.NewMessageDependencyArgumentBuilder().
								Add("id", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
								Build(t),
							false,
							true,
						).
						SetMessageArgument(ref.Message(t, "org.federation", "GetPostResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "Post")).
								Build(t),
						).
						AddResolver(testutil.NewMessageResolverGroupByName("post")).
						Build(t),
				).
				Build(t),
		).
		AddService(
			testutil.NewServiceBuilder("FederationService").
				AddMethod("GetPost", ref.Message(t, "org.federation", "GetPostRequest"), ref.Message(t, "org.federation", "GetPostResponse"), nil).
				SetRule(
					testutil.NewServiceRuleBuilder().
						AddDependency("post_service", ref.Service(t, "org.post", "PostService")).
						AddDependency("user_service", ref.Service(t, "org.user", "UserService")).
						Build(t),
				).
				AddMessage(ref.Message(t, "org.federation", "GetPostResponse"), ref.Message(t, "org.federation", "GetPostResponseArgument")).
				AddMessage(ref.Message(t, "org.federation", "Post"), ref.Message(t, "org.federation", "PostArgument")).
				AddMessage(ref.Message(t, "org.federation", "User"), ref.Message(t, "org.federation", "UserArgument")).
				Build(t),
		)

	federationFile := fb.Build(t)
	service := federationFile.Services[0]
	if diff := cmp.Diff(result.Services[0], service, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestAsync(t *testing.T) {
	r := resolver.New(testutil.Compile(t, filepath.Join(testutil.RepoRoot(), "testdata", "async.proto")))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Services))
	}

	fb := testutil.NewFileBuilder("async.proto")
	ref := testutil.NewBuilderReferenceManager(fb)

	fb.SetPackage("org.federation").
		SetGoPackage("example/federation", "federation").
		AddMessage(testutil.NewMessageBuilder("GetResponseArgument").Build(t)).
		AddMessage(testutil.NewMessageBuilder("AArgument").Build(t)).
		AddMessage(testutil.NewMessageBuilder("AAArgument").Build(t)).
		AddMessage(testutil.NewMessageBuilder("ABArgument").Build(t)).
		AddMessage(testutil.NewMessageBuilder("BArgument").Build(t)).
		AddMessage(testutil.NewMessageBuilder("CArgument").AddField("a", resolver.StringType).Build(t)).
		AddMessage(testutil.NewMessageBuilder("DArgument").AddField("b", resolver.StringType).Build(t)).
		AddMessage(
			testutil.NewMessageBuilder("EArgument").
				AddField("c", resolver.StringType).
				AddField("d", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("FArgument").
				AddField("c", resolver.StringType).
				AddField("d", resolver.StringType).
				Build(t),
		).
		AddMessage(testutil.NewMessageBuilder("GArgument").Build(t)).
		AddMessage(
			testutil.NewMessageBuilder("HArgument").
				AddField("e", resolver.StringType).
				AddField("f", resolver.StringType).
				AddField("g", resolver.StringType).
				Build(t),
		).
		AddMessage(testutil.NewMessageBuilder("IArgument").Build(t)).
		AddMessage(
			testutil.NewMessageBuilder("JArgument").
				AddField("i", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("AA").
				AddFieldWithRule("name", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("aa")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "AAArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("AB").
				AddFieldWithRule("name", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("ab")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "ABArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("A").
				AddFieldWithRule("name", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("a")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddMessageDependency(
							"aa",
							ref.Message(t, "org.federation", "AA"),
							testutil.NewMessageDependencyArgumentBuilder().Build(t),
							false,
							false,
						).
						AddMessageDependency(
							"ab",
							ref.Message(t, "org.federation", "AB"),
							testutil.NewMessageDependencyArgumentBuilder().Build(t),
							false,
							false,
						).
						SetMessageArgument(ref.Message(t, "org.federation", "AArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "AA")).
								Add(ref.Message(t, "org.federation", "AB")).
								Build(t),
						).
						AddResolver(testutil.NewMessageResolverGroupByName("aa")).
						AddResolver(testutil.NewMessageResolverGroupByName("ab")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("B").
				AddFieldWithRule("name", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("b")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "BArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("C").
				AddFieldWithRule("name", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("c")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "CArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("D").
				AddFieldWithRule("name", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("d")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "DArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("E").
				AddFieldWithRule("name", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("e")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "EArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("F").
				AddFieldWithRule("name", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("f")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "FArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("G").
				AddFieldWithRule("name", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("g")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "GArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("H").
				AddFieldWithRule("name", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("h")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "HArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("I").
				AddFieldWithRule("name", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("i")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "IArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("J").
				AddFieldWithRule("name", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("j")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "JArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(testutil.NewMessageBuilder("GetRequest").Build(t)).
		AddMessage(
			testutil.NewMessageBuilder("GetResponse").
				AddFieldWithRule(
					"hname",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "H"), resolver.StringType, "h.name").Build(t)).Build(t),
				).
				AddFieldWithRule(
					"jname",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "J"), resolver.StringType, "j.name").Build(t)).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddMessageDependency(
							"a",
							ref.Message(t, "org.federation", "A"),
							testutil.NewMessageDependencyArgumentBuilder().Build(t),
							false,
							true,
						).
						AddMessageDependency(
							"b",
							ref.Message(t, "org.federation", "B"),
							testutil.NewMessageDependencyArgumentBuilder().Build(t),
							false,
							true,
						).
						AddMessageDependency(
							"c",
							ref.Message(t, "org.federation", "C"),
							testutil.NewMessageDependencyArgumentBuilder().
								Add("a", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "A"), resolver.StringType, "a.name").Build(t)).
								Build(t),
							false,
							true,
						).
						AddMessageDependency(
							"d",
							ref.Message(t, "org.federation", "D"),
							testutil.NewMessageDependencyArgumentBuilder().
								Add("b", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "B"), resolver.StringType, "b.name").Build(t)).
								Build(t),
							false,
							true,
						).
						AddMessageDependency(
							"e",
							ref.Message(t, "org.federation", "E"),
							testutil.NewMessageDependencyArgumentBuilder().
								Add("c", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "C"), resolver.StringType, "c.name").Build(t)).
								Add("d", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "D"), resolver.StringType, "d.name").Build(t)).
								Build(t),
							false,
							true,
						).
						AddMessageDependency(
							"f",
							ref.Message(t, "org.federation", "F"),
							testutil.NewMessageDependencyArgumentBuilder().
								Add("c", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "C"), resolver.StringType, "c.name").Build(t)).
								Add("d", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "D"), resolver.StringType, "d.name").Build(t)).
								Build(t),
							false,
							true,
						).
						AddMessageDependency(
							"g",
							ref.Message(t, "org.federation", "G"),
							testutil.NewMessageDependencyArgumentBuilder().Build(t),
							false,
							true,
						).
						AddMessageDependency(
							"h",
							ref.Message(t, "org.federation", "H"),
							testutil.NewMessageDependencyArgumentBuilder().
								Add("e", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "E"), resolver.StringType, "e.name").Build(t)).
								Add("f", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "F"), resolver.StringType, "f.name").Build(t)).
								Add("g", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "G"), resolver.StringType, "g.name").Build(t)).
								Build(t),
							false,
							true,
						).
						AddMessageDependency(
							"i",
							ref.Message(t, "org.federation", "I"),
							testutil.NewMessageDependencyArgumentBuilder().Build(t),
							false,
							true,
						).
						AddMessageDependency(
							"j",
							ref.Message(t, "org.federation", "J"),
							testutil.NewMessageDependencyArgumentBuilder().
								Add("i", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "I"), resolver.StringType, "i.name").Build(t)).
								Build(t),
							false,
							true,
						).
						SetMessageArgument(ref.Message(t, "org.federation", "GetResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "A")).
								Add(ref.Message(t, "org.federation", "B")).
								Add(ref.Message(t, "org.federation", "G")).
								Add(ref.Message(t, "org.federation", "I")).
								Build(t),
						).
						AddResolver(
							testutil.NewMessageResolverGroupBuilder().
								AddStart(
									testutil.NewMessageResolverGroupBuilder().
										AddStart(
											testutil.NewMessageResolverGroupBuilder().
												AddStart(testutil.NewMessageResolverGroupByName("a")).
												SetEnd(testutil.NewMessageResolver("c")).
												Build(t),
										).
										AddStart(
											testutil.NewMessageResolverGroupBuilder().
												AddStart(testutil.NewMessageResolverGroupByName("b")).
												SetEnd(testutil.NewMessageResolver("d")).
												Build(t),
										).
										SetEnd(testutil.NewMessageResolver("e")).
										Build(t),
								).
								AddStart(
									testutil.NewMessageResolverGroupBuilder().
										AddStart(
											testutil.NewMessageResolverGroupBuilder().
												AddStart(testutil.NewMessageResolverGroupByName("a")).
												SetEnd(testutil.NewMessageResolver("c")).
												Build(t),
										).
										AddStart(
											testutil.NewMessageResolverGroupBuilder().
												AddStart(testutil.NewMessageResolverGroupByName("b")).
												SetEnd(testutil.NewMessageResolver("d")).
												Build(t),
										).
										SetEnd(testutil.NewMessageResolver("f")).
										Build(t),
								).
								AddStart(testutil.NewMessageResolverGroupByName("g")).
								SetEnd(testutil.NewMessageResolver("h")).
								Build(t),
						).
						AddResolver(
							testutil.NewMessageResolverGroupBuilder().
								AddStart(testutil.NewMessageResolverGroupByName("i")).
								SetEnd(testutil.NewMessageResolver("j")).
								Build(t),
						).
						Build(t),
				).
				Build(t),
		).
		AddService(
			testutil.NewServiceBuilder("FederationService").
				AddMethod("Get", ref.Message(t, "org.federation", "GetRequest"), ref.Message(t, "org.federation", "GetResponse"), nil).
				SetRule(testutil.NewServiceRuleBuilder().Build(t)).
				AddMessage(ref.Message(t, "org.federation", "GetResponse"), ref.Message(t, "org.federation", "GetResponseArgument")).
				AddMessage(ref.Message(t, "org.federation", "A"), ref.Message(t, "org.federation", "AArgument")).
				AddMessage(ref.Message(t, "org.federation", "AA"), ref.Message(t, "org.federation", "AAArgument")).
				AddMessage(ref.Message(t, "org.federation", "AB"), ref.Message(t, "org.federation", "ABArgument")).
				AddMessage(ref.Message(t, "org.federation", "B"), ref.Message(t, "org.federation", "BArgument")).
				AddMessage(ref.Message(t, "org.federation", "C"), ref.Message(t, "org.federation", "CArgument")).
				AddMessage(ref.Message(t, "org.federation", "D"), ref.Message(t, "org.federation", "DArgument")).
				AddMessage(ref.Message(t, "org.federation", "E"), ref.Message(t, "org.federation", "EArgument")).
				AddMessage(ref.Message(t, "org.federation", "F"), ref.Message(t, "org.federation", "FArgument")).
				AddMessage(ref.Message(t, "org.federation", "G"), ref.Message(t, "org.federation", "GArgument")).
				AddMessage(ref.Message(t, "org.federation", "H"), ref.Message(t, "org.federation", "HArgument")).
				AddMessage(ref.Message(t, "org.federation", "I"), ref.Message(t, "org.federation", "IArgument")).
				AddMessage(ref.Message(t, "org.federation", "J"), ref.Message(t, "org.federation", "JArgument")).
				Build(t),
		)

	federationFile := fb.Build(t)
	service := federationFile.Services[0]
	if diff := cmp.Diff(result.Services[0], service, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestAlias(t *testing.T) {
	r := resolver.New(testutil.Compile(t, filepath.Join(testutil.RepoRoot(), "testdata", "alias.proto")))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Services))
	}

	fb := testutil.NewFileBuilder("alias.proto")
	ref := testutil.NewBuilderReferenceManager(getNestedPostProtoBuilder(t), fb)

	fb.SetPackage("org.federation").
		SetGoPackage("example/federation", "federation").
		AddEnum(
			testutil.NewEnumBuilder("PostType").
				WithAlias(ref.Enum(t, "org.post", "PostDataType")).
				AddValueWithDefault("POST_TYPE_UNKNOWN").
				AddValueWithAlias(
					"POST_TYPE_FOO",
					ref.EnumValue(t, "org.post", "PostDataType", "POST_TYPE_A"),
				).
				AddValueWithAlias(
					"POST_TYPE_BAR",
					ref.EnumValue(t, "org.post", "PostDataType", "POST_TYPE_B"),
					ref.EnumValue(t, "org.post", "PostDataType", "POST_TYPE_C"),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("PostContent").
				AddEnum(
					testutil.NewEnumBuilder("Category").
						WithAlias(ref.Enum(t, "org.post", "PostContent.Category")).
						AddValueWithAlias("CATEGORY_A", ref.EnumValue(t, "org.post", "PostContent.Category", "CATEGORY_A")).
						AddValueWithAlias("CATEGORY_B", ref.EnumValue(t, "org.post", "PostContent.Category", "CATEGORY_B")).
						Build(t),
				).
				AddFieldWithTypeNameAndAlias(t, "category", "Category", false, ref.Field(t, "org.post", "PostContent", "category")).
				AddFieldWithAlias("head", resolver.StringType, ref.Field(t, "org.post", "PostContent", "head")).
				AddFieldWithAlias("body", resolver.StringType, ref.Field(t, "org.post", "PostContent", "body")).
				AddFieldWithAlias("dup_body", resolver.StringType, ref.Field(t, "org.post", "PostContent", "body")).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetAlias(ref.Message(t, "org.post", "PostContent")).
						SetDependencyGraph(testutil.NewDependencyGraphBuilder().Build(t)).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("PostData").
				AddFieldWithAlias("type", ref.Type(t, "org.federation", "PostType"), ref.Field(t, "org.post", "PostData", "type")).
				AddFieldWithAlias("title", resolver.StringType, ref.Field(t, "org.post", "PostData", "title")).
				AddFieldWithAlias("content", ref.Type(t, "org.federation", "PostContent"), ref.Field(t, "org.post", "PostData", "content")).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetAlias(ref.Message(t, "org.post", "PostData")).
						SetDependencyGraph(testutil.NewDependencyGraphBuilder().Build(t)).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("PostArgument").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("Post").
				AddFieldWithAutoBind("id", resolver.StringType, ref.Field(t, "org.post", "Post", "id")).
				AddFieldWithAutoBind("data", ref.Type(t, "org.federation", "PostData"), ref.Field(t, "org.post", "Post", "data")).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMethodCall(
							testutil.NewMethodCallBuilder(ref.Method(t, "org.post", "PostService", "GetPost")).
								SetRequest(
									testutil.NewRequestBuilder().
										AddField("id", resolver.StringType, testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
										Build(t),
								).
								SetResponse(
									testutil.NewResponseBuilder().
										AddField("post", "post", ref.Type(t, "org.post", "Post"), true, true).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "PostArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.post", "GetPostResponse")).
								Build(t),
						).
						AddResolver(
							testutil.NewMessageResolverGroupBuilder().
								SetEnd(testutil.NewMessageResolver("GetPost")).
								Build(t),
						).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostRequest").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostResponseArgument").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostResponse").
				AddFieldWithRule(
					"post",
					ref.Type(t, "org.federation", "Post"),
					testutil.NewFieldRuleBuilder(
						testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "Post"), ref.Type(t, "org.federation", "Post"), "post").
							Build(t),
					).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddMessageDependency(
							"post",
							ref.Message(t, "org.federation", "Post"),
							testutil.NewMessageDependencyArgumentBuilder().
								Add("id", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
								Build(t),
							false,
							true,
						).
						SetMessageArgument(ref.Message(t, "org.federation", "GetPostResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "Post")).
								Build(t),
						).
						AddResolver(testutil.NewMessageResolverGroupByName("post")).
						Build(t),
				).
				Build(t),
		).
		AddService(
			testutil.NewServiceBuilder("FederationService").
				AddMethod("GetPost", ref.Message(t, "org.federation", "GetPostRequest"), ref.Message(t, "org.federation", "GetPostResponse"), nil).
				SetRule(
					testutil.NewServiceRuleBuilder().
						AddDependency("post_service", ref.Service(t, "org.post", "PostService")).
						Build(t),
				).
				AddMessage(ref.Message(t, "org.federation", "GetPostResponse"), ref.Message(t, "org.federation", "GetPostResponseArgument")).
				AddMessage(ref.Message(t, "org.federation", "Post"), ref.Message(t, "org.federation", "PostArgument")).
				Build(t),
		)
	federationFile := fb.Build(t)
	federationService := federationFile.Services[0]

	if diff := cmp.Diff(result.Services[0], federationService, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func getUserProtoBuilder(t *testing.T) *testutil.FileBuilder {
	ub := testutil.NewFileBuilder("user.proto")
	ref := testutil.NewBuilderReferenceManager(ub)
	ub.SetPackage("org.user").
		SetGoPackage("example/user", "user").
		AddEnum(
			testutil.NewEnumBuilder("UserType").
				AddValue("USER_TYPE_1").
				AddValue("USER_TYPE_2").
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("Item").
				AddEnum(
					testutil.NewEnumBuilder("ItemType").
						AddValue("ITEM_TYPE_1").
						AddValue("ITEM_TYPE_2").
						AddValue("ITEM_TYPE_3").
						Build(t),
				).
				AddField("name", resolver.StringType).
				AddFieldWithTypeName(t, "type", "ItemType", false).
				AddField("value", resolver.Int64Type).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("User").
				AddMessage(
					testutil.NewMessageBuilder("ProfileEntry").
						SetIsMapEntry(true).
						AddField("key", resolver.StringType).
						AddField("value", resolver.AnyType).
						Build(t),
				).
				AddField("id", resolver.StringType).
				AddField("type", ref.Type(t, "org.user", "UserType")).
				AddField("name", resolver.StringType).
				AddField("age", resolver.Int64Type).
				AddField("desc", resolver.StringRepeatedType).
				AddField("main_item", ref.Type(t, "org.user", "Item")).
				AddField("items", ref.RepeatedType(t, "org.user", "Item")).
				AddFieldWithTypeName(t, "profile", "ProfileEntry", true).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetUserRequest").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetUserResponse").
				AddField("user", ref.Type(t, "org.user", "User")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetUsersRequest").
				AddField("ids", resolver.StringRepeatedType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetUsersResponse").
				AddField("users", ref.RepeatedType(t, "org.user", "User")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetUserResponseArgument").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetUsersResponseArgument").
				AddField("ids", resolver.StringRepeatedType).
				Build(t),
		).
		AddService(
			testutil.NewServiceBuilder("UserService").
				AddMethod("GetUser", ref.Message(t, "org.user", "GetUserRequest"), ref.Message(t, "org.user", "GetUserResponse"), nil).
				AddMethod("GetUsers", ref.Message(t, "org.user", "GetUsersRequest"), ref.Message(t, "org.user", "GetUsersResponse"), nil).
				AddMessage(nil, ref.Message(t, "org.user", "GetUserResponseArgument")).
				AddMessage(nil, ref.Message(t, "org.user", "GetUsersResponseArgument")).
				Build(t),
		)
	return ub
}

func getPostProtoBuilder(t *testing.T) *testutil.FileBuilder {
	pb := testutil.NewFileBuilder("post.proto")
	ref := testutil.NewBuilderReferenceManager(pb)

	pb.SetPackage("org.post").
		SetGoPackage("example/post", "post").
		AddMessage(
			testutil.NewMessageBuilder("Post").
				AddField("id", resolver.StringType).
				AddField("title", resolver.StringType).
				AddField("content", resolver.StringType).
				AddField("user_id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostRequest").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostResponse").
				AddField("post", ref.Type(t, "org.post", "Post")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostsRequest").
				AddField("ids", resolver.StringRepeatedType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostsResponse").
				AddField("posts", ref.RepeatedType(t, "org.post", "Post")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("CreatePost").
				AddField("title", resolver.StringType).
				AddField("content", resolver.StringType).
				AddField("user_id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("CreatePostRequest").
				AddField("post", ref.Type(t, "org.post", "CreatePost")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("CreatePostResponse").
				AddField("post", ref.Type(t, "org.post", "Post")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostResponseArgument").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostsResponseArgument").
				AddField("ids", resolver.StringRepeatedType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("CreatePostResponseArgument").
				AddField("post", ref.Type(t, "org.post", "CreatePost")).
				Build(t),
		).
		AddService(
			testutil.NewServiceBuilder("PostService").
				AddMethod("GetPost", ref.Message(t, "org.post", "GetPostRequest"), ref.Message(t, "org.post", "GetPostResponse"), nil).
				AddMethod("GetPosts", ref.Message(t, "org.post", "GetPostsRequest"), ref.Message(t, "org.post", "GetPostsResponse"), nil).
				AddMethod("CreatePost", ref.Message(t, "org.post", "CreatePostRequest"), ref.Message(t, "org.post", "CreatePostResponse"), nil).
				AddMessage(nil, ref.Message(t, "org.post", "GetPostResponseArgument")).
				AddMessage(nil, ref.Message(t, "org.post", "GetPostsResponseArgument")).
				AddMessage(nil, ref.Message(t, "org.post", "CreatePostResponseArgument")).
				Build(t),
		)
	return pb
}

func getNestedPostProtoBuilder(t *testing.T) *testutil.FileBuilder {
	t.Helper()
	pb := testutil.NewFileBuilder("nested_post.proto")
	ref := testutil.NewBuilderReferenceManager(pb)

	pb.SetPackage("org.post").
		SetGoPackage("example/post", "post").
		AddEnum(
			testutil.NewEnumBuilder("PostDataType").
				AddValue("POST_TYPE_A").
				AddValue("POST_TYPE_B").
				AddValue("POST_TYPE_C").
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("PostContent").
				AddEnum(
					testutil.NewEnumBuilder("Category").
						AddValue("CATEGORY_A").
						AddValue("CATEGORY_B").
						Build(t),
				).
				AddFieldWithTypeName(t, "category", "Category", false).
				AddField("head", resolver.StringType).
				AddField("body", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("PostData").
				AddField("type", ref.Type(t, "org.post", "PostDataType")).
				AddField("title", resolver.StringType).
				AddField("content", ref.Type(t, "org.post", "PostContent")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("Post").
				AddField("id", resolver.StringType).
				AddField("data", ref.Type(t, "org.post", "PostData")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostRequest").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostResponse").
				AddField("post", ref.Type(t, "org.post", "Post")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostResponseArgument").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddService(
			testutil.NewServiceBuilder("PostService").
				AddMethod("GetPost", ref.Message(t, "org.post", "GetPostRequest"), ref.Message(t, "org.post", "GetPostResponse"), nil).
				AddMessage(nil, ref.Message(t, "org.post", "GetPostResponseArgument")).
				Build(t),
		)
	return pb

}
