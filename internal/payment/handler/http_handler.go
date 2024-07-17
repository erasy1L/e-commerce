package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/erazr/ecommerce-microservices/internal/common/pb/order"
	"github.com/erazr/ecommerce-microservices/internal/common/pb/product"
	"github.com/erazr/ecommerce-microservices/internal/common/server/response"
	"github.com/erazr/ecommerce-microservices/internal/common/store"
	"github.com/erazr/ecommerce-microservices/internal/payment/domain/payment"
	"github.com/erazr/ecommerce-microservices/internal/payment/epay"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type PaymentHandler struct {
	repo payment.Repository

	ePayService epay.Service

	idempotencyCache store.Cache[payment.Payment]

	orderGRPCService   order.OrdersClient
	productGRPCService product.ProductsClient
}

func NewPaymentHandler(repo payment.Repository, ePayService epay.Service, configs ...func(*PaymentHandler)) *PaymentHandler {
	h := &PaymentHandler{
		repo:        repo,
		ePayService: ePayService,
	}

	for _, config := range configs {
		config(h)
	}

	return h
}

func WithIdempotencyCache(c store.Cache[payment.Payment]) func(*PaymentHandler) {
	return func(h *PaymentHandler) {
		h.idempotencyCache = c
	}
}

func WithOrderGRPCService(s order.OrdersClient) func(*PaymentHandler) {
	return func(h *PaymentHandler) {
		h.orderGRPCService = s
	}
}

func WithProductGRPCService(s product.ProductsClient) func(*PaymentHandler) {
	return func(h *PaymentHandler) {
		h.productGRPCService = s
	}
}

func (h *PaymentHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListPayment)
	r.Post("/", h.MakePayment)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetPayment)
		r.Put("/", h.UpdatePayment)
		r.Delete("/", h.DeletePayment)
	})

	r.Get("/search", h.SearchPayment)

	return r
}

// @Summary		Make payment
// @Description	make a payment
// @Tags			payments
// @Accept			json
// @Produce		json
// @Param			body	body					request	true	"request"
// @Success		200		epay.PaymentResponse	"payment response"
// @Failure		400		{array}					response.ErrorResponse
// @Failure		500
// @Router			/payments [post]
func (h *PaymentHandler) MakePayment(w http.ResponseWriter, r *http.Request) {
	req := request{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, []response.ErrorResponse{{
			Message: err.Error(),
			Field:   "body",
		}})
	}

	if errs := req.Validate(); errs != nil {
		response.BadRequest(w, r, errs)
		return
	}

	if val, has := h.idempotencyCache.Get(req.OrderID); has {
		render.PlainText(w, r, val.ID)
		return
	}

	payment := payment.Payment{
		ID:           store.GenerateID(),
		UserID:       req.UserID,
		OrderID:      req.OrderID,
		TotalPayment: req.TotalPayment,
		PaymentDate:  store.OnlyDate(req.PaymentDate),
		Status:       req.Status,
	}

	token, err := h.ePayService.Token()
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	id, err := h.repo.Create(r.Context(), payment)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	h.idempotencyCache.Set(req.OrderID, payment)

	time.Sleep(150 * time.Millisecond)
	_, err = h.orderGRPCService.UpdateOrderStatus(r.Context(), &order.UpdateOrderStatusRequest{
		OrderId: payment.OrderID,
		Status:  "pending",
	})
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	h.simulateOrderProccess(w, r, payment)

	h.ePayService.Pay(token)

	render.PlainText(w, r, id)
}

// @Summary		Get payment
// @Description	get a payment
// @Tags			payments
// @Accept			json
// @Produce		json
// @Param			id	path		string	true	"payment id"
// @Success		200	{object}	payment.Payment
// @Failure		404	{string}	string
// @Failure		500
// @Router			/payments/{id} [get]
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	p, err := h.repo.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, payment.ErrNotFound) {
			response.NotFound(w, r, err)
			return
		}

		response.NotFound(w, r, err)
		return
	}

	response.OK(w, r, p)
}

