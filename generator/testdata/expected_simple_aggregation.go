// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
package federation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"

	post "example/post"
	user "example/user"
)

// FederationServiceConfig configuration required to initialize the service that use GRPC Federation.
type FederationServiceConfig struct {
	// Client provides a factory that creates the gRPC Client needed to invoke methods of the gRPC Service on which the Federation Service depends.
	// If this interface is not provided, an error is returned during initialization.
	Client FederationServiceClientFactory // required
	// Resolver provides an interface to directly implement message resolver and field resolver not defined in Protocol Buffers.
	// If this interface is not provided, an error is returned during initialization.
	Resolver FederationServiceResolver // required
	// ErrorHandler Federation Service often needs to convert errors received from downstream services.
	// If an error occurs during method execution in the Federation Service, this error handler is called and the returned error is treated as a final error.
	ErrorHandler FederationServiceErrorHandler
	// Logger sets the logger used to output Debug/Info/Error information.
	Logger *slog.Logger
}

// FederationServiceClientFactory provides a factory that creates the gRPC Client needed to invoke methods of the gRPC Service on which the Federation Service depends.
type FederationServiceClientFactory interface {
	// Org_Post_PostServiceClient create a gRPC Client to be used to call methods in org.post.PostService.
	Org_Post_PostServiceClient(FederationServiceClientConfig) (post.PostServiceClient, error)
	// Org_User_UserServiceClient create a gRPC Client to be used to call methods in org.user.UserService.
	Org_User_UserServiceClient(FederationServiceClientConfig) (user.UserServiceClient, error)
}

// FederationServiceClientConfig information set in `dependencies` of the `grpc.federation.service` option.
// Hints for creating a gRPC Client.
type FederationServiceClientConfig struct {
	// Service returns the name of the service on Protocol Buffers.
	Service string
	// Name is the value set for `name` in `dependencies` of the `grpc.federation.service` option.
	// It must be unique among the services on which the Federation Service depends.
	Name string
}

// FederationServiceDependencyServiceClient has a gRPC client for all services on which the federation service depends.
// This is provided as an argument when implementing the custom resolver.
type FederationServiceDependencyServiceClient struct {
	Org_Post_PostServiceClient post.PostServiceClient
	Org_User_UserServiceClient user.UserServiceClient
}

// FederationServiceResolver provides an interface to directly implement message resolver and field resolver not defined in Protocol Buffers.
type FederationServiceResolver interface {
	// Resolve_Org_Federation_User_Age implements resolver for "org.federation.User.age".
	Resolve_Org_Federation_User_Age(context.Context, *Org_Federation_User_AgeArgument) (uint64, error)
	// Resolve_Org_Federation_Z implements resolver for "org.federation.Z".
	Resolve_Org_Federation_Z(context.Context, *Org_Federation_ZArgument) (*Z, error)
}

// FederationServiceUnimplementedResolver a structure implemented to satisfy the Resolver interface.
// An Unimplemented error is always returned.
// This is intended for use when there are many Resolver interfaces that do not need to be implemented,
// by embedding them in a resolver structure that you have created.
type FederationServiceUnimplementedResolver struct{}

// Resolve_Org_Federation_User_Age resolve "org.federation.User.age".
// This method always returns Unimplemented error.
func (FederationServiceUnimplementedResolver) Resolve_Org_Federation_User_Age(context.Context, *Org_Federation_User_AgeArgument) (ret uint64, e error) {
	e = grpcstatus.Errorf(grpccodes.Unimplemented, "method Resolve_Org_Federation_User_Age not implemented")
	return
}

// Resolve_Org_Federation_Z resolve "org.federation.Z".
// This method always returns Unimplemented error.
func (FederationServiceUnimplementedResolver) Resolve_Org_Federation_Z(context.Context, *Org_Federation_ZArgument) (ret *Z, e error) {
	e = grpcstatus.Errorf(grpccodes.Unimplemented, "method Resolve_Org_Federation_Z not implemented")
	return
}

