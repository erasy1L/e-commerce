package product

import (
	"context"

	"github.com/erazr/ecommerce-microservices/internal/common/store"
)

type Product struct {
	ID          string         `db:"id" json:"id"`
	Name        string         `db:"name" json:"name"`
	Description string         `db:"description" json:"description"`
	Price       float64        `db:"price" json:"price"`
	Category    string         `db:"category" json:"category"`
	Amount      int            `db:"amount" json:"amount"`
	AddedAt     store.OnlyDate `db:"added_at"`
} // @name Product

var (
	ErrExists             = &ProductError{"product already exists"}
	ErrNotFound           = &ProductError{"product not found"}
	ErrSearch             = &ProductError{"product search error"}
	ErrInsufficientAmount = &ProductError{"insufficient amount"}
)

type ProductError struct {
	message string
}

func (e *ProductError) Error() string {
	return e.message
}

func (e *ProductError) Is(err error) bool {
	return e == err
}

type Repository interface {
	List(ctx context.Context) ([]Product, error)
	Search(ctx context.Context, filter, value string) ([]Product, error)
	Get(ctx context.Context, id string) (Product, error)
	GetPriceByID(ctx context.Context, id string) (float64, error)
	Create(ctx context.Context, p Product) (string, error)
	Update(ctx context.Context, id string, p Product) error
	Delete(ctx context.Context, id string) error
}