// @Summary		Update payment
// @Description	update a payment
// @Tags			payments
// @Accept			json
// @Produce		json
// @Param			id		path	string	true	"payment id"
// @Param			body	body	request	true	"request"
// @Success		200
// @Failure		400	{array}		response.ErrorResponse
// @Failure		404	{string}	string
// @Failure		500
// @Router			/payments/{id} [put]
func (h *PaymentHandler) UpdatePayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := request{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, []response.ErrorResponse{{
			Message: err.Error(),
			Field:   "body",
		}})
		return
	}

	if errs := req.Validate(); errs != nil {
		response.BadRequest(w, r, errs)
		return
	}

	p := payment.Payment{
		UserID:       req.UserID,
		OrderID:      req.OrderID,
		TotalPayment: req.TotalPayment,
		PaymentDate:  store.OnlyDate(req.PaymentDate),
		Status:       req.Status,
	}

	if err := h.repo.Update(r.Context(), id, p); err != nil {
		if errors.Is(err, payment.ErrNotFound) {
			response.NotFound(w, r, err)
			return
		}

		response.InternalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary		Delete payment
// @Description	delete a payment
// @Tags			payments
// @Accept			json
// @Produce		json
// @Param			id	path	string	true	"payment id"
// @Success		200
// @Failure		404	{string}	string
// @Failure		500
// @Router			/payments/{id} [delete]
func (h *PaymentHandler) DeletePayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, payment.ErrNotFound) {
			response.NotFound(w, r, err)
			return
		}

		response.InternalServerError(w, r, err)
		return
	}
}

// @Summary		List payment
// @Description	list all payments
// @Tags			payments
// @Accept			json
// @Produce		json
// @Success		200	{array}	payment.Payment
// @Failure		500
// @Router			/payments [get]
func (h *PaymentHandler) ListPayment(w http.ResponseWriter, r *http.Request) {
	payments, err := h.repo.List(r.Context())
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.OK(w, r, payments)
}

// @Summary		Search payments
// @Description	Search payments by user, order, or status
// @Tags			payments
// @Accept			json
// @Produce		json
// @Param			user	path		string	false	"user id"
// @Param			order	path		string	false	"order id"
// @Param			status	path		string	false	"status"
// @Success		200		{array}		payment.Payment
// @Failure		404		{string}	string
// @Failure		400		{array}		response.ErrorResponse
// @Failure		500
// @Router			/payments/search [get]
func (h *PaymentHandler) SearchPayment(w http.ResponseWriter, r *http.Request) {
	user := chi.URLParam(r, "user")
	order := chi.URLParam(r, "order")
	status := chi.URLParam(r, "status")

	if user == "" && order == "" && status == "" {
		response.BadRequest(w, r, []response.ErrorResponse{{
			Message: "empty query",
			Field:   "query",
		}})
		return
	}

	if user != "" {
		payments, err := h.repo.Search(r.Context(), "user_id", user)
		if err != nil {
			if errors.Is(err, payment.ErrNotFound) {
				response.NotFound(w, r, err)
				return
			}

			response.InternalServerError(w, r, err)
			return
		}
		response.OK(w, r, payments)
	}
	if order != "" {
		payments, err := h.repo.Search(r.Context(), "order_id", order)
		if err != nil {
			if errors.Is(err, payment.ErrNotFound) {
				response.NotFound(w, r, err)
				return
			}

			response.InternalServerError(w, r, err)
			return
		}
		response.OK(w, r, payments)
	}
	if status != "" {
		payments, err := h.repo.Search(r.Context(), "status", status)
		if err != nil {
			if errors.Is(err, payment.ErrNotFound) {
				response.NotFound(w, r, err)
				return
			}

			response.InternalServerError(w, r, err)
			return
		}
		response.OK(w, r, payments)
	}

}

func (h *PaymentHandler) simulateOrderProccess(w http.ResponseWriter, r *http.Request, payment payment.Payment) {
	resp, err := h.orderGRPCService.GetOrderProductIDs(r.Context(), &order.GetOrderProductIDsRequest{
		OrderId: payment.OrderID,
	})
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	updates := []*product.UpdateProduct{}

	mapProductIDs := make(map[string]int)

	for _, productID := range resp.ProductIds {
		mapProductIDs[productID]++
	}

	for productID, amount := range mapProductIDs {
		updates = append(updates, &product.UpdateProduct{
			ProductId:  productID,
			Quantity:   int32(amount),
			UpdateType: product.UpdateType_DECREMENT,
		})
	}

	_, err = h.productGRPCService.UpdateProductStock(r.Context(), &product.UpdateProductStockRequest{
		Updates: updates,
	})
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	time.Sleep(150 * time.Millisecond)
	_, err = h.orderGRPCService.UpdateOrderStatus(context.Background(), &order.UpdateOrderStatusRequest{
		OrderId: payment.OrderID,
		Status:  "completed",
	})
	if err != nil {
		return
	}
}
