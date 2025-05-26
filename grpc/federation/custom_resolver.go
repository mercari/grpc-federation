package federation

import "context"

type CustomResolverInitializer interface {
	Init(ctx context.Context) error
}
