package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/erazr/ecommerce-microservices/internal/payment/domain/payment"
	"github.com/jmoiron/sqlx"
)

type PaymentRepository struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) *PaymentRepository {
	if db == nil {
		panic("db is required")
	}

	return &PaymentRepository{
		db: db,
	}
}

func (r *PaymentRepository) Create(ctx context.Context, payment payment.Payment) (string, error) {
	q := `INSERT INTO payments (id, user_id, order_id, total_payment, payment_date, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err := r.db.QueryRowContext(ctx, q,
		payment.ID,
		payment.UserID,
		payment.OrderID,
		payment.TotalPayment,
		payment.PaymentDate,
		payment.Status,
	).Scan(&payment.ID)
	if err != nil {
		return "", err
	}

	return payment.ID, nil
}

func (r *PaymentRepository) Get(ctx context.Context, id string) (payment.Payment, error) {
	p := payment.Payment{}

	err := r.db.GetContext(ctx, &p, "SELECT * FROM payments WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return p, payment.ErrNotFound
		}
		return p, err
	}

	return p, nil
}

func (r *PaymentRepository) Update(ctx context.Context, id string, p payment.Payment) error {
	q := `UPDATE payments SET user_id=$1, order_id=$2, total_payment=$3, payment_date=$4, status=$5 WHERE id=$6 RETURNING id`

	err := r.db.QueryRowContext(ctx, q,
		p.UserID,
		p.OrderID,
		p.TotalPayment,
		p.PaymentDate,
		p.Status,
		id,
	).Scan(&p.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return payment.ErrNotFound
		}

		return err
	}

	return nil
}

func (r *PaymentRepository) Delete(ctx context.Context, id string) error {
	err := r.db.QueryRowContext(ctx, "DELETE FROM payments WHERE id = $1 RETURNING id", id).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return payment.ErrNotFound
		}

		return err
	}

	return nil
}

func (r *PaymentRepository) List(ctx context.Context) ([]payment.Payment, error) {
	payments := []payment.Payment{}

	err := r.db.SelectContext(ctx, &payments, "SELECT * FROM payments")
	if err != nil {
		return payments, err
	}

	return payments, nil
}

func (r *PaymentRepository) Search(ctx context.Context, filter, value string) ([]payment.Payment, error) {
	payments := []payment.Payment{}

	err := r.db.SelectContext(ctx, &payments, "SELECT * FROM payments WHERE $1 = $2", filter, value)
	if err != nil {
		return payments, err
	}

	if len(payments) == 0 {
		return payments, payment.ErrNotFound
	}

	return payments, nil
}
