package handler

import (
	"github.com/erazr/ecommerce-microservices/internal/common/server/response"
)

type request struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Amount      int     `json:"amount"`
	AddedAt     string  `json:"added_at"`
} // @name ProductRequest

func (r *request) Validate() []response.ErrorResponse {
	var errs []response.ErrorResponse

	if r.Name == "" {
		errs = append(errs, response.ErrorResponse{
			Message: "name is required",
			Field:   "name",
		})
	}

	if r.Price <= 0 {
		errs = append(errs, response.ErrorResponse{
			Message: "price must be greater than 0",
			Field:   "price",
		})
	}

	if r.Category == "" {
		errs = append(errs, response.ErrorResponse{
			Message: "category is required",
			Field:   "category",
		})
	}

	if r.Amount <= 0 {
		errs = append(errs, response.ErrorResponse{
			Message: "amount must be greater than 0",
			Field:   "amount",
		})
	}

	return errs
}
