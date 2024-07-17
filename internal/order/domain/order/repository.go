package order

import "context"

type Repository interface {
	Search(ctx context.Context, filter, value string) ([]Order, error)
	List(ctx context.Context) ([]Order, error)
	Get(ctx context.Context, id string) (Order, error)
	Create(ctx context.Context, order Order) (string, error)
	Update(ctx context.Context, id string, order Order) error
	Delete(ctx context.Context, id string) error
}
