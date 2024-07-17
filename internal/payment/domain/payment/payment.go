package payment

import (
	"context"

	"github.com/erazr/ecommerce-microservices/internal/common/store"
)

type Payment struct {
	ID           string         `db:"id" json:"id"`
	UserID       string         `db:"user_id" json:"user_id"`
	OrderID      string         `db:"order_id" json:"order_id"`
	TotalPayment float64        `db:"total_payment" json:"total_payment"`
	PaymentDate  store.OnlyDate `db:"payment_date" json:"payment_date"`
	Status       string         `db:"status" json:"status"`
} // @name Payment

var (
	ErrExists             = &PaymentError{"payment already exists"}
	ErrNotFound           = &PaymentError{"payment not found"}
	ErrSearch             = &PaymentError{"payment search error"}
	ErrInsufficientAmount = &PaymentError{"insufficient amount"}
)

type PaymentError struct {
	message string
}

func (e *PaymentError) Error() string {
	return e.message
}

func (e *PaymentError) Is(err error) bool {
	return e == err
}

type Repository interface {
	Create(context.Context, Payment) (string, error)
	Get(ctx context.Context, id string) (Payment, error)
	Update(ctx context.Context, id string, payment Payment) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]Payment, error)
	Search(ctx context.Context, filter, value string) ([]Payment, error)
}
