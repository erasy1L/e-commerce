package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/erazr/ecommerce-microservices/internal/product/domain/product"
	"github.com/jmoiron/sqlx"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	if db == nil {
		panic("db is required")
	}

	return &ProductRepository{
		db: db,
	}
}

func (r *ProductRepository) List(ctx context.Context) (products []product.Product, err error) {
	products = []product.Product{}

	err = r.db.SelectContext(ctx, &products, "SELECT * FROM products")
	if err != nil {
		return
	}

	return
}

func (r *ProductRepository) Get(ctx context.Context, id string) (p product.Product, err error) {
	p = product.Product{}

	err = r.db.GetContext(ctx, &p, "SELECT * FROM products WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return p, product.ErrNotFound
		}
		return
	}

	return
}

func (r *ProductRepository) GetPriceByID(ctx context.Context, id string) (float64, error) {
	var price float64

	err := r.db.GetContext(ctx, &price, "SELECT price FROM products WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, product.ErrNotFound
		}

		return 0, err
	}

	return price, nil
}

func (r *ProductRepository) Create(ctx context.Context, p product.Product) (string, error) {
	q := "INSERT INTO products (id, name, description, price, category, amount, added_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id"

	args := []any{p.ID, p.Name, p.Description, p.Price, p.Category, p.Amount, p.AddedAt}

	err := r.db.QueryRowContext(ctx, q, args...).Scan(&p.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", product.ErrNotFound
		}

		return p.ID, err
	}

	return p.ID, nil
}

func (r *ProductRepository) Update(ctx context.Context, id string, p product.Product) error {
	q := `
		UPDATE products SET name = $1, description = $2, price = $3, category = $4, amount = $5, added_at = $6
		WHERE id = $7 RETURNING id
	`

	args := []any{p.Name, p.Description, p.Price, p.Category, p.Amount, p.AddedAt, id}

	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return product.ErrNotFound
		}
	}

	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	err := r.db.QueryRowContext(ctx, "DELETE FROM products WHERE id = $1 RETURNING ID", id).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return product.ErrNotFound
		}

		return err
	}

	return nil
}

func (r *ProductRepository) Search(ctx context.Context, filter string, value string) (products []product.Product, err error) {
	products = []product.Product{}

	err = r.db.SelectContext(ctx, &products, "SELECT * FROM products WHERE $1 = $2", filter, value)
	if err != nil {
		return
	}

	if len(products) == 0 {
		err = product.ErrNotFound
	}

	return
}
