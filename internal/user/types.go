package main

import (
	"regexp"
	"time"

	"github.com/erazr/ecommerce-microservices/internal/common/server/response"
	"github.com/erazr/ecommerce-microservices/internal/common/store"
)

type request struct {
	Name             string `json:"name"`
	Email            string `json:"email"`
	Role             string `json:"role"`
	RegistrationDate string `json:"registration_date"`
} // @name UserRequest

func (u *request) validate() []response.ErrorResponse {
	var errs []response.ErrorResponse

	if u.Name == "" {
		errs = append(errs, response.ErrorResponse{Message: "name is required", Field: "name"})
	}

	if ok, _ := regexp.MatchString(`^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$`, u.Email); !ok {
		errs = append(errs, response.ErrorResponse{Message: "invalid email address", Field: "email"})
	}

	if u.Role != "admin" && u.Role != "client" {
		errs = append(errs, response.ErrorResponse{Message: "invalid role", Field: "role"})
	}

	if _, err := time.Parse(store.DateLayout, u.RegistrationDate); err != nil {
		errs = append(errs, response.ErrorResponse{Message: "invalid registration_date format", Field: "registration_date"})
	}

	return errs
}
