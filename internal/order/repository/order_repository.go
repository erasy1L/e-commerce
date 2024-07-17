package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/erazr/ecommerce-microservices/internal/common/store"
	"github.com/erazr/ecommerce-microservices/internal/order/domain/order"
	"github.com/jmoiron/sqlx"
)

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, o order.Order) (string, error) {
	err := r.db.QueryRowContext(ctx, "INSERT INTO orders (id, user_id, ordered_date, total_price, status) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		o.ID,
		o.UserID,
		o.OrderedDate,
		o.TotalPrice,
		o.Status,
	).Scan(&o.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", order.ErrNotFound
		}

		return "", err
	}

	ids := o.TotalAmount()

	for productID, amount := range ids {
		_, err = r.db.ExecContext(ctx, "INSERT INTO order_products (order_id, product_id, amount) VALUES ($1, $2, $3)", o.ID, productID, amount)
		if err != nil {
			return "", err
		}
	}

	return o.ID, err
}

func (r *OrderRepository) Get(ctx context.Context, id string) (order.Order, error) {
	var o order.Order

	query := `
        SELECT 
            o.id, o.user_id, o.total_price, o.ordered_date, o.status, 
            op.product_id, op.amount
        FROM orders o
        JOIN order_products op ON o.id = op.order_id
        WHERE o.id = $1
    `

	rows, err := r.db.QueryxContext(ctx, query, id)
	if err != nil {
		return order.Order{}, err
	}
	defer rows.Close()

	var (
		orderID     string
		userID      string
		productID   string
		status      string
		totalPrice  float64
		orderedDate store.OnlyDate
		amount      int
	)

	o.ProductID = []string{}

	for rows.Next() {
		err := rows.Scan(&orderID, &userID, &totalPrice, &orderedDate, &status, &productID, &amount)
		if err != nil {
			return order.Order{}, err
		}

		if o.ID == "" {
			o = order.Order{
				ID:          orderID,
				UserID:      userID,
				TotalPrice:  totalPrice,
				OrderedDate: orderedDate,
				Status:      status,
				ProductID:   []string{},
			}
		}

		for i := 0; i < amount; i++ {
			o.AddProduct(productID)
		}
	}

	if o.ID == "" {
		return order.Order{}, order.ErrNotFound
	}

	return o, nil
}

func (r *OrderRepository) Delete(ctx context.Context, id string) error {
	err := r.db.QueryRowContext(ctx, "DELETE FROM orders WHERE id = $1 RETURNING id", id).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return order.ErrNotFound
		}

		return err
	}

	return err
}

func (r *OrderRepository) List(ctx context.Context) ([]order.Order, error) {
	var orders []order.Order

	query := `
        SELECT 
            o.id, o.user_id, o.total_price, o.ordered_date, o.status, 
            op.product_id, op.amount
        FROM orders o
        JOIN order_products op ON o.id = op.order_id
    `

	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderMap := make(map[string]*order.Order)

	for rows.Next() {
		var id, userID, productID, status string
		var totalPrice float64
		var orderedDate store.OnlyDate
		var amount int

		err := rows.Scan(&id, &userID, &totalPrice, &orderedDate, &status, &productID, &amount)
		if err != nil {
			return nil, err
		}

		if _, ok := orderMap[id]; !ok {
			orderMap[id] = &order.Order{
				ID:          id,
				UserID:      userID,
				TotalPrice:  totalPrice,
				OrderedDate: orderedDate,
				Status:      status,
				ProductID:   []string{},
			}
		}
		for i := 0; i < amount; i++ {
			orderMap[id].AddProduct(productID)
		}
	}

	for _, order := range orderMap {
		orders = append(orders, *order)
	}

	return orders, nil
}

func (r *OrderRepository) Update(ctx context.Context, id string, o order.Order) error {
	q := "UPDATE orders SET user_id = $1, total_price = $2, ordered_date = $3, status = $4 WHERE id = $5 RETURNING id"

	args := []any{o.UserID, o.TotalPrice, o.OrderedDate, o.Status, id}

	fmt.Println(o)

	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return order.ErrNotFound
		}

		return err
	}

	fmt.Println(o, "updated")

	return nil
}

func (r *OrderRepository) Search(ctx context.Context, filter, value string) ([]order.Order, error) {
	var orders []order.Order

	query := `
        SELECT 
            o.id, o.user_id, o.total_price, o.ordered_date, o.status, 
            op.product_id
        FROM orders o
        JOIN order_products op ON o.id = op.order_id
		WHERE $1 = $2
    `

	rows, err := r.db.QueryxContext(ctx, query, filter, value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderMap := make(map[string]*order.Order)

	for rows.Next() {
		var id, userID, productID, status string
		var totalPrice float64
		var orderedDate store.OnlyDate

		err := rows.Scan(&id, &userID, &totalPrice, &orderedDate, &status, &productID)
		if err != nil {
			return nil, err
		}

		if _, ok := orderMap[id]; !ok {
			orderMap[id] = &order.Order{
				ID:          id,
				UserID:      userID,
				TotalPrice:  totalPrice,
				OrderedDate: orderedDate,
				Status:      status,
				ProductID:   []string{},
			}
		}
		orderMap[id].ProductID = append(orderMap[id].ProductID, productID)
	}

	for _, order := range orderMap {
		orders = append(orders, *order)
	}

	return orders, nil
}

// func (r *OrderRepository) prepareArgs(order Order) (sets []string, args []any) {
// 	if order.ProductID != nil {
// 		args = append(args, order.ProductID)
// 		sets = append(sets, fmt.Sprintf("product_id = $%d", len(args)))
// 	}
// 	if order.TotalPrice != 0 {
// 		args = append(args, order.TotalPrice)
// 		sets = append(sets, fmt.Sprintf("total_price = $%d", len(args)))
// 	}
// 	if order.OrderedDate != "" {
// 		args = append(args, order.OrderedDate)
// 		sets = append(sets, fmt.Sprintf("ordered_date = $%d", len(args)))
// 	}
// 	if order.Status != "" {
// 		args = append(args, order.Status)
// 		sets = append(sets, fmt.Sprintf("status = $%d", len(args)))
// 	}

// 	return
// }
