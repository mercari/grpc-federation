package resolver_test

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	exprv1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/genproto/googleapis/rpc/code"

	"github.com/mercari/grpc-federation/internal/testutil"
	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/source"
)

func TestSimpleAggregation(t *testing.T) {
	fileName := filepath.Join(testutil.RepoRoot(), "testdata", "simple_aggregation.proto")
	fb := testutil.NewFileBuilder(fileName)
	ref := testutil.NewBuilderReferenceManager(getUserProtoBuilder(t), getPostProtoBuilder(t), fb)

	fb.SetPackage("org.federation").
		SetGoPackage("example/federation", "federation").
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
				AddMessage(
					testutil.NewMessageBuilder("ProfileEntry").
						SetIsMapEntry(true).
						AddField("key", resolver.StringType).
						AddField("value", resolver.AnyType).
						Build(t),
				).
				AddMessage(
					testutil.NewMessageBuilder("AttrA").
						AddFieldWithAlias("foo", resolver.StringType, ref.Field(t, "org.user", "User.AttrA", "foo")).
						SetRule(
							testutil.NewMessageRuleBuilder().
								SetAlias(ref.Message(t, "org.user", "User.AttrA")).
								Build(t),
						).
						Build(t),
				).
				AddMessage(
					testutil.NewMessageBuilder("AttrB").
						AddFieldWithAlias("bar", resolver.BoolType, ref.Field(t, "org.user", "User.AttrB", "bar")).
						SetRule(
							testutil.NewMessageRuleBuilder().
								SetAlias(ref.Message(t, "org.user", "User.AttrB")).
								Build(t),
						).
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
				AddFieldWithTypeNameAndAutoBind(t, "attr_a", "AttrA", false, ref.Field(t, "org.user", "User", "attr_a")).
				AddFieldWithTypeNameAndAutoBind(t, "b", "AttrB", false, ref.Field(t, "org.user", "User", "b")).
				AddOneof(testutil.NewOneofBuilder("attr").AddFieldNames("attr_a", "b").Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("res").
								SetUsed(true).
								SetCall(
									testutil.NewCallExprBuilder().
										SetMethod(ref.Method(t, "org.user", "UserService", "GetUser")).
										SetRequest(
											testutil.NewRequestBuilder().
												AddField(
													"id",
													resolver.StringType,
													testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "user_id").Build(t),
												).
												Build(t),
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
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("user").
								SetUsed(true).
								SetAutoBind(true).
								SetBy(testutil.NewCELValueBuilder("res.user", ref.Type(t, "org.user", "User")).Build(t)).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "UserArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.user", "GetUserResponse")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(testutil.NewVariableDefinitionGroupByName("res")).
								SetEnd(testutil.NewVariableDefinition("user")).
								Build(t),
						).
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
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("res").
								SetUsed(true).
								SetCall(
									testutil.NewCallExprBuilder().
										SetMethod(ref.Method(t, "org.post", "PostService", "GetPost")).
										SetRequest(
											testutil.NewRequestBuilder().
												AddField("id", resolver.StringType, testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
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
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("post").
								SetUsed(true).
								SetAutoBind(true).
								SetBy(testutil.NewCELValueBuilder("res.post", ref.Type(t, "org.post", "Post")).Build(t)).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("user").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "User")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Inline(testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.post", "GetPostResponse"), ref.Type(t, "org.post", "Post"), "post").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("z").
								SetUsed(false).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "Z")).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("m").
								SetUsed(true).
								SetAutoBind(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "M")).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "PostArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "M")).
								Add(ref.Message(t, "org.federation", "Z")).
								Add(ref.Message(t, "org.post", "GetPostResponse"), ref.Message(t, "org.federation", "User")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("m")).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(
									testutil.NewVariableDefinitionGroupBuilder().
										AddStart(testutil.NewVariableDefinitionGroupByName("res")).
										SetEnd(testutil.NewVariableDefinition("post")).
										Build(t),
								).
								SetEnd(testutil.NewVariableDefinition("user")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("z")).
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
				AddFieldWithRule("const", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("foo")).Build(t)).
				AddFieldWithRule("uuid", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewByValue("uuid.string()", resolver.StringType)).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("post").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "Post")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("id", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("uuid").
								SetUsed(true).
								SetBy(testutil.NewCELValueBuilder("grpc.federation.uuid.newRandom()", resolver.NewCELStandardLibraryMessageType("uuid", "UUID")).Build(t)).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "GetPostResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "Post")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("post")).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("uuid")).
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
				AddMessage(ref.Message(t, "org.federation", "M"), ref.Message(t, "org.federation", "MArgument")).
				AddMessage(ref.Message(t, "org.federation", "Post"), ref.Message(t, "org.federation", "PostArgument")).
				AddMessage(ref.Message(t, "org.federation", "User"), ref.Message(t, "org.federation", "UserArgument")).
				AddMessage(ref.Message(t, "org.federation", "Z"), ref.Message(t, "org.federation", "ZArgument")).
				Build(t),
		)

	federationFile := fb.Build(t)
	federationService := federationFile.Services[0]

	r := resolver.New(testutil.Compile(t, fileName))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("faield to get files. expected 1 but got %d", len(result.Files))
	}
	if len(result.Files[0].Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Files[0].Services))
	}
	if diff := cmp.Diff(result.Files[0].Services[0], federationService, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}

	t.Run("candidates", func(t *testing.T) {
		candidates := r.Candidates(&source.Location{
			FileName: fileName,
			Message: &source.Message{
				Name: "Post",
				Option: &source.MessageOption{
					Def: &source.VariableDefinitionOption{
						Call: &source.CallExprOption{
							Method: true,
						},
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
	fileName := filepath.Join(testutil.RepoRoot(), "testdata", "create_post.proto")
	fb := testutil.NewFileBuilder(fileName)
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
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("cp").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "CreatePost")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("title", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "title").Build(t)).
												Add("content", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "content").Build(t)).
												Add("user_id", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "user_id").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("res").
								SetUsed(true).
								SetCall(
									testutil.NewCallExprBuilder().
										SetMethod(ref.Method(t, "org.post", "PostService", "CreatePost")).
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
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("p").
								SetUsed(true).
								SetBy(testutil.NewCELValueBuilder("res.post", ref.Type(t, "org.post", "Post")).Build(t)).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "CreatePostResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "CreatePost")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(testutil.NewVariableDefinitionGroupBuilder().
									AddStart(testutil.NewVariableDefinitionGroupByName("cp")).
									SetEnd(testutil.NewVariableDefinition("res")).
									Build(t),
								).
								SetEnd(testutil.NewVariableDefinition("p")).
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
				AddMessage(ref.Message(t, "org.federation", "CreatePost"), ref.Message(t, "org.federation", "CreatePostArgument")).
				AddMessage(ref.Message(t, "org.federation", "CreatePostResponse"), ref.Message(t, "org.federation", "CreatePostResponseArgument")).
				Build(t),
		)

	federationFile := fb.Build(t)
	federationService := federationFile.Services[0]

	r := resolver.New(testutil.Compile(t, fileName))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("faield to get files. expected 1 but got %d", len(result.Files))
	}
	if len(result.Files[0].Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Files[0].Services))
	}
	if diff := cmp.Diff(result.Files[0].Services[0], federationService, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestMinimum(t *testing.T) {
	fileName := filepath.Join(testutil.RepoRoot(), "testdata", "minimum.proto")
	r := resolver.New(testutil.Compile(t, fileName))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("faield to get files. expected 1 but got %d", len(result.Files))
	}
	if len(result.Files[0].Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Files[0].Services))
	}

	fb := testutil.NewFileBuilder(fileName)
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
		AddEnum(
			testutil.NewEnumBuilder("PostType").
				AddValue("POST_TYPE_1").
				AddValue("POST_TYPE_2").
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostRequest").
				AddField("id", resolver.StringType).
				AddField("type", ref.Type(t, "org.federation", "PostType")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostResponseArgument").
				AddField("id", resolver.StringType).
				AddField("type", ref.Type(t, "org.federation", "PostType")).
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
	if diff := cmp.Diff(result.Files[0].Services[0], service, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestCustomResolver(t *testing.T) {
	fileName := filepath.Join(testutil.RepoRoot(), "testdata", "custom_resolver.proto")
	r := resolver.New(testutil.Compile(t, fileName))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("faield to get files. expected 1 but got %d", len(result.Files))
	}
	if len(result.Files[0].Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Files[0].Services))
	}

	fb := testutil.NewFileBuilder(fileName)
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
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("res").
								SetUsed(true).
								SetCall(
									testutil.NewCallExprBuilder().
										SetMethod(ref.Method(t, "org.user", "UserService", "GetUser")).
										SetRequest(
											testutil.NewRequestBuilder().
												AddField("id", resolver.StringType, testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "user_id").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("u").
								SetUsed(true).
								SetBy(testutil.NewCELValueBuilder("res.user", ref.Type(t, "org.user", "User")).Build(t)).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "UserArgument")).
						SetCustomResolver(true).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.user", "GetUserResponse")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(testutil.NewVariableDefinitionGroupByName("res")).
								SetEnd(testutil.NewVariableDefinition("u")).
								Build(t),
						).
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
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("res").
								SetUsed(true).
								SetCall(
									testutil.NewCallExprBuilder().
										SetMethod(ref.Method(t, "org.post", "PostService", "GetPost")).
										SetRequest(
											testutil.NewRequestBuilder().
												AddField("id", resolver.StringType, testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("post").
								SetUsed(true).
								SetAutoBind(true).
								SetBy(testutil.NewCELValueBuilder("res.post", ref.Type(t, "org.post", "Post")).Build(t)).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("user").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "User")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Inline(testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.post", "GetPostResponse"), ref.Type(t, "org.post", "Post"), "post").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "PostArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.post", "GetPostResponse"), ref.Message(t, "org.federation", "User")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(testutil.NewVariableDefinitionGroupBuilder().
									AddStart(testutil.NewVariableDefinitionGroupByName("res")).
									SetEnd(testutil.NewVariableDefinition("post")).
									Build(t),
								).
								SetEnd(testutil.NewVariableDefinition("user")).
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
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("post").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "Post")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("id", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "GetPostResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "Post")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("post")).
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
	if diff := cmp.Diff(result.Files[0].Services[0], service, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestAsync(t *testing.T) {
	fileName := filepath.Join(testutil.RepoRoot(), "testdata", "async.proto")
	r := resolver.New(testutil.Compile(t, fileName))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("faield to get files. expected 1 but got %d", len(result.Files))
	}
	if len(result.Files[0].Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Files[0].Services))
	}

	fb := testutil.NewFileBuilder(fileName)
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
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("aa").
								SetUsed(false).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "AA")).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("ab").
								SetUsed(false).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "AB")).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "AArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "AA")).
								Add(ref.Message(t, "org.federation", "AB")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("aa")).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("ab")).
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
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("a").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "A")).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("b").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "B")).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("c").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "C")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("a", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "A"), resolver.StringType, "a.name").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("d").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "D")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("b", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "B"), resolver.StringType, "b.name").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("e").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "E")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("c", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "C"), resolver.StringType, "c.name").Build(t)).
												Add("d", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "D"), resolver.StringType, "d.name").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("f").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "F")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("c", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "C"), resolver.StringType, "c.name").Build(t)).
												Add("d", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "D"), resolver.StringType, "d.name").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("g").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "G")).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("h").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "H")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("e", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "E"), resolver.StringType, "e.name").Build(t)).
												Add("f", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "F"), resolver.StringType, "f.name").Build(t)).
												Add("g", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "G"), resolver.StringType, "g.name").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("i").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "I")).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("j").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "J")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("i", testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "I"), resolver.StringType, "i.name").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
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
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(
									testutil.NewVariableDefinitionGroupBuilder().
										AddStart(
											testutil.NewVariableDefinitionGroupBuilder().
												AddStart(testutil.NewVariableDefinitionGroupByName("a")).
												SetEnd(testutil.NewVariableDefinition("c")).
												Build(t),
										).
										AddStart(
											testutil.NewVariableDefinitionGroupBuilder().
												AddStart(testutil.NewVariableDefinitionGroupByName("b")).
												SetEnd(testutil.NewVariableDefinition("d")).
												Build(t),
										).
										SetEnd(testutil.NewVariableDefinition("e")).
										Build(t),
								).
								AddStart(
									testutil.NewVariableDefinitionGroupBuilder().
										AddStart(
											testutil.NewVariableDefinitionGroupBuilder().
												AddStart(testutil.NewVariableDefinitionGroupByName("a")).
												SetEnd(testutil.NewVariableDefinition("c")).
												Build(t),
										).
										AddStart(
											testutil.NewVariableDefinitionGroupBuilder().
												AddStart(testutil.NewVariableDefinitionGroupByName("b")).
												SetEnd(testutil.NewVariableDefinition("d")).
												Build(t),
										).
										SetEnd(testutil.NewVariableDefinition("f")).
										Build(t),
								).
								AddStart(testutil.NewVariableDefinitionGroupByName("g")).
								SetEnd(testutil.NewVariableDefinition("h")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(testutil.NewVariableDefinitionGroupByName("i")).
								SetEnd(testutil.NewVariableDefinition("j")).
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
				AddMessage(ref.Message(t, "org.federation", "A"), ref.Message(t, "org.federation", "AArgument")).
				AddMessage(ref.Message(t, "org.federation", "AA"), ref.Message(t, "org.federation", "AAArgument")).
				AddMessage(ref.Message(t, "org.federation", "AB"), ref.Message(t, "org.federation", "ABArgument")).
				AddMessage(ref.Message(t, "org.federation", "B"), ref.Message(t, "org.federation", "BArgument")).
				AddMessage(ref.Message(t, "org.federation", "C"), ref.Message(t, "org.federation", "CArgument")).
				AddMessage(ref.Message(t, "org.federation", "D"), ref.Message(t, "org.federation", "DArgument")).
				AddMessage(ref.Message(t, "org.federation", "E"), ref.Message(t, "org.federation", "EArgument")).
				AddMessage(ref.Message(t, "org.federation", "F"), ref.Message(t, "org.federation", "FArgument")).
				AddMessage(ref.Message(t, "org.federation", "G"), ref.Message(t, "org.federation", "GArgument")).
				AddMessage(ref.Message(t, "org.federation", "GetResponse"), ref.Message(t, "org.federation", "GetResponseArgument")).
				AddMessage(ref.Message(t, "org.federation", "H"), ref.Message(t, "org.federation", "HArgument")).
				AddMessage(ref.Message(t, "org.federation", "I"), ref.Message(t, "org.federation", "IArgument")).
				AddMessage(ref.Message(t, "org.federation", "J"), ref.Message(t, "org.federation", "JArgument")).
				Build(t),
		)

	federationFile := fb.Build(t)
	service := federationFile.Services[0]
	if diff := cmp.Diff(result.Files[0].Services[0], service, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestAlias(t *testing.T) {
	fileName := filepath.Join(testutil.RepoRoot(), "testdata", "alias.proto")
	r := resolver.New(testutil.Compile(t, fileName))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("faield to get files. expected 1 but got %d", len(result.Files))
	}
	if len(result.Files[0].Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Files[0].Services))
	}

	fb := testutil.NewFileBuilder(fileName)
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
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("res").
								SetUsed(true).
								SetCall(
									testutil.NewCallExprBuilder().
										SetMethod(ref.Method(t, "org.post", "PostService", "GetPost")).
										SetRequest(
											testutil.NewRequestBuilder().
												AddField("id", resolver.StringType, testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("post").
								SetUsed(true).
								SetAutoBind(true).
								SetBy(testutil.NewCELValueBuilder("res.post", ref.Type(t, "org.post", "Post")).Build(t)).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "PostArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.post", "GetPostResponse")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(testutil.NewVariableDefinitionGroupByName("res")).
								SetEnd(testutil.NewVariableDefinition("post")).
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
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("post").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "Post")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("id", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "GetPostResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "Post")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("post")).
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

	if diff := cmp.Diff(result.Files[0].Services[0], federationService, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestAutobind(t *testing.T) {
	fileName := filepath.Join(testutil.RepoRoot(), "testdata", "autobind.proto")
	r := resolver.New(testutil.Compile(t, fileName))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("faield to get files. expected 1 but got %d", len(result.Files))
	}
	if len(result.Files[0].Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Files[0].Services))
	}

	fb := testutil.NewFileBuilder(fileName)
	ref := testutil.NewBuilderReferenceManager(getPostProtoBuilder(t), fb)

	fb.SetPackage("org.federation").
		SetGoPackage("example/federation", "federation").
		AddMessage(
			testutil.NewMessageBuilder("UserArgument").
				AddField("user_id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("User").
				AddFieldWithRule(
					"uid",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "user_id").Build(t),
					).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "UserArgument")).
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
				AddFieldWithAutoBind("title", resolver.StringType, ref.Field(t, "org.post", "Post", "title")).
				AddFieldWithAutoBind("content", resolver.StringType, ref.Field(t, "org.post", "Post", "content")).
				AddFieldWithAutoBind("uid", resolver.StringType, ref.Field(t, "org.federation", "User", "uid")).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("res").
								SetUsed(true).
								SetCall(
									testutil.NewCallExprBuilder().
										SetMethod(ref.Method(t, "org.post", "PostService", "GetPost")).
										SetRequest(
											testutil.NewRequestBuilder().
												AddField("id", resolver.StringType, testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("_def1").
								SetUsed(true).
								SetAutoBind(true).
								SetBy(testutil.NewCELValueBuilder("res.post", ref.Type(t, "org.post", "Post")).Build(t)).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("_def2").
								SetUsed(true).
								SetAutoBind(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "User")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("user_id", resolver.NewStringValue("foo")).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "PostArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "User")).
								Add(ref.Message(t, "org.post", "GetPostResponse")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(testutil.NewVariableDefinitionGroupByName("res")).
								SetEnd(testutil.NewVariableDefinition("_def1")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("_def2")).
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
				AddFieldWithAutoBind("id", resolver.StringType, ref.Field(t, "org.federation", "Post", "id")).
				AddFieldWithAutoBind("title", resolver.StringType, ref.Field(t, "org.federation", "Post", "title")).
				AddFieldWithAutoBind("content", resolver.StringType, ref.Field(t, "org.federation", "Post", "content")).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("_def0").
								SetUsed(true).
								SetAutoBind(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "Post")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("id", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "GetPostResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "Post")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("_def0")).
						Build(t),
				).
				Build(t),
		).
		AddService(
			testutil.NewServiceBuilder("FederationService").
				AddMethod("GetPost", ref.Message(t, "org.federation", "GetPostRequest"), ref.Message(t, "org.federation", "GetPostResponse"), nil).
				SetRule(
					testutil.NewServiceRuleBuilder().
						AddDependency("", ref.Service(t, "org.post", "PostService")).
						Build(t),
				).
				AddMessage(ref.Message(t, "org.federation", "GetPostResponse"), ref.Message(t, "org.federation", "GetPostResponseArgument")).
				AddMessage(ref.Message(t, "org.federation", "Post"), ref.Message(t, "org.federation", "PostArgument")).
				AddMessage(ref.Message(t, "org.federation", "User"), ref.Message(t, "org.federation", "UserArgument")).
				Build(t),
		)
	federationFile := fb.Build(t)
	federationService := federationFile.Services[0]

	if diff := cmp.Diff(result.Files[0].Services[0], federationService, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestConstValue(t *testing.T) {
	fileName := filepath.Join(testutil.RepoRoot(), "testdata", "const_value.proto")
	r := resolver.New(testutil.Compile(t, fileName))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("faield to get files. expected 1 but got %d", len(result.Files))
	}
	if len(result.Files[0].Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Files[0].Services))
	}

	fb := testutil.NewFileBuilder(fileName)
	ref := testutil.NewBuilderReferenceManager(getContentProtoBuilder(t), fb)

	fb.SetPackage("org.federation").
		SetGoPackage("example/federation", "federation").
		AddEnum(
			testutil.NewEnumBuilder("ContentType").
				WithAlias(ref.Enum(t, "content", "ContentType")).
				AddValueWithAlias("CONTENT_TYPE_1", ref.EnumValue(t, "content", "ContentType", "CONTENT_TYPE_1")).
				AddValueWithAlias("CONTENT_TYPE_2", ref.EnumValue(t, "content", "ContentType", "CONTENT_TYPE_2")).
				AddValueWithAlias("CONTENT_TYPE_3", ref.EnumValue(t, "content", "ContentType", "CONTENT_TYPE_3")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("ContentArgument").
				AddField("by_field", resolver.StringType).
				AddField("double_field", resolver.DoubleType).
				AddField("doubles_field", resolver.DoubleRepeatedType).
				AddField("float_field", resolver.FloatType).
				AddField("floats_field", resolver.FloatRepeatedType).
				AddField("int32_field", resolver.Int32Type).
				AddField("int32s_field", resolver.Int32RepeatedType).
				AddField("int64_field", resolver.Int64Type).
				AddField("int64s_field", resolver.Int64RepeatedType).
				AddField("uint32_field", resolver.Uint32Type).
				AddField("uint32s_field", resolver.Uint32RepeatedType).
				AddField("uint64_field", resolver.Uint64Type).
				AddField("uint64s_field", resolver.Uint64RepeatedType).
				AddField("sint32_field", resolver.Sint32Type).
				AddField("sint32s_field", resolver.Sint32RepeatedType).
				AddField("sint64_field", resolver.Sint64Type).
				AddField("sint64s_field", resolver.Sint64RepeatedType).
				AddField("fixed32_field", resolver.Fixed32Type).
				AddField("fixed32s_field", resolver.Fixed32RepeatedType).
				AddField("fixed64_field", resolver.Fixed64Type).
				AddField("fixed64s_field", resolver.Fixed64RepeatedType).
				AddField("sfixed32_field", resolver.Sfixed32Type).
				AddField("sfixed32s_field", resolver.Sfixed32RepeatedType).
				AddField("sfixed64_field", resolver.Sfixed64Type).
				AddField("sfixed64s_field", resolver.Sfixed64RepeatedType).
				AddField("bool_field", resolver.BoolType).
				AddField("bools_field", resolver.BoolRepeatedType).
				AddField("string_field", resolver.StringType).
				AddField("strings_field", resolver.StringRepeatedType).
				AddField("byte_string_field", resolver.BytesType).
				AddField("byte_strings_field", resolver.BytesRepeatedType).
				AddField("enum_field", ref.Type(t, "org.federation", "ContentType")).
				AddField("enums_field", ref.RepeatedType(t, "org.federation", "ContentType")).
				AddField("env_field", resolver.StringType).
				AddField("envs_field", resolver.StringRepeatedType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("Content").
				AddFieldWithRule(
					"by_field",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "by_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "by_field")).Build(t),
				).
				AddFieldWithRule(
					"double_field",
					resolver.DoubleType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.DoubleType, resolver.DoubleType, "double_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "double_field")).Build(t),
				).
				AddFieldWithRule(
					"doubles_field",
					resolver.DoubleRepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.DoubleRepeatedType, resolver.DoubleRepeatedType, "doubles_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "doubles_field")).Build(t),
				).
				AddFieldWithRule(
					"float_field",
					resolver.FloatType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.FloatType, resolver.FloatType, "float_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "float_field")).Build(t),
				).
				AddFieldWithRule(
					"floats_field",
					resolver.FloatRepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.FloatRepeatedType, resolver.FloatRepeatedType, "floats_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "floats_field")).Build(t),
				).
				AddFieldWithRule(
					"int32_field",
					resolver.Int32Type,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Int32Type, resolver.Int32Type, "int32_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "int32_field")).Build(t),
				).
				AddFieldWithRule(
					"int32s_field",
					resolver.Int32RepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Int32RepeatedType, resolver.Int32RepeatedType, "int32s_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "int32s_field")).Build(t),
				).
				AddFieldWithRule(
					"int64_field",
					resolver.Int64Type,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Int64Type, resolver.Int64Type, "int64_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "int64_field")).Build(t),
				).
				AddFieldWithRule(
					"int64s_field",
					resolver.Int64RepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Int64RepeatedType, resolver.Int64RepeatedType, "int64s_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "int64s_field")).Build(t),
				).
				AddFieldWithRule(
					"uint32_field",
					resolver.Uint32Type,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Uint32Type, resolver.Uint32Type, "uint32_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "uint32_field")).Build(t),
				).
				AddFieldWithRule(
					"uint32s_field",
					resolver.Uint32RepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Uint32RepeatedType, resolver.Uint32RepeatedType, "uint32s_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "uint32s_field")).Build(t),
				).
				AddFieldWithRule(
					"uint64_field",
					resolver.Uint64Type,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Uint64Type, resolver.Uint64Type, "uint64_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "uint64_field")).Build(t),
				).
				AddFieldWithRule(
					"uint64s_field",
					resolver.Uint64RepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Uint64RepeatedType, resolver.Uint64RepeatedType, "uint64s_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "uint64s_field")).Build(t),
				).
				AddFieldWithRule(
					"sint32_field",
					resolver.Sint32Type,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Sint32Type, resolver.Sint32Type, "sint32_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "sint32_field")).Build(t),
				).
				AddFieldWithRule(
					"sint32s_field",
					resolver.Sint32RepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Sint32RepeatedType, resolver.Sint32RepeatedType, "sint32s_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "sint32s_field")).Build(t),
				).
				AddFieldWithRule(
					"sint64_field",
					resolver.Sint64Type,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Sint64Type, resolver.Sint64Type, "sint64_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "sint64_field")).Build(t),
				).
				AddFieldWithRule(
					"sint64s_field",
					resolver.Sint64RepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Sint64RepeatedType, resolver.Sint64RepeatedType, "sint64s_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "sint64s_field")).Build(t),
				).
				AddFieldWithRule(
					"fixed32_field",
					resolver.Fixed32Type,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Fixed32Type, resolver.Fixed32Type, "fixed32_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "fixed32_field")).Build(t),
				).
				AddFieldWithRule(
					"fixed32s_field",
					resolver.Fixed32RepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Fixed32RepeatedType, resolver.Fixed32RepeatedType, "fixed32s_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "fixed32s_field")).Build(t),
				).
				AddFieldWithRule(
					"fixed64_field",
					resolver.Fixed64Type,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Fixed64Type, resolver.Fixed64Type, "fixed64_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "fixed64_field")).Build(t),
				).
				AddFieldWithRule(
					"fixed64s_field",
					resolver.Fixed64RepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Fixed64RepeatedType, resolver.Fixed64RepeatedType, "fixed64s_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "fixed64s_field")).Build(t),
				).
				AddFieldWithRule(
					"sfixed32_field",
					resolver.Sfixed32Type,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Sfixed32Type, resolver.Sfixed32Type, "sfixed32_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "sfixed32_field")).Build(t),
				).
				AddFieldWithRule(
					"sfixed32s_field",
					resolver.Sfixed32RepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Sfixed32RepeatedType, resolver.Sfixed32RepeatedType, "sfixed32s_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "sfixed32s_field")).Build(t),
				).
				AddFieldWithRule(
					"sfixed64_field",
					resolver.Sfixed64Type,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Sfixed64Type, resolver.Sfixed64Type, "sfixed64_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "sfixed64_field")).Build(t),
				).
				AddFieldWithRule(
					"sfixed64s_field",
					resolver.Sfixed64RepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.Sfixed64RepeatedType, resolver.Sfixed64RepeatedType, "sfixed64s_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "sfixed64s_field")).Build(t),
				).
				AddFieldWithRule(
					"bool_field",
					resolver.BoolType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.BoolType, resolver.BoolType, "bool_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "bool_field")).Build(t),
				).
				AddFieldWithRule(
					"bools_field",
					resolver.BoolRepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.BoolRepeatedType, resolver.BoolRepeatedType, "bools_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "bools_field")).Build(t),
				).
				AddFieldWithRule(
					"string_field",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "string_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "string_field")).Build(t),
				).
				AddFieldWithRule(
					"strings_field",
					resolver.StringRepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.StringRepeatedType, resolver.StringRepeatedType, "strings_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "strings_field")).Build(t),
				).
				AddFieldWithRule(
					"byte_string_field",
					resolver.BytesType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.BytesType, resolver.BytesType, "byte_string_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "byte_string_field")).Build(t),
				).
				AddFieldWithRule(
					"byte_strings_field",
					resolver.BytesRepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.BytesRepeatedType, resolver.BytesRepeatedType, "byte_strings_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "byte_strings_field")).Build(t),
				).
				AddFieldWithRule(
					"enum_field",
					ref.Type(t, "org.federation", "ContentType"),
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(ref.Type(t, "org.federation", "ContentType"), ref.Type(t, "org.federation", "ContentType"), "enum_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "enum_field")).Build(t),
				).
				AddFieldWithRule(
					"enums_field",
					ref.RepeatedType(t, "org.federation", "ContentType"),
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(ref.RepeatedType(t, "org.federation", "ContentType"), ref.RepeatedType(t, "org.federation", "ContentType"), "enums_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "enums_field")).Build(t),
				).
				AddFieldWithRule(
					"env_field",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "env_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "env_field")).Build(t),
				).
				AddFieldWithRule(
					"envs_field",
					resolver.StringRepeatedType,
					testutil.NewFieldRuleBuilder(
						testutil.NewMessageArgumentValueBuilder(resolver.StringRepeatedType, resolver.StringRepeatedType, "envs_field").
							Build(t),
					).SetAlias(ref.Field(t, "content", "Content", "envs_field")).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetAlias(ref.Message(t, "content", "Content")).
						SetMessageArgument(ref.Message(t, "org.federation", "ContentArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetResponseArgument").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetResponse").
				AddFieldWithRule(
					"content",
					ref.Type(t, "org.federation", "Content"),
					testutil.NewFieldRuleBuilder(
						testutil.NewNameReferenceValueBuilder(ref.Type(t, "content", "GetContentResponse"), ref.Type(t, "content", "Content"), "content").
							Build(t),
					).Build(t),
				).
				AddFieldWithRule(
					"content2",
					ref.Type(t, "org.federation", "Content"),
					testutil.NewFieldRuleBuilder(
						testutil.NewNameReferenceValueBuilder(ref.Type(t, "org.federation", "Content"), ref.Type(t, "org.federation", "Content"), "content2").
							Build(t),
					).Build(t),
				).
				AddFieldWithRule(
					"cel_expr",
					resolver.Int64Type,
					testutil.NewFieldRuleBuilder(
						testutil.NewNameReferenceValueBuilder(resolver.Int64Type, resolver.Int64Type, "content.int32_field + content.sint32_field + content2.int64_field + content2.sint64_field").
							Build(t),
					).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("res").
								SetUsed(true).
								SetCall(
									testutil.NewCallExprBuilder().
										SetMethod(ref.Method(t, "content", "ContentService", "GetContent")).
										SetRequest(
											testutil.NewRequestBuilder().
												AddField("by_field", resolver.StringType, testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
												AddField("double_field", resolver.DoubleType, resolver.NewDoubleValue(1.23)).
												AddField("doubles_field", resolver.DoubleRepeatedType, resolver.NewDoublesValue(4.56, 7.89)).
												AddField("float_field", resolver.FloatType, resolver.NewFloatValue(4.56)).
												AddField("floats_field", resolver.FloatRepeatedType, resolver.NewFloatsValue(7.89, 1.23)).
												AddField("int32_field", resolver.Int32Type, resolver.NewInt32Value(-1)).
												AddField("int32s_field", resolver.Int32RepeatedType, resolver.NewInt32sValue(-2, -3)).
												AddField("int64_field", resolver.Int64Type, resolver.NewInt64Value(-4)).
												AddField("int64s_field", resolver.Int64RepeatedType, resolver.NewInt64sValue(-5, -6)).
												AddField("uint32_field", resolver.Uint32Type, resolver.NewUint32Value(1)).
												AddField("uint32s_field", resolver.Uint32RepeatedType, resolver.NewUint32sValue(2, 3)).
												AddField("uint64_field", resolver.Uint64Type, resolver.NewUint64Value(4)).
												AddField("uint64s_field", resolver.Uint64RepeatedType, resolver.NewUint64sValue(5, 6)).
												AddField("sint32_field", resolver.Sint32Type, resolver.NewSint32Value(-7)).
												AddField("sint32s_field", resolver.Sint32RepeatedType, resolver.NewSint32sValue(-8, -9)).
												AddField("sint64_field", resolver.Sint64Type, resolver.NewSint64Value(-10)).
												AddField("sint64s_field", resolver.Sint64RepeatedType, resolver.NewSint64sValue(-11, -12)).
												AddField("fixed32_field", resolver.Fixed32Type, resolver.NewFixed32Value(10)).
												AddField("fixed32s_field", resolver.Fixed32RepeatedType, resolver.NewFixed32sValue(11, 12)).
												AddField("fixed64_field", resolver.Fixed64Type, resolver.NewFixed64Value(13)).
												AddField("fixed64s_field", resolver.Fixed64RepeatedType, resolver.NewFixed64sValue(14, 15)).
												AddField("sfixed32_field", resolver.Sfixed32Type, resolver.NewSfixed32Value(-14)).
												AddField("sfixed32s_field", resolver.Sfixed32RepeatedType, resolver.NewSfixed32sValue(-15, -16)).
												AddField("sfixed64_field", resolver.Sfixed64Type, resolver.NewSfixed64Value(-17)).
												AddField("sfixed64s_field", resolver.Sfixed64RepeatedType, resolver.NewSfixed64sValue(-18, -19)).
												AddField("bool_field", resolver.BoolType, resolver.NewBoolValue(true)).
												AddField("bools_field", resolver.BoolRepeatedType, resolver.NewBoolsValue(true, false)).
												AddField("string_field", resolver.StringType, resolver.NewStringValue("foo")).
												AddField("strings_field", resolver.StringRepeatedType, resolver.NewStringsValue("hello", "world")).
												AddField("byte_string_field", resolver.BytesType, resolver.NewByteStringValue([]byte("foo"))).
												AddField("byte_strings_field", resolver.BytesRepeatedType, resolver.NewByteStringsValue([]byte("foo"), []byte("bar"))).
												AddField("enum_field", ref.Type(t, "content", "ContentType"), resolver.NewEnumValue(ref.EnumValue(t, "content", "ContentType", "CONTENT_TYPE_1"))).
												AddField(
													"enums_field",
													ref.RepeatedType(t, "content", "ContentType"),
													resolver.NewEnumsValue(
														ref.EnumValue(t, "content", "ContentType", "CONTENT_TYPE_2"),
														ref.EnumValue(t, "content", "ContentType", "CONTENT_TYPE_3"),
													),
												).
												AddField("env_field", resolver.StringType, resolver.NewEnvValue("foo")).
												AddField("envs_field", resolver.StringRepeatedType, resolver.NewEnvsValue("foo", "bar")).
												AddField(
													"message_field",
													ref.Type(t, "content", "Content"),
													resolver.NewMessageValue(
														ref.Type(t, "content", "Content"),
														map[string]*resolver.Value{
															"double_field":       resolver.NewDoubleValue(1.23),
															"doubles_field":      resolver.NewDoublesValue(4.56, 7.89),
															"float_field":        resolver.NewFloatValue(4.56),
															"floats_field":       resolver.NewFloatsValue(7.89, 1.23),
															"int32_field":        resolver.NewInt32Value(-1),
															"int32s_field":       resolver.NewInt32sValue(-2, -3),
															"int64_field":        resolver.NewInt64Value(-4),
															"int64s_field":       resolver.NewInt64sValue(-5, -6),
															"uint32_field":       resolver.NewUint32Value(1),
															"uint32s_field":      resolver.NewUint32sValue(2, 3),
															"uint64_field":       resolver.NewUint64Value(4),
															"uint64s_field":      resolver.NewUint64sValue(5, 6),
															"sint32_field":       resolver.NewSint32Value(-7),
															"sint32s_field":      resolver.NewSint32sValue(-8, -9),
															"sint64_field":       resolver.NewSint64Value(-10),
															"sint64s_field":      resolver.NewSint64sValue(-11, -12),
															"fixed32_field":      resolver.NewFixed32Value(10),
															"fixed32s_field":     resolver.NewFixed32sValue(11, 12),
															"fixed64_field":      resolver.NewFixed64Value(13),
															"fixed64s_field":     resolver.NewFixed64sValue(14, 15),
															"sfixed32_field":     resolver.NewSfixed32Value(-14),
															"sfixed32s_field":    resolver.NewSfixed32sValue(-15, -16),
															"sfixed64_field":     resolver.NewSfixed64Value(-17),
															"sfixed64s_field":    resolver.NewSfixed64sValue(-18, -19),
															"bool_field":         resolver.NewBoolValue(true),
															"bools_field":        resolver.NewBoolsValue(true, false),
															"string_field":       resolver.NewStringValue("foo"),
															"strings_field":      resolver.NewStringsValue("hello", "world"),
															"byte_string_field":  resolver.NewByteStringValue([]byte("foo")),
															"byte_strings_field": resolver.NewByteStringsValue([]byte("foo"), []byte("bar")),
															"enum_field":         resolver.NewEnumValue(ref.EnumValue(t, "content", "ContentType", "CONTENT_TYPE_1")),
															"enums_field": resolver.NewEnumsValue(
																ref.EnumValue(t, "content", "ContentType", "CONTENT_TYPE_2"),
																ref.EnumValue(t, "content", "ContentType", "CONTENT_TYPE_3"),
															),
															"env_field":      resolver.NewEnvValue("foo"),
															"envs_field":     resolver.NewEnvsValue("foo", "bar"),
															"message_field":  resolver.NewMessageValue(ref.Type(t, "content", "Content"), map[string]*resolver.Value{}),
															"messages_field": resolver.NewMessagesValue(ref.RepeatedType(t, "content", "Content"), map[string]*resolver.Value{}, map[string]*resolver.Value{}),
														},
													),
												).
												AddField(
													"messages_field",
													ref.RepeatedType(t, "content", "Content"),
													resolver.NewMessagesValue(ref.RepeatedType(t, "content", "Content"), map[string]*resolver.Value{}, map[string]*resolver.Value{})).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("content").
								SetUsed(true).
								SetBy(testutil.NewCELValueBuilder("res.content", ref.Type(t, "content", "Content")).Build(t)).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("content2").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "Content")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("by_field", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
												Add("double_field", resolver.NewDoubleValue(1.23)).
												Add("doubles_field", resolver.NewDoublesValue(4.56, 7.89)).
												Add("float_field", resolver.NewFloatValue(4.56)).
												Add("floats_field", resolver.NewFloatsValue(7.89, 1.23)).
												Add("int32_field", resolver.NewInt32Value(-1)).
												Add("int32s_field", resolver.NewInt32sValue(-2, -3)).
												Add("int64_field", resolver.NewInt64Value(-4)).
												Add("int64s_field", resolver.NewInt64sValue(-5, -6)).
												Add("uint32_field", resolver.NewUint32Value(1)).
												Add("uint32s_field", resolver.NewUint32sValue(2, 3)).
												Add("uint64_field", resolver.NewUint64Value(4)).
												Add("uint64s_field", resolver.NewUint64sValue(5, 6)).
												Add("sint32_field", resolver.NewSint32Value(-7)).
												Add("sint32s_field", resolver.NewSint32sValue(-8, -9)).
												Add("sint64_field", resolver.NewSint64Value(-10)).
												Add("sint64s_field", resolver.NewSint64sValue(-11, -12)).
												Add("fixed32_field", resolver.NewFixed32Value(10)).
												Add("fixed32s_field", resolver.NewFixed32sValue(11, 12)).
												Add("fixed64_field", resolver.NewFixed64Value(13)).
												Add("fixed64s_field", resolver.NewFixed64sValue(14, 15)).
												Add("sfixed32_field", resolver.NewSfixed32Value(-14)).
												Add("sfixed32s_field", resolver.NewSfixed32sValue(-15, -16)).
												Add("sfixed64_field", resolver.NewSfixed64Value(-17)).
												Add("sfixed64s_field", resolver.NewSfixed64sValue(-18, -19)).
												Add("bool_field", resolver.NewBoolValue(true)).
												Add("bools_field", resolver.NewBoolsValue(true, false)).
												Add("string_field", resolver.NewStringValue("foo")).
												Add("strings_field", resolver.NewStringsValue("hello", "world")).
												Add("byte_string_field", resolver.NewByteStringValue([]byte("foo"))).
												Add("byte_strings_field", resolver.NewByteStringsValue([]byte("foo"), []byte("bar"))).
												Add("enum_field", resolver.NewEnumValue(ref.EnumValue(t, "org.federation", "ContentType", "CONTENT_TYPE_1"))).
												Add(
													"enums_field",
													resolver.NewEnumsValue(
														ref.EnumValue(t, "org.federation", "ContentType", "CONTENT_TYPE_2"),
														ref.EnumValue(t, "org.federation", "ContentType", "CONTENT_TYPE_3"),
													),
												).
												Add("env_field", resolver.NewEnvValue("foo")).
												Add("envs_field", resolver.NewEnvsValue("foo", "bar")).
												Add(
													"message_field",
													resolver.NewMessageValue(
														ref.Type(t, "org.federation", "Content"),
														map[string]*resolver.Value{
															"double_field":       resolver.NewDoubleValue(1.23),
															"doubles_field":      resolver.NewDoublesValue(4.56, 7.89),
															"float_field":        resolver.NewFloatValue(4.56),
															"floats_field":       resolver.NewFloatsValue(7.89, 1.23),
															"int32_field":        resolver.NewInt32Value(-1),
															"int32s_field":       resolver.NewInt32sValue(-2, -3),
															"int64_field":        resolver.NewInt64Value(-4),
															"int64s_field":       resolver.NewInt64sValue(-5, -6),
															"uint32_field":       resolver.NewUint32Value(1),
															"uint32s_field":      resolver.NewUint32sValue(2, 3),
															"uint64_field":       resolver.NewUint64Value(4),
															"uint64s_field":      resolver.NewUint64sValue(5, 6),
															"sint32_field":       resolver.NewSint32Value(-7),
															"sint32s_field":      resolver.NewSint32sValue(-8, -9),
															"sint64_field":       resolver.NewSint64Value(-10),
															"sint64s_field":      resolver.NewSint64sValue(-11, -12),
															"fixed32_field":      resolver.NewFixed32Value(10),
															"fixed32s_field":     resolver.NewFixed32sValue(11, 12),
															"fixed64_field":      resolver.NewFixed64Value(13),
															"fixed64s_field":     resolver.NewFixed64sValue(14, 15),
															"sfixed32_field":     resolver.NewSfixed32Value(-14),
															"sfixed32s_field":    resolver.NewSfixed32sValue(-15, -16),
															"sfixed64_field":     resolver.NewSfixed64Value(-17),
															"sfixed64s_field":    resolver.NewSfixed64sValue(-18, -19),
															"bool_field":         resolver.NewBoolValue(true),
															"bools_field":        resolver.NewBoolsValue(true, false),
															"string_field":       resolver.NewStringValue("foo"),
															"strings_field":      resolver.NewStringsValue("hello", "world"),
															"byte_string_field":  resolver.NewByteStringValue([]byte("foo")),
															"byte_strings_field": resolver.NewByteStringsValue([]byte("foo"), []byte("bar")),
															"enum_field":         resolver.NewEnumValue(ref.EnumValue(t, "org.federation", "ContentType", "CONTENT_TYPE_1")),
															"enums_field": resolver.NewEnumsValue(
																ref.EnumValue(t, "org.federation", "ContentType", "CONTENT_TYPE_2"),
																ref.EnumValue(t, "org.federation", "ContentType", "CONTENT_TYPE_3"),
															),
															"env_field":      resolver.NewEnvValue("foo"),
															"envs_field":     resolver.NewEnvsValue("foo", "bar"),
															"message_field":  resolver.NewMessageValue(ref.Type(t, "org.federation", "Content"), map[string]*resolver.Value{}),
															"messages_field": resolver.NewMessagesValue(ref.RepeatedType(t, "org.federation", "Content"), map[string]*resolver.Value{}, map[string]*resolver.Value{}),
														},
													),
												).
												Add(
													"messages_field",
													resolver.NewMessagesValue(ref.RepeatedType(t, "org.federation", "Content"), map[string]*resolver.Value{}, map[string]*resolver.Value{})).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "GetResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "content", "GetContentResponse")).
								Add(ref.Message(t, "org.federation", "Content")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(testutil.NewVariableDefinitionGroupByName("res")).
								SetEnd(testutil.NewVariableDefinition("content")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("content2")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetRequest").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddService(
			testutil.NewServiceBuilder("FederationService").
				AddMethod("Get", ref.Message(t, "org.federation", "GetRequest"), ref.Message(t, "org.federation", "GetResponse"), nil).
				SetRule(
					testutil.NewServiceRuleBuilder().
						AddDependency("", ref.Service(t, "content", "ContentService")).
						Build(t),
				).
				AddMessage(ref.Message(t, "org.federation", "Content"), ref.Message(t, "org.federation", "ContentArgument")).
				AddMessage(ref.Message(t, "org.federation", "GetResponse"), ref.Message(t, "org.federation", "GetResponseArgument")).
				Build(t),
		)

	content := ref.Message(t, "org.federation", "Content")
	content.Fields = append(
		content.Fields,
		&resolver.Field{
			Name: "message_field",
			Type: resolver.NewMessageType(content, false),
			Rule: testutil.NewFieldRuleBuilder(
				testutil.NewMessageArgumentValueBuilder(resolver.NewMessageType(content, false), resolver.NewMessageType(content, false), "message_field").
					Build(t),
			).SetAlias(ref.Field(t, "content", "Content", "message_field")).Build(t),
		},
		&resolver.Field{
			Name: "messages_field",
			Type: ref.RepeatedType(t, "org.federation", "Content"),
			Rule: testutil.NewFieldRuleBuilder(
				testutil.NewMessageArgumentValueBuilder(ref.RepeatedType(t, "org.federation", "Content"), ref.RepeatedType(t, "org.federation", "Content"), "messages_field").
					Build(t),
			).SetAlias(ref.Field(t, "content", "Content", "messages_field")).Build(t),
		},
	)

	contentArg := ref.Message(t, "org.federation", "ContentArgument")
	contentArg.Fields = append(
		contentArg.Fields,
		&resolver.Field{
			Name: "message_field",
			Type: resolver.NewMessageType(content, false),
		},
		&resolver.Field{
			Name: "messages_field",
			Type: resolver.NewMessageType(content, true),
		},
	)

	federationFile := fb.Build(t)
	federationService := federationFile.Services[0]

	if diff := cmp.Diff(result.Files[0].Services[0], federationService, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestMultiUser(t *testing.T) {
	fileName := filepath.Join(testutil.RepoRoot(), "testdata", "multi_user.proto")
	fb := testutil.NewFileBuilder(fileName)
	ref := testutil.NewBuilderReferenceManager(getUserProtoBuilder(t), fb)

	fb.SetPackage("org.federation").
		SetGoPackage("example/federation", "federation").
		AddMessage(testutil.NewMessageBuilder("SubArgument").Build(t)).
		AddMessage(testutil.NewMessageBuilder("UserIDArgument").Build(t)).
		AddMessage(
			testutil.NewMessageBuilder("Sub").
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "SubArgument")).
						SetCustomResolver(true).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("UserID").
				AddFieldWithRule("value", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("xxx")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "UserIDArgument")).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("_def0").
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "Sub")).
										Build(t),
								).
								Build(t),
						).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "Sub")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("_def0")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("UserArgument").
				AddField("user_id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("User").
				AddFieldWithAutoBind("id", resolver.StringType, ref.Field(t, "org.user", "User", "id")).
				AddFieldWithRule("name", resolver.StringType, testutil.NewFieldRuleBuilder(nil).SetCustomResolver(true).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("res").
								SetUsed(true).
								SetCall(
									testutil.NewCallExprBuilder().
										SetMethod(ref.Method(t, "org.user", "UserService", "GetUser")).
										SetRequest(
											testutil.NewRequestBuilder().
												AddField("id", resolver.StringType, testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "user_id").Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("user").
								SetUsed(true).
								SetAutoBind(true).
								SetBy(testutil.NewCELValueBuilder("res.user", ref.Type(t, "org.user", "User")).Build(t)).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("_def2").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "Sub")).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "UserArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "Sub")).
								Add(ref.Message(t, "org.user", "GetUserResponse")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("_def2")).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(testutil.NewVariableDefinitionGroupByName("res")).
								SetEnd(testutil.NewVariableDefinition("user")).
								Build(t),
						).
						Build(t),
				).
				Build(t),
		).
		AddMessage(testutil.NewMessageBuilder("GetRequest").Build(t)).
		AddMessage(testutil.NewMessageBuilder("GetResponseArgument").Build(t)).
		AddMessage(
			testutil.NewMessageBuilder("GetResponse").
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
				AddFieldWithRule(
					"user2",
					ref.Type(t, "org.federation", "User"),
					testutil.NewFieldRuleBuilder(
						testutil.NewNameReferenceValueBuilder(
							ref.Type(t, "org.federation", "User"),
							ref.Type(t, "org.federation", "User"),
							"user2",
						).Build(t),
					).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("uid").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "UserID")).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("user").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "User")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("user_id", resolver.NewStringValue("1")).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("user2").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "User")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("user_id", resolver.NewStringValue("2")).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "GetResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "UserID")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(testutil.NewVariableDefinitionGroupByName("uid")).
								SetEnd(testutil.NewVariableDefinition("user")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(testutil.NewVariableDefinitionGroupByName("uid")).
								SetEnd(testutil.NewVariableDefinition("user2")).
								Build(t),
						).
						Build(t),
				).
				Build(t),
		).
		AddService(
			testutil.NewServiceBuilder("FederationService").
				AddMethod("Get", ref.Message(t, "org.federation", "GetRequest"), ref.Message(t, "org.federation", "GetResponse"), nil).
				SetRule(
					testutil.NewServiceRuleBuilder().
						AddDependency("", ref.Service(t, "org.user", "UserService")).
						Build(t),
				).
				AddMessage(ref.Message(t, "org.federation", "GetResponse"), ref.Message(t, "org.federation", "GetResponseArgument")).
				AddMessage(ref.Message(t, "org.federation", "Sub"), ref.Message(t, "org.federation", "SubArgument")).
				AddMessage(ref.Message(t, "org.federation", "User"), ref.Message(t, "org.federation", "UserArgument")).
				AddMessage(ref.Message(t, "org.federation", "UserID"), ref.Message(t, "org.federation", "UserIDArgument")).
				Build(t),
		)

	federationFile := fb.Build(t)
	federationService := federationFile.Services[0]

	r := resolver.New(testutil.Compile(t, fileName))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("faield to get files. expected 1 but got %d", len(result.Files))
	}
	if len(result.Files[0].Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Files[0].Services))
	}
	if diff := cmp.Diff(result.Files[0].Services[0], federationService, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestOneof(t *testing.T) {
	fileName := filepath.Join(testutil.RepoRoot(), "testdata", "oneof.proto")
	fb := testutil.NewFileBuilder(fileName)
	ref := testutil.NewBuilderReferenceManager(getUserProtoBuilder(t), fb)

	fb.SetPackage("org.federation").
		SetGoPackage("example/federation", "federation").
		AddMessage(
			testutil.NewMessageBuilder("UserArgument").
				AddField("user_id", resolver.StringType).
				Build(t),
		).
		AddMessage(testutil.NewMessageBuilder("MArgument").Build(t)).
		AddMessage(
			testutil.NewMessageBuilder("UserSelectionArgument").
				AddField("value", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("M").
				AddFieldWithRule("value", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("foo")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "MArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("User").
				AddFieldWithRule(
					"id",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "user_id").Build(t)).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("_def0").
								SetCall(
									testutil.NewCallExprBuilder().
										SetMethod(ref.Method(t, "org.user", "UserService", "GetUser")).
										SetRequest(
											testutil.NewRequestBuilder().
												AddField("id", resolver.StringType, testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "user_id").Build(t)).
												Build(t),
										).
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
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("_def0")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("UserSelection").
				AddFieldWithRule(
					"user_a",
					ref.Type(t, "org.federation", "User"),
					testutil.NewFieldRuleBuilder(nil).
						SetOneof(
							testutil.NewFieldOneofRuleBuilder().
								SetIf("m.value == $.value", resolver.BoolType).
								AddVariableDefinition(
									testutil.NewVariableDefinitionBuilder().
										SetName("ua").
										SetUsed(true).
										SetMessage(
											testutil.NewMessageExprBuilder().
												SetMessage(ref.Message(t, "org.federation", "User")).
												SetArgs(testutil.NewMessageDependencyArgumentBuilder().
													Add("user_id", resolver.NewStringValue("a")).
													Build(t),
												).
												Build(t),
										).
										Build(t),
								).
								SetBy("ua", ref.Type(t, "org.federation", "User")).
								SetDependencyGraph(
									testutil.NewDependencyGraphBuilder().
										Add(ref.Message(t, "org.federation", "User")).
										Build(t),
								).
								AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("ua")).
								Build(t),
						).
						Build(t),
				).
				AddFieldWithRule(
					"user_b",
					ref.Type(t, "org.federation", "User"),
					testutil.NewFieldRuleBuilder(nil).
						SetOneof(
							testutil.NewFieldOneofRuleBuilder().
								SetIf("m.value != $.value", resolver.BoolType).
								AddVariableDefinition(
									testutil.NewVariableDefinitionBuilder().
										SetName("ub").
										SetUsed(true).
										SetMessage(
											testutil.NewMessageExprBuilder().
												SetMessage(ref.Message(t, "org.federation", "User")).
												SetArgs(testutil.NewMessageDependencyArgumentBuilder().
													Add("user_id", resolver.NewStringValue("b")).
													Build(t),
												).
												Build(t),
										).
										Build(t),
								).
								SetBy("ub", ref.Type(t, "org.federation", "User")).
								SetDependencyGraph(
									testutil.NewDependencyGraphBuilder().
										Add(ref.Message(t, "org.federation", "User")).
										Build(t),
								).
								AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("ub")).
								Build(t),
						).
						Build(t),
				).
				AddFieldWithRule(
					"user_c",
					ref.Type(t, "org.federation", "User"),
					testutil.NewFieldRuleBuilder(nil).
						SetOneof(
							testutil.NewFieldOneofRuleBuilder().
								SetDefault(true).
								AddVariableDefinition(
									testutil.NewVariableDefinitionBuilder().
										SetName("uc").
										SetUsed(true).
										SetMessage(
											testutil.NewMessageExprBuilder().
												SetMessage(ref.Message(t, "org.federation", "User")).
												SetArgs(testutil.NewMessageDependencyArgumentBuilder().
													Add("user_id", resolver.NewStringValue("c")).
													Build(t),
												).
												Build(t),
										).
										Build(t),
								).
								SetBy("uc", ref.Type(t, "org.federation", "User")).
								SetDependencyGraph(
									testutil.NewDependencyGraphBuilder().
										Add(ref.Message(t, "org.federation", "User")).
										Build(t),
								).
								AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("uc")).
								Build(t),
						).
						Build(t),
				).
				AddOneof(testutil.NewOneofBuilder("user").AddFieldNames("user_a", "user_b", "user_c").Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("m").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "M")).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "UserSelectionArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "M")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("m")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(testutil.NewMessageBuilder("GetRequest").Build(t)).
		AddMessage(testutil.NewMessageBuilder("GetResponseArgument").Build(t)).
		AddMessage(
			testutil.NewMessageBuilder("GetResponse").
				AddFieldWithRule(
					"user",
					ref.Type(t, "org.federation", "User"),
					testutil.NewFieldRuleBuilder(
						testutil.NewNameReferenceValueBuilder(
							ref.Type(t, "org.federation", "User"),
							ref.Type(t, "org.federation", "User"),
							"sel.user",
						).Build(t),
					).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("sel").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "UserSelection")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("value", resolver.NewStringValue("foo")).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "GetResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "UserSelection")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("sel")).
						Build(t),
				).
				Build(t),
		).
		AddService(
			testutil.NewServiceBuilder("FederationService").
				AddMethod("Get", ref.Message(t, "org.federation", "GetRequest"), ref.Message(t, "org.federation", "GetResponse"), nil).
				SetRule(testutil.NewServiceRuleBuilder().Build(t)).
				AddMessage(ref.Message(t, "org.federation", "GetResponse"), ref.Message(t, "org.federation", "GetResponseArgument")).
				AddMessage(ref.Message(t, "org.federation", "M"), ref.Message(t, "org.federation", "MArgument")).
				AddMessage(ref.Message(t, "org.federation", "User"), ref.Message(t, "org.federation", "UserArgument")).
				AddMessage(ref.Message(t, "org.federation", "UserSelection"), ref.Message(t, "org.federation", "UserSelectionArgument")).
				Build(t),
		)

	federationFile := fb.Build(t)
	federationService := federationFile.Services[0]

	r := resolver.New(testutil.Compile(t, fileName))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("faield to get files. expected 1 but got %d", len(result.Files))
	}
	if len(result.Files[0].Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Files[0].Services))
	}
	if diff := cmp.Diff(result.Files[0].Services[0], federationService, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestValidation(t *testing.T) {
	fileName := filepath.Join(testutil.RepoRoot(), "testdata", "validation.proto")
	r := resolver.New(testutil.Compile(t, fileName))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("faield to get files. expected 1 but got %d", len(result.Files))
	}
	if len(result.Files[0].Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Files[0].Services))
	}

	fb := testutil.NewFileBuilder(fileName)
	ref := testutil.NewBuilderReferenceManager(fb)

	fb.SetPackage("org.federation").
		SetGoPackage("example/federation", "federation").
		AddMessage(
			testutil.NewMessageBuilder("PostArgument").Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("Post").
				AddFieldWithRule("id", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("some-id")).Build(t)).
				AddFieldWithRule("title", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("some-title")).Build(t)).
				AddFieldWithRule("content", resolver.StringType, testutil.NewFieldRuleBuilder(resolver.NewStringValue("some-content")).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().SetMessageArgument(ref.Message(t, "org.federation", "PostArgument")).Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("CustomMessageArgument").
				AddField("message", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("CustomMessage").
				AddFieldWithRule("message", resolver.StringType, testutil.NewFieldRuleBuilder(&resolver.Value{CEL: testutil.NewCELValueBuilder("$.message", resolver.StringType).Build(t)}).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().SetMessageArgument(ref.Message(t, "org.federation", "CustomMessageArgument")).Build(t),
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
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("post").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "Post")).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("_def1").
								SetValidation(
									testutil.NewValidationExprBuilder().
										SetError(
											testutil.NewGRPCErrorBuilder().
												SetCode(code.Code_FAILED_PRECONDITION).
												SetMessage("validation message 1").
												SetIf("post.id != 'some-id'").
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("_def2").
								SetValidation(
									testutil.NewValidationExprBuilder().
										SetError(
											testutil.NewGRPCErrorBuilder().
												SetCode(code.Code_FAILED_PRECONDITION).
												SetMessage("validation message 2").
												AddDetail(
													testutil.NewGRPCErrorDetailBuilder().
														SetIf("post.title != 'some-title'").
														AddMessage(
															testutil.NewVariableDefinitionBuilder().
																SetName("_def2_err_detail0_msg0").
																SetUsed(true).
																SetMessage(
																	testutil.NewMessageExprBuilder().
																		SetMessage(ref.Message(t, "org.federation", "CustomMessage")).
																		SetArgs(
																			testutil.NewMessageDependencyArgumentBuilder().
																				Add("message", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "message").Build(t)).
																				Build(t),
																		).
																		Build(t),
																).
																Build(t),
														).
														AddMessage(
															testutil.NewVariableDefinitionBuilder().
																SetName("_def2_err_detail0_msg1").
																SetUsed(true).
																SetIdx(1).
																SetMessage(
																	testutil.NewMessageExprBuilder().
																		SetMessage(ref.Message(t, "org.federation", "CustomMessage")).
																		SetArgs(
																			testutil.NewMessageDependencyArgumentBuilder().
																				Add("message", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "message").Build(t)).
																				Build(t),
																		).
																		Build(t),
																).
																Build(t),
														).
														AddPreconditionFailure(&resolver.PreconditionFailure{
															Violations: []*resolver.PreconditionFailureViolation{
																{
																	Type:        testutil.NewCELValueBuilder("'some-type'", resolver.StringType).Build(t),
																	Subject:     testutil.NewCELValueBuilder("'some-subject'", resolver.StringType).Build(t),
																	Description: testutil.NewCELValueBuilder("'some-description'", resolver.StringType).Build(t),
																},
															},
														}).
														AddBadRequest(&resolver.BadRequest{
															FieldViolations: []*resolver.BadRequestFieldViolation{
																{
																	Field:       testutil.NewCELValueBuilder("'some-field'", resolver.StringType).Build(t),
																	Description: testutil.NewCELValueBuilder("'some-description'", resolver.StringType).Build(t),
																},
															},
														}).
														AddLocalizedMessage(&resolver.LocalizedMessage{
															Locale:  "en-US",
															Message: testutil.NewCELValueBuilder("'some-message'", resolver.StringType).Build(t),
														}).
														Build(t),
												).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "GetPostResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "Post")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(testutil.NewVariableDefinitionGroupByName("post")).
								SetEnd(testutil.NewVariableDefinition("_def1")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(testutil.NewVariableDefinitionGroupByName("post")).
								SetEnd(testutil.NewVariableDefinition("_def2")).
								Build(t),
						).
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
					nil,
				).
				SetRule(
					testutil.NewServiceRuleBuilder().Build(t),
				).
				AddMessage(ref.Message(t, "org.federation", "CustomMessage"), ref.Message(t, "org.federation", "CustomMessageArgument")).
				AddMessage(ref.Message(t, "org.federation", "GetPostResponse"), ref.Message(t, "org.federation", "GetPostResponseArgument")).
				AddMessage(ref.Message(t, "org.federation", "Post"), ref.Message(t, "org.federation", "PostArgument")).
				Build(t),
		)
	federationFile := fb.Build(t)
	service := federationFile.Services[0]
	if diff := cmp.Diff(result.Files[0].Services[0], service, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestMap(t *testing.T) {
	fileName := filepath.Join(testutil.RepoRoot(), "testdata", "map.proto")
	fb := testutil.NewFileBuilder(fileName)
	ref := testutil.NewBuilderReferenceManager(getUserProtoBuilder(t), getPostProtoBuilder(t), fb)

	fb.SetPackage("org.federation").
		SetGoPackage("example/federation", "federation").
		AddMessage(
			testutil.NewMessageBuilder("PostsArgument").
				AddField("post_ids", resolver.StringRepeatedType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("UserArgument").
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
					testutil.NewFieldRuleBuilder(nil).SetMessageCustomResolver(true).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("res").
								SetUsed(true).
								SetCall(
									testutil.NewCallExprBuilder().
										SetMethod(ref.Method(t, "org.user", "UserService", "GetUser")).
										SetRequest(
											testutil.NewRequestBuilder().
												AddField(
													"id",
													resolver.StringType,
													testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "user_id").Build(t),
												).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("user").
								SetUsed(true).
								SetAutoBind(true).
								SetBy(testutil.NewCELValueBuilder("res.user", ref.Type(t, "org.user", "User")).Build(t)).
								Build(t),
						).
						SetCustomResolver(true).
						SetMessageArgument(ref.Message(t, "org.federation", "UserArgument")).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(testutil.NewVariableDefinitionGroupByName("res")).
								SetEnd(testutil.NewVariableDefinition("user")).
								Build(t),
						).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.user", "GetUserResponse")).
								Build(t),
						).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("Posts").
				AddFieldWithRule(
					"ids",
					resolver.StringRepeatedType,
					testutil.NewFieldRuleBuilder(resolver.NewByValue("ids", resolver.StringRepeatedType)).Build(t),
				).
				AddFieldWithRule(
					"titles",
					resolver.StringRepeatedType,
					testutil.NewFieldRuleBuilder(resolver.NewByValue("posts.map(post, post.title)", resolver.StringRepeatedType)).Build(t),
				).
				AddFieldWithRule(
					"contents",
					resolver.StringRepeatedType,
					testutil.NewFieldRuleBuilder(resolver.NewByValue("posts.map(post, post.content)", resolver.StringRepeatedType)).Build(t),
				).
				AddFieldWithRule(
					"users",
					ref.RepeatedType(t, "org.federation", "User"),
					testutil.NewFieldRuleBuilder(resolver.NewByValue("users", ref.RepeatedType(t, "org.user", "User"))).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("res").
								SetUsed(true).
								SetCall(
									testutil.NewCallExprBuilder().
										SetMethod(ref.Method(t, "org.post", "PostService", "GetPosts")).
										SetRequest(
											testutil.NewRequestBuilder().
												AddField(
													"ids",
													resolver.StringRepeatedType,
													resolver.NewByValue("$.post_ids", resolver.StringRepeatedType),
												).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("posts").
								SetUsed(true).
								SetBy(testutil.NewCELValueBuilder("res.posts", ref.RepeatedType(t, "org.post", "Post")).Build(t)).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("ids").
								SetUsed(true).
								SetMap(
									testutil.NewMapExprBuilder().
										SetIterator(
											testutil.NewIteratorBuilder().
												SetName("post").
												SetSource("posts").
												Build(t),
										).
										SetExpr(
											testutil.NewMapIteratorExprBuilder().
												SetBy(testutil.NewCELValueBuilder("post.id", resolver.StringType).Build(t)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("users").
								SetUsed(true).
								SetMap(
									testutil.NewMapExprBuilder().
										SetIterator(
											testutil.NewIteratorBuilder().
												SetName("iter").
												SetSource("posts").
												Build(t),
										).
										SetExpr(
											testutil.NewMapIteratorExprBuilder().
												SetMessage(
													testutil.NewMessageExprBuilder().
														SetMessage(ref.Message(t, "org.federation", "User")).
														SetArgs(
															testutil.NewMessageDependencyArgumentBuilder().
																Add("user_id", resolver.NewByValue("iter.user_id", resolver.StringType)).
																Build(t),
														).
														Build(t),
												).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "PostsArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.post", "GetPostsResponse"), ref.Message(t, "org.federation", "User")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(
									testutil.NewVariableDefinitionGroupBuilder().
										AddStart(testutil.NewVariableDefinitionGroupByName("res")).
										SetEnd(testutil.NewVariableDefinition("posts")).
										Build(t),
								).
								SetEnd(testutil.NewVariableDefinition("ids")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(
									testutil.NewVariableDefinitionGroupBuilder().
										AddStart(testutil.NewVariableDefinitionGroupByName("res")).
										SetEnd(testutil.NewVariableDefinition("posts")).
										Build(t),
								).
								SetEnd(testutil.NewVariableDefinition("users")).
								Build(t),
						).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostsRequest").
				AddField("ids", resolver.StringRepeatedType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostsResponseArgument").
				AddField("ids", resolver.StringRepeatedType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostsResponse").
				AddFieldWithRule(
					"posts",
					ref.Type(t, "org.federation", "Posts"),
					testutil.NewFieldRuleBuilder(resolver.NewByValue("posts", ref.Type(t, "org.federation", "Posts"))).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("posts").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "Posts")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("post_ids", resolver.NewByValue("$.ids", resolver.StringRepeatedType)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "GetPostsResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "Posts")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("posts")).
						Build(t),
				).
				Build(t),
		).
		AddService(
			testutil.NewServiceBuilder("FederationService").
				AddMethod(
					"GetPosts",
					ref.Message(t, "org.federation", "GetPostsRequest"),
					ref.Message(t, "org.federation", "GetPostsResponse"),
					nil,
				).
				SetRule(testutil.NewServiceRuleBuilder().Build(t)).
				AddMessage(ref.Message(t, "org.federation", "GetPostsResponse"), ref.Message(t, "org.federation", "GetPostsResponseArgument")).
				AddMessage(ref.Message(t, "org.federation", "Posts"), ref.Message(t, "org.federation", "PostsArgument")).
				AddMessage(ref.Message(t, "org.federation", "User"), ref.Message(t, "org.federation", "UserArgument")).
				Build(t),
		)

	federationFile := fb.Build(t)
	federationService := federationFile.Services[0]

	r := resolver.New(testutil.Compile(t, fileName))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("faield to get files. expected 1 but got %d", len(result.Files))
	}
	if len(result.Files[0].Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Files[0].Services))
	}
	if diff := cmp.Diff(result.Files[0].Services[0], federationService, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestCondition(t *testing.T) {
	fileName := filepath.Join(testutil.RepoRoot(), "testdata", "condition.proto")
	fb := testutil.NewFileBuilder(fileName)
	ref := testutil.NewBuilderReferenceManager(getPostProtoBuilder(t), fb)

	fb.SetPackage("org.federation").
		SetGoPackage("example/federation", "federation").
		AddMessage(
			testutil.NewMessageBuilder("PostArgument").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("UserArgument").
				AddField("user_id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("User").
				AddFieldWithRule(
					"id",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(resolver.NewByValue("$.user_id", resolver.StringType)).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						SetMessageArgument(ref.Message(t, "org.federation", "UserArgument")).
						Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("Post").
				AddFieldWithRule(
					"id",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(resolver.NewByValue("post.id", resolver.StringType)).Build(t),
				).
				AddFieldWithRule(
					"title",
					resolver.StringType,
					testutil.NewFieldRuleBuilder(resolver.NewByValue("post.title)", resolver.StringType)).Build(t),
				).
				AddFieldWithRule(
					"user",
					ref.Type(t, "org.federation", "User"),
					testutil.NewFieldRuleBuilder(resolver.NewByValue("users[0]", ref.Type(t, "org.federation", "User"))).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetIf("$.id != ''").
								SetName("res").
								SetUsed(true).
								SetCall(
									testutil.NewCallExprBuilder().
										SetMethod(ref.Method(t, "org.post", "PostService", "GetPost")).
										SetRequest(
											testutil.NewRequestBuilder().
												AddField(
													"id",
													resolver.StringType,
													resolver.NewByValue("$.id", resolver.StringType),
												).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetIf("res != null").
								SetName("post").
								SetUsed(true).
								SetBy(testutil.NewCELValueBuilder("res.post", ref.Type(t, "org.post", "Post")).Build(t)).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetIf("post != null").
								SetName("user").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "User")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("user_id", resolver.NewByValue("post.user_id", resolver.StringType)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("posts").
								SetUsed(true).
								SetBy(testutil.NewCELValueBuilder("[post]", ref.RepeatedType(t, "org.post", "Post")).Build(t)).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetIf("user != null").
								SetName("users").
								SetUsed(true).
								SetMap(
									testutil.NewMapExprBuilder().
										SetIterator(
											testutil.NewIteratorBuilder().
												SetName("iter").
												SetSource("posts").
												Build(t),
										).
										SetExpr(
											testutil.NewMapIteratorExprBuilder().
												SetMessage(
													testutil.NewMessageExprBuilder().
														SetMessage(ref.Message(t, "org.federation", "User")).
														SetArgs(
															testutil.NewMessageDependencyArgumentBuilder().
																Add("user_id", resolver.NewByValue("iter.user_id", resolver.StringType)).
																Build(t),
														).
														Build(t),
												).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetIf("users.size() > 0").
								SetName("_def5").
								SetValidation(
									testutil.NewValidationExprBuilder().
										SetError(
											testutil.NewGRPCErrorBuilder().
												SetCode(code.Code_INVALID_ARGUMENT).
												SetIf("users[0].id == ''").
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "PostArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.post", "GetPostResponse"), ref.Message(t, "org.federation", "User")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(
									testutil.NewVariableDefinitionGroupBuilder().
										AddStart(
											testutil.NewVariableDefinitionGroupBuilder().
												AddStart(
													testutil.NewVariableDefinitionGroupBuilder().
														AddStart(testutil.NewVariableDefinitionGroupByName("res")).
														SetEnd(testutil.NewVariableDefinition("post")).
														Build(t),
												).
												SetEnd(testutil.NewVariableDefinition("posts")).
												Build(t),
										).
										AddStart(
											testutil.NewVariableDefinitionGroupBuilder().
												AddStart(
													testutil.NewVariableDefinitionGroupBuilder().
														AddStart(testutil.NewVariableDefinitionGroupByName("res")).
														SetEnd(testutil.NewVariableDefinition("post")).
														Build(t),
												).
												SetEnd(testutil.NewVariableDefinition("user")).
												Build(t),
										).
										SetEnd(testutil.NewVariableDefinition("users")).
										Build(t),
								).
								SetEnd(testutil.NewVariableDefinition("_def5")).
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
					testutil.NewFieldRuleBuilder(resolver.NewByValue("post", ref.Type(t, "org.federation", "Post"))).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("post").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "Post")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("id", resolver.NewByValue("$.id", resolver.StringType)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "GetPostResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "Post")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("post")).
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
					nil,
				).
				SetRule(testutil.NewServiceRuleBuilder().Build(t)).
				AddMessage(ref.Message(t, "org.federation", "GetPostResponse"), ref.Message(t, "org.federation", "GetPostResponseArgument")).
				AddMessage(ref.Message(t, "org.federation", "Post"), ref.Message(t, "org.federation", "PostArgument")).
				AddMessage(ref.Message(t, "org.federation", "User"), ref.Message(t, "org.federation", "UserArgument")).
				Build(t),
		)

	federationFile := fb.Build(t)
	federationService := federationFile.Services[0]

	r := resolver.New(testutil.Compile(t, fileName))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("faield to get files. expected 1 but got %d", len(result.Files))
	}
	if len(result.Files[0].Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Files[0].Services))
	}
	if diff := cmp.Diff(result.Files[0].Services[0], federationService, testutil.ResolverCmpOpts()...); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestErrorHandler(t *testing.T) {
	fileName := filepath.Join(testutil.RepoRoot(), "testdata", "error_handler.proto")
	fb := testutil.NewFileBuilder(fileName)
	ref := testutil.NewBuilderReferenceManager(getPostProtoBuilder(t), fb)

	fb.SetPackage("org.federation").
		SetGoPackage("example/federation", "federation").
		AddMessage(
			testutil.NewMessageBuilder("CustomMessageArgument").
				AddField("msg", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("CustomMessage").
				AddFieldWithRule("msg", resolver.StringType, testutil.NewFieldRuleBuilder(&resolver.Value{CEL: testutil.NewCELValueBuilder("'custom error message:' + $.msg", resolver.StringType).Build(t)}).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().SetMessageArgument(ref.Message(t, "org.federation", "CustomMessageArgument")).Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("LocalizedMessageArgument").
				AddField("value", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("LocalizedMessage").
				AddFieldWithRule("value", resolver.StringType, testutil.NewFieldRuleBuilder(&resolver.Value{CEL: testutil.NewCELValueBuilder("'localized value:' + $.value", resolver.StringType).Build(t)}).Build(t)).
				SetRule(
					testutil.NewMessageRuleBuilder().SetMessageArgument(ref.Message(t, "org.federation", "LocalizedMessageArgument")).Build(t),
				).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("PostArgument").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetPostResponseArgument").
				AddField("id", resolver.StringType).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("Post").
				AddFieldWithAutoBind("id", resolver.StringType, ref.Field(t, "org.post", "Post", "id")).
				AddFieldWithAutoBind("title", resolver.StringType, ref.Field(t, "org.post", "Post", "title")).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("res").
								SetUsed(true).
								SetCall(
									testutil.NewCallExprBuilder().
										SetMethod(ref.Method(t, "org.post", "PostService", "GetPost")).
										SetRequest(
											testutil.NewRequestBuilder().
												AddField(
													"id",
													resolver.StringType,
													resolver.NewByValue("$.id", resolver.StringType),
												).
												Build(t),
										).
										AddError(
											testutil.NewGRPCErrorBuilder().
												AddVariableDefinition(
													testutil.NewVariableDefinitionBuilder().
														SetName("id").
														SetUsed(true).
														SetBy(testutil.NewCELValueBuilder("$.id", resolver.StringType).Build(t)).
														Build(t),
												).
												SetIf("id == ''").
												SetCode(code.Code_FAILED_PRECONDITION).
												SetMessage("'id must be not empty'").
												AddDetail(
													testutil.NewGRPCErrorDetailBuilder().
														AddDef(
															testutil.NewVariableDefinitionBuilder().
																SetName("localized_msg").
																SetUsed(true).
																SetMessage(
																	testutil.NewMessageExprBuilder().
																		SetMessage(ref.Message(t, "org.federation", "LocalizedMessage")).
																		SetArgs(
																			testutil.NewMessageDependencyArgumentBuilder().
																				Add("value", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
																				Build(t),
																		).
																		Build(t),
																).
																Build(t),
														).
														AddMessage(
															testutil.NewVariableDefinitionBuilder().
																SetName("_def0_err_detail0_msg0").
																SetUsed(true).
																SetMessage(
																	testutil.NewMessageExprBuilder().
																		SetMessage(ref.Message(t, "org.federation", "CustomMessage")).
																		SetArgs(
																			testutil.NewMessageDependencyArgumentBuilder().
																				Add("msg", testutil.NewMessageArgumentValueBuilder(resolver.StringType, resolver.StringType, "id").Build(t)).
																				Build(t),
																		).
																		Build(t),
																).
																Build(t),
														).
														AddPreconditionFailure(&resolver.PreconditionFailure{
															Violations: []*resolver.PreconditionFailureViolation{
																{
																	Type:        testutil.NewCELValueBuilder("'some-type'", resolver.StringType).Build(t),
																	Subject:     testutil.NewCELValueBuilder("'some-subject'", resolver.StringType).Build(t),
																	Description: testutil.NewCELValueBuilder("'some-description'", resolver.StringType).Build(t),
																},
															},
														}).
														AddLocalizedMessage(&resolver.LocalizedMessage{
															Locale:  "en-US",
															Message: testutil.NewCELValueBuilder("localized_msg.value", resolver.StringType).Build(t),
														}).
														Build(t),
												).
												Build(t),
										).
										AddError(
											testutil.NewGRPCErrorBuilder().
												SetIf("true").
												SetIgnore(true).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("post").
								SetUsed(true).
								SetAutoBind(true).
								SetBy(testutil.NewCELValueBuilder("res.post", ref.Type(t, "org.post", "Post")).Build(t)).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "PostArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.post", "GetPostResponse")).
								Build(t),
						).
						AddVariableDefinitionGroup(
							testutil.NewVariableDefinitionGroupBuilder().
								AddStart(testutil.NewVariableDefinitionGroupByName("res")).
								SetEnd(testutil.NewVariableDefinition("post")).
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
			testutil.NewMessageBuilder("GetPostResponse").
				AddFieldWithRule(
					"post",
					ref.Type(t, "org.federation", "Post"),
					testutil.NewFieldRuleBuilder(resolver.NewByValue("post", ref.Type(t, "org.federation", "Post"))).Build(t),
				).
				SetRule(
					testutil.NewMessageRuleBuilder().
						AddVariableDefinition(
							testutil.NewVariableDefinitionBuilder().
								SetName("post").
								SetUsed(true).
								SetMessage(
									testutil.NewMessageExprBuilder().
										SetMessage(ref.Message(t, "org.federation", "Post")).
										SetArgs(
											testutil.NewMessageDependencyArgumentBuilder().
												Add("id", resolver.NewByValue("$.id", resolver.StringType)).
												Build(t),
										).
										Build(t),
								).
								Build(t),
						).
						SetMessageArgument(ref.Message(t, "org.federation", "GetPostResponseArgument")).
						SetDependencyGraph(
							testutil.NewDependencyGraphBuilder().
								Add(ref.Message(t, "org.federation", "Post")).
								Build(t),
						).
						AddVariableDefinitionGroup(testutil.NewVariableDefinitionGroupByName("post")).
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
					nil,
				).
				SetRule(
					testutil.NewServiceRuleBuilder().
						AddDependency("", ref.Service(t, "org.post", "PostService")).
						Build(t),
				).
				AddMessage(ref.Message(t, "org.federation", "CustomMessage"), ref.Message(t, "org.federation", "CustomMessageArgument")).
				AddMessage(ref.Message(t, "org.federation", "GetPostResponse"), ref.Message(t, "org.federation", "GetPostResponseArgument")).
				AddMessage(ref.Message(t, "org.federation", "LocalizedMessage"), ref.Message(t, "org.federation", "LocalizedMessageArgument")).
				AddMessage(ref.Message(t, "org.federation", "Post"), ref.Message(t, "org.federation", "PostArgument")).
				Build(t),
		)

	federationFile := fb.Build(t)
	federationService := federationFile.Services[0]

	r := resolver.New(testutil.Compile(t, fileName))
	result, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("faield to get files. expected 1 but got %d", len(result.Files))
	}
	if len(result.Files[0].Services) != 1 {
		t.Fatalf("faield to get services. expected 1 but got %d", len(result.Files[0].Services))
	}
	if diff := cmp.Diff(result.Files[0].Services[0], federationService, testutil.ResolverCmpOpts()...); diff != "" {
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
				AddMessage(
					testutil.NewMessageBuilder("AttrA").
						AddField("foo", resolver.StringType).
						Build(t),
				).
				AddMessage(
					testutil.NewMessageBuilder("AttrB").
						AddField("bar", resolver.BoolType).
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
				AddFieldWithTypeName(t, "attr_a", "AttrA", false).
				AddFieldWithTypeName(t, "b", "AttrB", false).
				AddOneof(testutil.NewOneofBuilder("attr").AddFieldNames("attr_a", "b").Build(t)).
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
				Build(t),
		)
	return pb
}

func getContentProtoBuilder(t *testing.T) *testutil.FileBuilder {
	t.Helper()
	pb := testutil.NewFileBuilder("content.proto")
	ref := testutil.NewBuilderReferenceManager(pb)

	pb.SetPackage("content").
		SetGoPackage("example/content", "content").
		AddEnum(
			testutil.NewEnumBuilder("ContentType").
				AddValue("CONTENT_TYPE_1").
				AddValue("CONTENT_TYPE_2").
				AddValue("CONTENT_TYPE_3").
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("Content").
				AddField("by_field", resolver.StringType).
				AddField("double_field", resolver.DoubleType).
				AddField("doubles_field", resolver.DoubleRepeatedType).
				AddField("float_field", resolver.FloatType).
				AddField("floats_field", resolver.FloatRepeatedType).
				AddField("int32_field", resolver.Int32Type).
				AddField("int32s_field", resolver.Int32RepeatedType).
				AddField("int64_field", resolver.Int64Type).
				AddField("int64s_field", resolver.Int64RepeatedType).
				AddField("uint32_field", resolver.Uint32Type).
				AddField("uint32s_field", resolver.Uint32RepeatedType).
				AddField("uint64_field", resolver.Uint64Type).
				AddField("uint64s_field", resolver.Uint64RepeatedType).
				AddField("sint32_field", resolver.Sint32Type).
				AddField("sint32s_field", resolver.Sint32RepeatedType).
				AddField("sint64_field", resolver.Sint64Type).
				AddField("sint64s_field", resolver.Sint64RepeatedType).
				AddField("fixed32_field", resolver.Fixed32Type).
				AddField("fixed32s_field", resolver.Fixed32RepeatedType).
				AddField("fixed64_field", resolver.Fixed64Type).
				AddField("fixed64s_field", resolver.Fixed64RepeatedType).
				AddField("sfixed32_field", resolver.Sfixed32Type).
				AddField("sfixed32s_field", resolver.Sfixed32RepeatedType).
				AddField("sfixed64_field", resolver.Sfixed64Type).
				AddField("sfixed64s_field", resolver.Sfixed64RepeatedType).
				AddField("bool_field", resolver.BoolType).
				AddField("bools_field", resolver.BoolRepeatedType).
				AddField("string_field", resolver.StringType).
				AddField("strings_field", resolver.StringRepeatedType).
				AddField("byte_string_field", resolver.BytesType).
				AddField("byte_strings_field", resolver.BytesRepeatedType).
				AddField("enum_field", ref.Type(t, "content", "ContentType")).
				AddField("enums_field", ref.RepeatedType(t, "content", "ContentType")).
				AddField("env_field", resolver.StringType).
				AddField("envs_field", resolver.StringRepeatedType).
				AddFieldWithTypeName(t, "message_field", "Content", false).
				AddFieldWithTypeName(t, "messages_field", "Content", true).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetContentRequest").
				AddField("by_field", resolver.StringType).
				AddField("double_field", resolver.DoubleType).
				AddField("doubles_field", resolver.DoubleRepeatedType).
				AddField("float_field", resolver.FloatType).
				AddField("floats_field", resolver.FloatRepeatedType).
				AddField("int32_field", resolver.Int32Type).
				AddField("int32s_field", resolver.Int32RepeatedType).
				AddField("int64_field", resolver.Int64Type).
				AddField("int64s_field", resolver.Int64RepeatedType).
				AddField("uint32_field", resolver.Uint32Type).
				AddField("uint32s_field", resolver.Uint32RepeatedType).
				AddField("uint64_field", resolver.Uint64Type).
				AddField("uint64s_field", resolver.Uint64RepeatedType).
				AddField("sint32_field", resolver.Sint32Type).
				AddField("sint32s_field", resolver.Sint32RepeatedType).
				AddField("sint64_field", resolver.Sint64Type).
				AddField("sint64s_field", resolver.Sint64RepeatedType).
				AddField("fixed32_field", resolver.Fixed32Type).
				AddField("fixed32s_field", resolver.Fixed32RepeatedType).
				AddField("fixed64_field", resolver.Fixed64Type).
				AddField("fixed64s_field", resolver.Fixed64RepeatedType).
				AddField("sfixed32_field", resolver.Sfixed32Type).
				AddField("sfixed32s_field", resolver.Sfixed32RepeatedType).
				AddField("sfixed64_field", resolver.Sfixed64Type).
				AddField("sfixed64s_field", resolver.Sfixed64RepeatedType).
				AddField("bool_field", resolver.BoolType).
				AddField("bools_field", resolver.BoolRepeatedType).
				AddField("string_field", resolver.StringType).
				AddField("strings_field", resolver.StringRepeatedType).
				AddField("byte_string_field", resolver.BytesType).
				AddField("byte_strings_field", resolver.BytesRepeatedType).
				AddField("enum_field", ref.Type(t, "content", "ContentType")).
				AddField("enums_field", ref.RepeatedType(t, "content", "ContentType")).
				AddField("env_field", resolver.StringType).
				AddField("envs_field", resolver.StringRepeatedType).
				AddField("message_field", ref.Type(t, "content", "Content")).
				AddField("messages_field", ref.RepeatedType(t, "content", "Content")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetContentResponse").
				AddField("content", ref.Type(t, "content", "Content")).
				Build(t),
		).
		AddMessage(
			testutil.NewMessageBuilder("GetContentResponseArgument").
				AddField("by_field", resolver.StringType).
				AddField("double_field", resolver.DoubleType).
				AddField("doubles_field", resolver.DoubleRepeatedType).
				AddField("float_field", resolver.FloatType).
				AddField("floats_field", resolver.FloatRepeatedType).
				AddField("int32_field", resolver.Int32Type).
				AddField("int32s_field", resolver.Int32RepeatedType).
				AddField("int64_field", resolver.Int64Type).
				AddField("int64s_field", resolver.Int64RepeatedType).
				AddField("uint32_field", resolver.Uint32Type).
				AddField("uint32s_field", resolver.Uint32RepeatedType).
				AddField("uint64_field", resolver.Uint64Type).
				AddField("uint64s_field", resolver.Uint64RepeatedType).
				AddField("sint32_field", resolver.Sint32Type).
				AddField("sint32s_field", resolver.Sint32RepeatedType).
				AddField("sint64_field", resolver.Sint64Type).
				AddField("sint64s_field", resolver.Sint64RepeatedType).
				AddField("fixed32_field", resolver.Fixed32Type).
				AddField("fixed32s_field", resolver.Fixed32RepeatedType).
				AddField("fixed64_field", resolver.Fixed64Type).
				AddField("fixed64s_field", resolver.Fixed64RepeatedType).
				AddField("sfixed32_field", resolver.Sfixed32Type).
				AddField("sfixed32s_field", resolver.Sfixed32RepeatedType).
				AddField("sfixed64_field", resolver.Sfixed64Type).
				AddField("sfixed64s_field", resolver.Sfixed64RepeatedType).
				AddField("bool_field", resolver.BoolType).
				AddField("bools_field", resolver.BoolRepeatedType).
				AddField("string_field", resolver.StringType).
				AddField("strings_field", resolver.StringRepeatedType).
				AddField("byte_string_field", resolver.BytesType).
				AddField("byte_strings_field", resolver.BytesRepeatedType).
				AddField("enum_field", ref.Type(t, "content", "ContentType")).
				AddField("enums_field", ref.RepeatedType(t, "content", "ContentType")).
				AddField("env_field", resolver.StringType).
				AddField("envs_field", resolver.StringRepeatedType).
				AddField("message_field", ref.Type(t, "content", "Content")).
				AddField("messages_field", ref.RepeatedType(t, "content", "Content")).
				Build(t),
		).
		AddService(
			testutil.NewServiceBuilder("ContentService").
				AddMethod("GetContent", ref.Message(t, "content", "GetContentRequest"), ref.Message(t, "content", "GetContentResponse"), nil).
				Build(t),
		)
	return pb
}

func TestGRPCError_ReferenceNames(t *testing.T) {
	v := &resolver.GRPCError{
		If: &resolver.CELValue{
			CheckedExpr: &exprv1.CheckedExpr{
				ReferenceMap: map[int64]*exprv1.Reference{
					0: {Name: "name1"},
				},
			},
		},
		Details: []*resolver.GRPCErrorDetail{
			{
				If: &resolver.CELValue{
					CheckedExpr: &exprv1.CheckedExpr{
						ReferenceMap: map[int64]*exprv1.Reference{
							0: {Name: "name2"},
						},
					},
				},
				PreconditionFailures: []*resolver.PreconditionFailure{
					{
						Violations: []*resolver.PreconditionFailureViolation{
							{
								Type: &resolver.CELValue{
									CheckedExpr: &exprv1.CheckedExpr{
										ReferenceMap: map[int64]*exprv1.Reference{
											0: {Name: "name3"},
										},
									},
								},
								Subject: &resolver.CELValue{
									CheckedExpr: &exprv1.CheckedExpr{
										ReferenceMap: map[int64]*exprv1.Reference{
											0: {Name: "name4"},
										},
									},
								},
								Description: &resolver.CELValue{
									CheckedExpr: &exprv1.CheckedExpr{
										ReferenceMap: map[int64]*exprv1.Reference{
											0: {Name: "name5"},
										},
									},
								},
							},
							{
								Type: &resolver.CELValue{
									CheckedExpr: &exprv1.CheckedExpr{
										ReferenceMap: map[int64]*exprv1.Reference{
											0: {Name: "name3"},
										},
									},
								},
								Subject: &resolver.CELValue{
									CheckedExpr: &exprv1.CheckedExpr{
										ReferenceMap: map[int64]*exprv1.Reference{
											0: {Name: "name4"},
										},
									},
								},
								Description: &resolver.CELValue{
									CheckedExpr: &exprv1.CheckedExpr{
										ReferenceMap: map[int64]*exprv1.Reference{
											0: {Name: "name6"},
										},
									},
								},
							},
						},
					},
				},
				BadRequests: []*resolver.BadRequest{
					{
						FieldViolations: []*resolver.BadRequestFieldViolation{
							{
								Field: &resolver.CELValue{
									CheckedExpr: &exprv1.CheckedExpr{
										ReferenceMap: map[int64]*exprv1.Reference{
											0: {Name: "name7"},
										},
									},
								},
								Description: &resolver.CELValue{
									CheckedExpr: &exprv1.CheckedExpr{
										ReferenceMap: map[int64]*exprv1.Reference{
											0: {Name: "name8"},
										},
									},
								},
							},
						},
					},
				},
				LocalizedMessages: []*resolver.LocalizedMessage{
					{
						Locale: "en-US",
						Message: &resolver.CELValue{
							CheckedExpr: &exprv1.CheckedExpr{
								ReferenceMap: map[int64]*exprv1.Reference{
									0: {Name: "name9"},
								},
							},
						},
					},
				},
			},
		},
	}

	expected := []string{
		"name1",
		"name2",
		"name3",
		"name4",
		"name5",
		"name6",
		"name7",
		"name8",
		"name9",
	}
	got := v.ReferenceNames()
	// the map order is not guaranteed in Go
	if len(expected) != len(got) {
		t.Errorf("the number of reference names is different: expected: %d, got: %d", len(expected), len(got))
	}
	for _, e := range expected {
		var found bool
		for _, g := range got {
			if g == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%q does not exist in the received reference names: %v", e, got)
		}
	}
}