// FederationServiceErrorHandler Federation Service often needs to convert errors received from downstream services.
// If an error occurs during method execution in the Federation Service, this error handler is called and the returned error is treated as a final error.
type FederationServiceErrorHandler func(ctx context.Context, methodName string, err error) error

const (
	FederationService_DependentMethod_Org_Post_PostService_GetPost = "/org.post.PostService/GetPost"
	FederationService_DependentMethod_Org_User_UserService_GetUser = "/org.user.UserService/GetUser"
)

// FederationServiceRecoveredError represents recovered error.
type FederationServiceRecoveredError struct {
	Message string
	Stack   []string
}

func (e *FederationServiceRecoveredError) Error() string {
	return fmt.Sprintf("recovered error: %s", e.Message)
}

// FederationService represents Federation Service.
type FederationService struct {
	*UnimplementedFederationServiceServer
	cfg          FederationServiceConfig
	logger       *slog.Logger
	errorHandler FederationServiceErrorHandler
	resolver     FederationServiceResolver
	client       *FederationServiceDependencyServiceClient
}

// Org_Federation_GetPostResponseArgument is argument for "org.federation.GetPostResponse" message.
type Org_Federation_GetPostResponseArgument struct {
	Id     string
	Post   *Post
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_MArgument is argument for "org.federation.M" message.
type Org_Federation_MArgument struct {
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_PostArgument is argument for "org.federation.Post" message.
type Org_Federation_PostArgument struct {
	Id     string
	M      *M
	Post   *post.Post
	User   *User
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_UserArgument is argument for "org.federation.User" message.
type Org_Federation_UserArgument struct {
	Content string
	Id      string
	Title   string
	User    *user.User
	UserId  string
	Client  *FederationServiceDependencyServiceClient
}

// Org_Federation_User_AgeArgument is custom resolver's argument for "age" field of "org.federation.User" message.
type Org_Federation_User_AgeArgument struct {
	*Org_Federation_UserArgument
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_ZArgument is argument for "org.federation.Z" message.
type Org_Federation_ZArgument struct {
	Client *FederationServiceDependencyServiceClient
}

// NewFederationService creates FederationService instance by FederationServiceConfig.
func NewFederationService(cfg FederationServiceConfig) (*FederationService, error) {
	if err := validateFederationServiceConfig(cfg); err != nil {
		return nil, err
	}
	Org_Post_PostServiceClient, err := cfg.Client.Org_Post_PostServiceClient(FederationServiceClientConfig{
		Service: "org.post.PostService",
		Name:    "",
	})
	if err != nil {
		return nil, err
	}
	Org_User_UserServiceClient, err := cfg.Client.Org_User_UserServiceClient(FederationServiceClientConfig{
		Service: "org.user.UserService",
		Name:    "",
	})
	if err != nil {
		return nil, err
	}
	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	errorHandler := cfg.ErrorHandler
	if errorHandler == nil {
		errorHandler = func(ctx context.Context, methodName string, err error) error { return err }
	}
	return &FederationService{
		cfg:          cfg,
		logger:       logger,
		errorHandler: errorHandler,
		resolver:     cfg.Resolver,
		client: &FederationServiceDependencyServiceClient{
			Org_Post_PostServiceClient: Org_Post_PostServiceClient,
			Org_User_UserServiceClient: Org_User_UserServiceClient,
		},
	}, nil
}

func validateFederationServiceConfig(cfg FederationServiceConfig) error {
	if cfg.Client == nil {
		return fmt.Errorf("Client field in FederationServiceConfig is not set. this field must be set")
	}
	if cfg.Resolver == nil {
		return fmt.Errorf("Resolver field in FederationServiceConfig is not set. this field must be set")
	}
	return nil
}

func withTimeoutFederationService[T any](ctx context.Context, method string, timeout time.Duration, fn func(context.Context) (*T, error)) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		ret   *T
		errch = make(chan error)
	)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errch <- recoverErrorFederationService(r, debug.Stack())
			}
		}()

		res, err := fn(ctx)
		ret = res
		errch <- err
	}()
	select {
	case <-ctx.Done():
		status := grpcstatus.New(grpccodes.DeadlineExceeded, ctx.Err().Error())
		withDetails, err := status.WithDetails(&errdetails.ErrorInfo{
			Metadata: map[string]string{
				"method":  method,
				"timeout": timeout.String(),
			},
		})
		if err != nil {
			return nil, status.Err()
		}
		return nil, withDetails.Err()
	case err := <-errch:
		return ret, err
	}
}

