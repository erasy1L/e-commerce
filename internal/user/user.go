package main

import (
	"context"

	"github.com/erazr/ecommerce-microservices/internal/common/store"
)

type User struct {
	ID               string
	Name             string
	Email            string
	RegistrationDate store.OnlyDate `db:"registration_date"`
	Role             string
} // @name User

var (
	ErrExists   = &UserError{"user already exists"}
	ErrNotFound = &UserError{"user not found"}
	ErrSearch   = &UserError{"user search error"}
)

type UserError struct {
	message string
}

func (e *UserError) Error() string {
	return e.message
}

func (e *UserError) Is(err error) bool {
	return e == err
}

type repository interface {
	List(ctx context.Context) ([]User, error)
	Search(ctx context.Context, filter, value string) ([]User, error)
	Create(context.Context, User) (string, error)
	Get(ctx context.Context, id string) (User, error)
	Update(ctx context.Context, id string, u User) error
	Delete(ctx context.Context, id string) error
}
