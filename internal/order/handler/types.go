package handler

import (
	"time"

	"github.com/erazr/ecommerce-microservices/internal/common/server/response"
	"github.com/erazr/ecommerce-microservices/internal/common/store"
)

type request struct {
	UserID    string   `json:"user_id"`
	ProductID []string `json:"product_id"`
	// TotalPrice  float64
	OrderedDate string `json:"ordered_date"`
	Status      string `json:"status"`
} // @name OrderRequest

func (r request) Validate() []response.ErrorResponse {
	var errs []response.ErrorResponse

	if r.ProductID == nil {
		errs = append(errs, response.ErrorResponse{
			Message: "ProductID is required",
			Field:   "product_id",
		})
	}

	if _, err := time.Parse(store.DateLayout, r.OrderedDate); err != nil {
		errs = append(errs, response.ErrorResponse{
			Message: "Invalid date format",
			Field:   "ordered_date",
		})
	}

	if r.Status == "" {
		errs = append(errs, response.ErrorResponse{
			Message: "Status is required",
			Field:   "status",
		})
	}

	return errs
}