func withRetryFederationService[T any](b backoff.BackOff, fn func() (*T, error)) (*T, error) {
	var res *T
	if err := backoff.Retry(func() (err error) {
		res, err = fn()
		return
	}, b); err != nil {
		return nil, err
	}
	return res, nil
}

func recoverErrorFederationService(v interface{}, rawStack []byte) *FederationServiceRecoveredError {
	msg := fmt.Sprint(v)
	lines := strings.Split(msg, "\n")
	if len(lines) <= 1 {
		lines := strings.Split(string(rawStack), "\n")
		stack := make([]string, 0, len(lines))
		for _, line := range lines {
			if line == "" {
				continue
			}
			stack = append(stack, strings.TrimPrefix(line, "\t"))
		}
		return &FederationServiceRecoveredError{
			Message: msg,
			Stack:   stack,
		}
	}
	// If panic occurs under singleflight, singleflight's recover catches the error and gives a stack trace.
	// Therefore, once the stack trace is removed.
	stack := make([]string, 0, len(lines))
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		stack = append(stack, strings.TrimPrefix(line, "\t"))
	}
	return &FederationServiceRecoveredError{
		Message: lines[0],
		Stack:   stack,
	}
}

func (s *FederationService) goWithRecover(eg *errgroup.Group, fn func() (interface{}, error)) {
	eg.Go(func() (e error) {
		defer func() {
			if r := recover(); r != nil {
				e = recoverErrorFederationService(r, debug.Stack())
			}
		}()
		_, err := fn()
		return err
	})
}

func (s *FederationService) outputErrorLog(ctx context.Context, err error) {
	if err == nil {
		return
	}
	if status, ok := grpcstatus.FromError(err); ok {
		s.logger.ErrorContext(ctx, status.Message(),
			slog.Group("grpc_status",
				slog.String("code", status.Code().String()),
				slog.Any("details", status.Details()),
			),
		)
		return
	}
	var recoveredErr *FederationServiceRecoveredError
	if errors.As(err, &recoveredErr) {
		trace := make([]interface{}, 0, len(recoveredErr.Stack))
		for idx, stack := range recoveredErr.Stack {
			trace = append(trace, slog.String(fmt.Sprint(idx+1), stack))
		}
		s.logger.ErrorContext(ctx, recoveredErr.Message, slog.Group("stack_trace", trace...))
		return
	}
	s.logger.ErrorContext(ctx, err.Error())
}

// GetPost implements "org.federation.FederationService/GetPost" method.
func (s *FederationService) GetPost(ctx context.Context, req *GetPostRequest) (res *GetPostResponse, e error) {
	defer func() {
		if r := recover(); r != nil {
			e = recoverErrorFederationService(r, debug.Stack())
			s.outputErrorLog(ctx, e)
		}
	}()
	res, err := withTimeoutFederationService[GetPostResponse](ctx, "org.federation.FederationService/GetPost", 60000000000 /* 1m0s */, func(ctx context.Context) (*GetPostResponse, error) {
		return s.resolve_Org_Federation_GetPostResponse(ctx, &Org_Federation_GetPostResponseArgument{
			Client: s.client,
			Id:     req.Id,
		})
	})
	if err != nil {
		s.outputErrorLog(ctx, err)
		return nil, err
	}
	return res, nil
}

