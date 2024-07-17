package main

import (
	"context"
	"errors"

	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type userRepository struct {
	db *sqlx.DB
}

func newUserRepository(db *sqlx.DB) *userRepository {
	if db == nil {
		panic("db is required")
	}

	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, u User) (id string, err error) {
	q := `
		INSERT INTO users (id, name, email, registration_date, role)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`

	args := []any{u.ID, u.Name, u.Email, u.RegistrationDate, u.Role}

	err = r.db.QueryRowContext(ctx, q, args...).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrExists
		}
		if err, ok := err.(*pq.Error); ok && err.Code.Name() == "unique_violation" {
			return "", ErrExists
		}
		return
	}

	return
}

func (r *userRepository) Update(ctx context.Context, id string, u User) (err error) {
	q := "UPDATE users SET name = $1, email = $2, role = $3 WHERE id = $4 RETURNING id"

	args := []any{u.Name, u.Email, u.Role, id}

	err = r.db.QueryRowContext(ctx, q, args...).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrNotFound
		}
	}

	return
}

func (r *userRepository) Get(ctx context.Context, id string) (u User, err error) {
	u = User{}

	q := `
	SELECT * FROM users WHERE id = $1
	`

	if err = r.db.GetContext(ctx, &u, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrNotFound
			return
		}
	}

	return
}

func (r *userRepository) Delete(ctx context.Context, id string) (err error) {
	q := `
	DELETE FROM users WHERE id = $1 RETURNING id
	`

	if err = r.db.QueryRowContext(ctx, q, id).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrNotFound
			return
		}
	}

	return
}

func (r *userRepository) List(ctx context.Context) (users []User, err error) {
	users = []User{}

	q := "SELECT * FROM users"

	err = r.db.SelectContext(ctx, &users, q)
	if err != nil {
		return
	}

	return
}

func (r *userRepository) Search(ctx context.Context, filter, value string) (users []User, err error) {
	users = []User{}

	q := "SELECT * FROM users WHERE $1 = $2"

	err = r.db.SelectContext(ctx, &users, q, value)
	if err != nil {
		return
	}

	if len(users) == 0 {
		err = ErrNotFound
		return
	}

	return
}
