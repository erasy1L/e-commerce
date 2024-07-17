package order

import "github.com/erazr/ecommerce-microservices/internal/common/store"

type Order struct {
	ID          string         `db:"id" json:"id"`
	UserID      string         `db:"user_id" json:"user_id"`
	ProductID   []string       `db:"-" json:"product_id"`
	TotalPrice  float64        `db:"total_price" json:"total_price"`
	OrderedDate store.OnlyDate `db:"ordered_date" json:"ordered_date"`
	Status      string         `db:"status" json:"status"`
} // @name Order

var (
	ErrExists   = &OrderError{"order already exists"}
	ErrNotFound = &OrderError{"order not found"}
	ErrSearch   = &OrderError{"order search error"}
)

type OrderError struct {
	message string
}

func (e *OrderError) Error() string {
	return e.message
}

func (e *OrderError) Is(err error) bool {
	return e == err
}

func (o *Order) AddProduct(productID string) {
	o.ProductID = append(o.ProductID, productID)
}

func (o *Order) RemoveProduct(productID string) {
	for i, id := range o.ProductID {
		if id == productID {
			o.ProductID = append(o.ProductID[:i], o.ProductID[i+1:]...)
			return
		}
	}
}

func (o *Order) TotalAmount() map[string]int {
	totals := map[string]int{}

	for _, id := range o.ProductID {
		totals[id]++
	}

	return totals
}
