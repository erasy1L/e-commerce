package handler

import (
	"github.com/erazr/ecommerce-microservices/internal/common/server/response"
)

type request struct {
	UserID       string  `json:"user_id"`
	OrderID      string  `json:"order_id"`
	TotalPayment float64 `json:"total_payment"`
	PaymentDate  string  `json:"payment_date"`
	Status       string  `json:"status"`
} // @name PaymentRequest

func (req *request) Validate() []response.ErrorResponse {
	var errs []response.ErrorResponse

	if req.UserID == "" {
		errs = append(errs, response.ErrorResponse{
			Message: "empty user_id",
			Field:   "user_id",
		})
	}
	if req.OrderID == "" {
		errs = append(errs, response.ErrorResponse{
			Message: "empty order_id",
			Field:   "order_id",
		})
	}
	if req.TotalPayment == 0 {
		errs = append(errs, response.ErrorResponse{
			Message: "empty total_payment",
			Field:   "total_payment",
		})
	}
	if req.PaymentDate == "" {
		errs = append(errs, response.ErrorResponse{
			Message: "empty payment_date",
			Field:   "payment_date",
		})
	}
	if req.Status == "" {
		errs = append(errs, response.ErrorResponse{
			Message: "empty status",
			Field:   "status",
		})
	}

	return errs
}
