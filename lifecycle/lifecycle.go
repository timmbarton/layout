package lifecycle

import "context"

type Lifecycle interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetName() string
}