// resolve_Org_Federation_GetPostResponse resolve "org.federation.GetPostResponse" message.
func (s *FederationService) resolve_Org_Federation_GetPostResponse(ctx context.Context, req *Org_Federation_GetPostResponseArgument) (*GetPostResponse, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.GetPostResponse", slog.Any("message_args", s.logvalue_Org_Federation_GetPostResponseArgument(req)))
	var (
		sg        singleflight.Group
		valueMu   sync.RWMutex
		valuePost *Post
	)

	// This section's codes are generated by the following proto definition.
	/*
	   {
	     name: "post"
	     message: "Post"
	     args { name: "id", by: "$.id" }
	   }
	*/
	resPostIface, err, _ := sg.Do("post_org.federation.Post", func() (interface{}, error) {
		valueMu.RLock()
		args := &Org_Federation_PostArgument{
			Client: s.client,
			Id:     req.Id, // { name: "id", by: "$.id" }
		}
		valueMu.RUnlock()
		return s.resolve_Org_Federation_Post(ctx, args)
	})
	if err != nil {
		return nil, err
	}
	resPost := resPostIface.(*Post)
	valueMu.Lock()
	valuePost = resPost // { name: "post", message: "Post" ... }
	valueMu.Unlock()

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Post = valuePost

	// create a message value to be returned.
	ret := &GetPostResponse{}

	// field binding section.
	ret.Post = valuePost // (grpc.federation.field).by = "post"
	ret.Literal = "foo"  // (grpc.federation.field).string = "foo"

	s.logger.DebugContext(ctx, "resolved org.federation.GetPostResponse", slog.Any("org.federation.GetPostResponse", s.logvalue_Org_Federation_GetPostResponse(ret)))
	return ret, nil
}

// resolve_Org_Federation_M resolve "org.federation.M" message.
func (s *FederationService) resolve_Org_Federation_M(ctx context.Context, req *Org_Federation_MArgument) (*M, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.M", slog.Any("message_args", s.logvalue_Org_Federation_MArgument(req)))

	// create a message value to be returned.
	ret := &M{}

	// field binding section.
	ret.Foo = "foo" // (grpc.federation.field).string = "foo"
	ret.Bar = 1     // (grpc.federation.field).int64 = 1

	s.logger.DebugContext(ctx, "resolved org.federation.M", slog.Any("org.federation.M", s.logvalue_Org_Federation_M(ret)))
	return ret, nil
}

