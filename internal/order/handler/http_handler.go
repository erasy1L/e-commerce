package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/erazr/ecommerce-microservices/internal/common/pb/product"
	"github.com/erazr/ecommerce-microservices/internal/common/server/response"
	"github.com/erazr/ecommerce-microservices/internal/common/store"
	"github.com/erazr/ecommerce-microservices/internal/order/domain/order"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type OrderHandler struct {
	repo order.Repository

	idempotencyCache store.Cache[order.Order]

	productGRPCService product.ProductsClient
}

func NewOrderHandler(repo order.Repository, configs ...func(h *OrderHandler)) OrderHandler {
	h := OrderHandler{
		repo: repo,
	}

	for _, cfg := range configs {
		cfg(&h)
	}

	return h
}

func WithIdempotencyCache(c store.Cache[order.Order]) func(h *OrderHandler) {
	return func(h *OrderHandler) {
		h.idempotencyCache = c
	}
}

func WithProductGRPCService(s product.ProductsClient) func(h *OrderHandler) {
	return func(h *OrderHandler) {
		h.productGRPCService = s
	}
}

func (h *OrderHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.placeOrder)
	r.Get("/", h.listOrders)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.getOrder)
		r.Put("/", h.updateOrder)
		r.Delete("/", h.deleteOrder)
	})

	r.Get("/search", h.Search)

	return r
}

// @Summary		Place an order
// @Description	place an order
// @Tags			orders
// @Accept			json
// @Produce		json
// @Param			body	body		request	true	"request"
// @Success		200		{string}	string	"order id"
// @Failure		400		{array}		response.ErrorResponse
// @Failure		500
// @Router			/orders [post]
func (h *OrderHandler) placeOrder(w http.ResponseWriter, r *http.Request) {
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

	idempotencyKey := h.generateIdempotencyKey(req)

	stored, has := h.idempotencyCache.Get(idempotencyKey)
	if has {
		render.PlainText(w, r, stored.ID)
		return
	}

	// Check if products are available
	availabes, err := h.productGRPCService.ProductsAvailable(r.Context(), &product.ProductsAvailableRequest{
		ProductIds: req.ProductID,
	})
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	for _, p := range availabes.GetAvailability() {
		if !p.Available {
			response.BadRequest(w, r, []response.ErrorResponse{{
				Message: fmt.Sprintf("product %s with id: %s is not available, stock is %d", p.GetName(), p.GetProductId(), p.GetStock()),
				Field:   "product_id",
			}})
			return
		}
	}

	// Get product prices
	prices, err := h.productGRPCService.GetProductPrices(r.Context(), &product.GetProductPricesRequest{
		ProductIds: req.ProductID,
	})
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	o := order.Order{
		ID:          store.GenerateID(),
		UserID:      req.UserID,
		ProductID:   req.ProductID,
		OrderedDate: store.OnlyDate(req.OrderedDate),
		Status:      req.Status,
	}

	// Calculate total price
	for _, price := range prices.GetPrices() {
		o.TotalPrice += float64(price)
	}

	id, err := h.repo.Create(r.Context(), o)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	h.idempotencyCache.Set(idempotencyKey, o)

	render.PlainText(w, r, id)
}

// @Summary		List orders
// @Description	list all orders
// @Tags			orders
// @Accept			json
// @Produce		json
// @Success		200	{array}	order.Order
// @Failure		500
// @Router			/orders [get]
func (h *OrderHandler) listOrders(w http.ResponseWriter, r *http.Request) {
	products, err := h.repo.List(r.Context())
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	render.JSON(w, r, products)
}

// @Summary		Get order
// @Description	Get order by id
// @Tags			orders
// @Accept			json
// @Produce		json
// @Param			id	path		string	true	"order id"
// @Success		200	{object}	order.Order
// @Failure		400	{array}		response.ErrorResponse
// @Failure		404	{string}	string
// @Failure		500
// @Router			/orders/{id} [get]
func (h *OrderHandler) getOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	o, err := h.repo.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, order.ErrNotFound) {
			response.NotFound(w, r, err)
			return
		}

		response.InternalServerError(w, r, err)
		return
	}

	render.JSON(w, r, o)
}

// @Summary		Update order
// @Description	Update order by id
// @Tags			orders
// @Accept			json
// @Produce		json
// @Param			id	path	string	true	"order id"
// @Success		200
// @Failure		400	{array}		response.ErrorResponse
// @Failure		404	{string}	string
// @Failure		500
// @Router			/orders/{id} [put]
func (h *OrderHandler) updateOrder(w http.ResponseWriter, r *http.Request) {
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

	err := h.repo.Update(r.Context(), id, order.Order{
		ProductID: req.ProductID,
		// TotalPrice:  req.TotalPrice,
		OrderedDate: store.OnlyDate(req.OrderedDate),
		Status:      req.Status,
	})
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
}

// @Summary		Delete order
// @Description	Delete order by id
// @Tags			orders
// @Accept			json
// @Produce		json
// @Param			id	path	string	true	"order id"
// @Success		200
// @Failure		400	{array}		response.ErrorResponse
// @Failure		404	{string}	string
// @Failure		500
// @Router			/orders/{id} [delete]
func (h *OrderHandler) deleteOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.repo.Delete(r.Context(), id)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
}

// @Summary		Search order
// @Description	Search order by user id and status
// @Tags			orders
// @Accept			json
// @Produce		json
// @Param			user	query	string	false	"user id"
// @Param			status	query	string	false	"status"
// @Success		200
// @Failure		400	{array}		response.ErrorResponse
// @Failure		404	{string}	string
// @Failure		500
// @Router			/orders/search [get]
func (h *OrderHandler) Search(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("user")
	status := r.URL.Query().Get("status")

	products, err := h.repo.Search(r.Context(), productID, status)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	render.JSON(w, r, products)
}

func (h *OrderHandler) generateIdempotencyKey(req request) string {
	// Normalize the ProductID slice by sorting it
	sort.Strings(req.ProductID)

	keyData := fmt.Sprintf("%s|%s|%s", req.UserID, req.ProductID, req.OrderedDate)

	hash := sha256.Sum256([]byte(keyData))

	return hex.EncodeToString(hash[:])
}