// resolve_Org_Federation_Post resolve "org.federation.Post" message.
func (s *FederationService) resolve_Org_Federation_Post(ctx context.Context, req *Org_Federation_PostArgument) (*Post, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.Post", slog.Any("message_args", s.logvalue_Org_Federation_PostArgument(req)))
	var (
		sg        singleflight.Group
		valueM    *M
		valueMu   sync.RWMutex
		valuePost *post.Post
		valueUser *User
	)
	// A tree view of message dependencies is shown below.
	/*
	               m ─┐
	   GetPost ─┐     │
	            user ─┤
	               z ─┤
	*/
	eg, ctx := errgroup.WithContext(ctx)

	s.goWithRecover(eg, func() (interface{}, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   {
		     name: "m"
		     message: "M"
		     autobind: true
		   }
		*/
		resMIface, err, _ := sg.Do("m_org.federation.M", func() (interface{}, error) {
			valueMu.RLock()
			args := &Org_Federation_MArgument{
				Client: s.client,
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_M(ctx, args)
		})
		if err != nil {
			return nil, err
		}
		resM := resMIface.(*M)
		valueMu.Lock()
		valueM = resM // { name: "m", message: "M" ... }
		valueMu.Unlock()
		return nil, nil
	})

	s.goWithRecover(eg, func() (interface{}, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   resolver: {
		     method: "org.post.PostService/GetPost"
		     request { field: "id", by: "$.id" }
		     response { name: "post", field: "post", autobind: true }
		   }
		*/
		resGetPostResponseIface, err, _ := sg.Do("org.post.PostService/GetPost", func() (interface{}, error) {
			valueMu.RLock()
			args := &post.GetPostRequest{
				Id: req.Id, // { field: "id", by: "$.id" }
			}
			valueMu.RUnlock()
			return withTimeoutFederationService[post.GetPostResponse](ctx, "org.post.PostService/GetPost", 10000000000 /* 10s */, func(ctx context.Context) (*post.GetPostResponse, error) {
				var b backoff.BackOff = backoff.NewConstantBackOff(2000000000 /* 2s */)
				b = backoff.WithMaxRetries(b, 3)
				b = backoff.WithContext(b, ctx)
				return withRetryFederationService[post.GetPostResponse](b, func() (*post.GetPostResponse, error) {
					return s.client.Org_Post_PostServiceClient.GetPost(ctx, args)
				})
			})
		})
		if err != nil {
			if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_Post_PostService_GetPost, err); err != nil {
				return nil, err
			}
		}
		resGetPostResponse := resGetPostResponseIface.(*post.GetPostResponse)
		valueMu.Lock()
		valuePost = resGetPostResponse.GetPost() // { name: "post", field: "post", autobind: true }
		valueMu.Unlock()

		// This section's codes are generated by the following proto definition.
		/*
		   {
		     name: "user"
		     message: "User"
		     args { inline: "post" }
		   }
		*/
		resUserIface, err, _ := sg.Do("user_org.federation.User", func() (interface{}, error) {
			valueMu.RLock()
			args := &Org_Federation_UserArgument{
				Client:  s.client,
				Id:      valuePost.GetId(),      // { inline: "post" }
				Title:   valuePost.GetTitle(),   // { inline: "post" }
				Content: valuePost.GetContent(), // { inline: "post" }
				UserId:  valuePost.GetUserId(),  // { inline: "post" }
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_User(ctx, args)
		})
		if err != nil {
			return nil, err
		}
		resUser := resUserIface.(*User)
		valueMu.Lock()
		valueUser = resUser // { name: "user", message: "User" ... }
		valueMu.Unlock()
		return nil, nil
	})

	s.goWithRecover(eg, func() (interface{}, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   {
		     name: "z"
		     message: "Z"
		   }
		*/
		if _, err, _ := sg.Do("z_org.federation.Z", func() (interface{}, error) {
			valueMu.RLock()
			args := &Org_Federation_ZArgument{
				Client: s.client,
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_Z(ctx, args)
		}); err != nil {
			return nil, err
		}
		valueMu.Lock()
		valueMu.Unlock()
		return nil, nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.M = valueM
	req.Post = valuePost
	req.User = valueUser

	// create a message value to be returned.
	ret := &Post{}

	// field binding section.
	ret.Id = valuePost.GetId()           // { name: "post", autobind: true }
	ret.Title = valuePost.GetTitle()     // { name: "post", autobind: true }
	ret.Content = valuePost.GetContent() // { name: "post", autobind: true }
	ret.User = valueUser                 // (grpc.federation.field).by = "user"
	ret.Foo = valueM.GetFoo()            // { name: "m", autobind: true }
	ret.Bar = valueM.GetBar()            // { name: "m", autobind: true }

	s.logger.DebugContext(ctx, "resolved org.federation.Post", slog.Any("org.federation.Post", s.logvalue_Org_Federation_Post(ret)))
	return ret, nil
}

// resolve_Org_Federation_User resolve "org.federation.User" message.
func (s *FederationService) resolve_Org_Federation_User(ctx context.Context, req *Org_Federation_UserArgument) (*User, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.User", slog.Any("message_args", s.logvalue_Org_Federation_UserArgument(req)))
	var (
		sg        singleflight.Group
		valueMu   sync.RWMutex
		valueUser *user.User
	)

	// This section's codes are generated by the following proto definition.
	/*
	   resolver: {
	     method: "org.user.UserService/GetUser"
	     request { field: "id", by: "$.user_id" }
	     response { name: "user", field: "user", autobind: true }
	   }
	*/
	resGetUserResponseIface, err, _ := sg.Do("org.user.UserService/GetUser", func() (interface{}, error) {
		valueMu.RLock()
		args := &user.GetUserRequest{
			Id: req.UserId, // { field: "id", by: "$.user_id" }
		}
		valueMu.RUnlock()
		return withTimeoutFederationService[user.GetUserResponse](ctx, "org.user.UserService/GetUser", 20000000000 /* 20s */, func(ctx context.Context) (*user.GetUserResponse, error) {
			eb := backoff.NewExponentialBackOff()
			eb.InitialInterval = 1000000000 /* 1s */
			eb.RandomizationFactor = 0.7
			eb.Multiplier = 1.7
			eb.MaxInterval = 30000000000    /* 30s */
			eb.MaxElapsedTime = 20000000000 /* 20s */

			var b backoff.BackOff = eb
			b = backoff.WithMaxRetries(b, 3)
			b = backoff.WithContext(b, ctx)
			return withRetryFederationService[user.GetUserResponse](b, func() (*user.GetUserResponse, error) {
				return s.client.Org_User_UserServiceClient.GetUser(ctx, args)
			})
		})
	})
	if err != nil {
		if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_User_UserService_GetUser, err); err != nil {
			return nil, err
		}
	}
	resGetUserResponse := resGetUserResponseIface.(*user.GetUserResponse)
	valueMu.Lock()
	valueUser = resGetUserResponse.GetUser() // { name: "user", field: "user", autobind: true }
	valueMu.Unlock()

	// assign named parameters to message arguments to pass to the custom resolver.
	req.User = valueUser

	// create a message value to be returned.
	ret := &User{}

	// field binding section.
	ret.Id = valueUser.GetId()                                                            // { name: "user", autobind: true }
	ret.Type = s.cast_Org_User_UserType__to__Org_Federation_UserType(valueUser.GetType()) // { name: "user", autobind: true }
	ret.Name = valueUser.GetName()                                                        // { name: "user", autobind: true }
	{
		// (grpc.federation.field).custom_resolver = true
		var err error
		ret.Age, err = s.resolver.Resolve_Org_Federation_User_Age(ctx, &Org_Federation_User_AgeArgument{
			Client:                      s.client,
			Org_Federation_UserArgument: req,
		})
		if err != nil {
			return nil, err
		}
	}
	ret.Desc = valueUser.GetDesc()                                                                    // { name: "user", autobind: true }
	ret.MainItem = s.cast_Org_User_Item__to__Org_Federation_Item(valueUser.GetMainItem())             // { name: "user", autobind: true }
	ret.Items = s.cast_repeated_Org_User_Item__to__repeated_Org_Federation_Item(valueUser.GetItems()) // { name: "user", autobind: true }
	ret.Profile = valueUser.GetProfile()                                                              // { name: "user", autobind: true }

	switch {
	case s.cast_Org_User_User_AttrA___to__Org_Federation_User_AttrA_(valueUser.GetAttrA()) != nil:
		ret.Attr = s.cast_Org_User_User_AttrA___to__Org_Federation_User_AttrA_(valueUser.GetAttrA())
	case s.cast_Org_User_User_B__to__Org_Federation_User_B(valueUser.GetB()) != nil:
		ret.Attr = s.cast_Org_User_User_B__to__Org_Federation_User_B(valueUser.GetB())
	}

	s.logger.DebugContext(ctx, "resolved org.federation.User", slog.Any("org.federation.User", s.logvalue_Org_Federation_User(ret)))
	return ret, nil
}

// cast_Org_User_Item_ItemType__to__Org_Federation_Item_ItemType cast from "org.user.Item.ItemType" to "org.federation.Item.ItemType".
func (s *FederationService) cast_Org_User_Item_ItemType__to__Org_Federation_Item_ItemType(from user.Item_ItemType) Item_ItemType {
	switch from {
	case user.Item_ITEM_TYPE_1:
		return Item_ITEM_TYPE_1
	case user.Item_ITEM_TYPE_2:
		return Item_ITEM_TYPE_2
	case user.Item_ITEM_TYPE_3:
		return Item_ITEM_TYPE_3
	default:
		return 0
	}
}

// cast_Org_User_Item__to__Org_Federation_Item cast from "org.user.Item" to "org.federation.Item".
func (s *FederationService) cast_Org_User_Item__to__Org_Federation_Item(from *user.Item) *Item {
	if from == nil {
		return nil
	}

	return &Item{
		Name:  from.GetName(),
		Type:  s.cast_Org_User_Item_ItemType__to__Org_Federation_Item_ItemType(from.GetType()),
		Value: from.GetValue(),
	}
}

// cast_Org_User_User_AttrA__to__Org_Federation_User_AttrA cast from "org.user.User.AttrA" to "org.federation.User.AttrA".
func (s *FederationService) cast_Org_User_User_AttrA__to__Org_Federation_User_AttrA(from *user.User_AttrA) *User_AttrA {
	if from == nil {
		return nil
	}

	return &User_AttrA{
		Foo: from.GetFoo(),
	}
}

// cast_Org_User_User_AttrB__to__Org_Federation_User_AttrB cast from "org.user.User.AttrB" to "org.federation.User.AttrB".
func (s *FederationService) cast_Org_User_User_AttrB__to__Org_Federation_User_AttrB(from *user.User_AttrB) *User_AttrB {
	if from == nil {
		return nil
	}

	return &User_AttrB{
		Bar: from.GetBar(),
	}
}

// cast_Org_User_User_AttrA___to__Org_Federation_User_AttrA_ cast from "org.user.User.attr_a" to "org.federation.User.attr_a".
func (s *FederationService) cast_Org_User_User_AttrA___to__Org_Federation_User_AttrA_(from *user.User_AttrA) *User_AttrA_ {
	if from == nil {
		return nil
	}
	return &User_AttrA_{
		AttrA: s.cast_Org_User_User_AttrA__to__Org_Federation_User_AttrA(from),
	}
}

// cast_Org_User_User_B__to__Org_Federation_User_B cast from "org.user.User.b" to "org.federation.User.b".
func (s *FederationService) cast_Org_User_User_B__to__Org_Federation_User_B(from *user.User_AttrB) *User_B {
	if from == nil {
		return nil
	}
	return &User_B{
		B: s.cast_Org_User_User_AttrB__to__Org_Federation_User_AttrB(from),
	}
}

// cast_Org_User_UserType__to__Org_Federation_UserType cast from "org.user.UserType" to "org.federation.UserType".
func (s *FederationService) cast_Org_User_UserType__to__Org_Federation_UserType(from user.UserType) UserType {
	switch from {
	case user.UserType_USER_TYPE_1:
		return UserType_USER_TYPE_1
	case user.UserType_USER_TYPE_2:
		return UserType_USER_TYPE_2
	default:
		return 0
	}
}

// cast_repeated_Org_User_Item__to__repeated_Org_Federation_Item cast from "repeated org.user.Item" to "repeated org.federation.Item".
func (s *FederationService) cast_repeated_Org_User_Item__to__repeated_Org_Federation_Item(from []*user.Item) []*Item {
	ret := make([]*Item, 0, len(from))
	for _, v := range from {
		ret = append(ret, s.cast_Org_User_Item__to__Org_Federation_Item(v))
	}
	return ret
}

// resolve_Org_Federation_Z resolve "org.federation.Z" message.
func (s *FederationService) resolve_Org_Federation_Z(ctx context.Context, req *Org_Federation_ZArgument) (*Z, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.Z", slog.Any("message_args", s.logvalue_Org_Federation_ZArgument(req)))

	// create a message value to be returned.
	// `custom_resolver = true` in "grpc.federation.message" option.
	ret, err := s.resolver.Resolve_Org_Federation_Z(ctx, req)
	if err != nil {
		return nil, err
	}

	s.logger.DebugContext(ctx, "resolved org.federation.Z", slog.Any("org.federation.Z", s.logvalue_Org_Federation_Z(ret)))
	return ret, nil
}

func (s *FederationService) logvalue_Google_Protobuf_Any(v *anypb.Any) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("type_url", v.GetTypeUrl()),
		slog.String("value", string(v.GetValue())),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponse(v *GetPostResponse) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("post", s.logvalue_Org_Federation_Post(v.GetPost())),
		slog.String("literal", v.GetLiteral()),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponseArgument(v *Org_Federation_GetPostResponseArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.Id),
	)
}

func (s *FederationService) logvalue_Org_Federation_Item(v *Item) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
		slog.String("type", s.logvalue_Org_Federation_Item_ItemType(v.GetType()).String()),
		slog.Int64("value", v.GetValue()),
	)
}

func (s *FederationService) logvalue_Org_Federation_Item_ItemType(v Item_ItemType) slog.Value {
	switch v {
	case Item_ITEM_TYPE_1:
		return slog.StringValue("ITEM_TYPE_1")
	case Item_ITEM_TYPE_2:
		return slog.StringValue("ITEM_TYPE_2")
	case Item_ITEM_TYPE_3:
		return slog.StringValue("ITEM_TYPE_3")
	}
	return slog.StringValue("")
}

func (s *FederationService) logvalue_Org_Federation_M(v *M) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("foo", v.GetFoo()),
		slog.Int64("bar", v.GetBar()),
	)
}

func (s *FederationService) logvalue_Org_Federation_MArgument(v *Org_Federation_MArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_Post(v *Post) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
		slog.String("title", v.GetTitle()),
		slog.String("content", v.GetContent()),
		slog.Any("user", s.logvalue_Org_Federation_User(v.GetUser())),
		slog.String("foo", v.GetFoo()),
		slog.Int64("bar", v.GetBar()),
	)
}

func (s *FederationService) logvalue_Org_Federation_PostArgument(v *Org_Federation_PostArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.Id),
	)
}

func (s *FederationService) logvalue_Org_Federation_User(v *User) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
		slog.String("type", s.logvalue_Org_Federation_UserType(v.GetType()).String()),
		slog.String("name", v.GetName()),
		slog.Uint64("age", v.GetAge()),
		slog.Any("desc", v.GetDesc()),
		slog.Any("main_item", s.logvalue_Org_Federation_Item(v.GetMainItem())),
		slog.Any("items", s.logvalue_repeated_Org_Federation_Item(v.GetItems())),
		slog.Any("profile", s.logvalue_Org_Federation_User_ProfileEntry(v.GetProfile())),
		slog.Any("attr_a", s.logvalue_Org_Federation_User_AttrA(v.GetAttrA())),
		slog.Any("b", s.logvalue_Org_Federation_User_AttrB(v.GetB())),
	)
}

func (s *FederationService) logvalue_Org_Federation_UserArgument(v *Org_Federation_UserArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.Id),
		slog.String("title", v.Title),
		slog.String("content", v.Content),
		slog.String("user_id", v.UserId),
	)
}

func (s *FederationService) logvalue_Org_Federation_UserType(v UserType) slog.Value {
	switch v {
	case UserType_USER_TYPE_1:
		return slog.StringValue("USER_TYPE_1")
	case UserType_USER_TYPE_2:
		return slog.StringValue("USER_TYPE_2")
	}
	return slog.StringValue("")
}

func (s *FederationService) logvalue_Org_Federation_User_AttrA(v *User_AttrA) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("foo", v.GetFoo()),
	)
}

func (s *FederationService) logvalue_Org_Federation_User_AttrB(v *User_AttrB) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Bool("bar", v.GetBar()),
	)
}

func (s *FederationService) logvalue_Org_Federation_User_ProfileEntry(v map[string]*anypb.Any) slog.Value {
	attrs := make([]slog.Attr, 0, len(v))
	for key, value := range v {
		attrs = append(attrs, slog.Attr{
			Key:   fmt.Sprint(key),
			Value: s.logvalue_Google_Protobuf_Any(value),
		})
	}
	return slog.GroupValue(attrs...)
}

func (s *FederationService) logvalue_Org_Federation_Z(v *Z) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("foo", v.GetFoo()),
	)
}

func (s *FederationService) logvalue_Org_Federation_ZArgument(v *Org_Federation_ZArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_repeated_Org_Federation_Item(v []*Item) slog.Value {
	attrs := make([]slog.Attr, 0, len(v))
	for idx, vv := range v {
		attrs = append(attrs, slog.Attr{
			Key:   fmt.Sprint(idx),
			Value: s.logvalue_Org_Federation_Item(vv),
		})
	}
	return slog.GroupValue(attrs...)
}
