package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erazr/ecommerce-microservices/internal/common/server/response"
	"github.com/erazr/ecommerce-microservices/internal/common/store"
	"github.com/erazr/ecommerce-microservices/internal/product/domain/product"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type ProductHandler struct {
	repo product.Repository
}

func NewProductHandler(repo product.Repository) ProductHandler {
	return ProductHandler{repo: repo}
}

func (h *ProductHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.listProducts)
	r.Post("/", h.createProduct)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.getProduct)
		r.Put("/", h.updateProduct)
		r.Delete("/", h.deleteProduct)
	})

	r.Get("/search", h.searchProduct)

	return r
}

//	@Summary		update products
//	@Description	update product by id
//	@Tags			products
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string	true	"Product ID"
//	@Param			body	body	request	true	"Product data"
//	@Success		200
//	@Failure		500
//	@Failure		404	{string}	string
//	@Router			/products/{id} [put]
func (h *ProductHandler) updateProduct(w http.ResponseWriter, r *http.Request) {
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

	err := h.repo.Update(r.Context(), id, product.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		Amount:      req.Amount,
		AddedAt:     store.OnlyDate(req.AddedAt),
	})
	if err != nil {
		if errors.Is(err, product.ErrNotFound) {
			response.NotFound(w, r, err)
			return
		}

		response.InternalServerError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)

}

//	@Summary		Create product
//	@Description	create product
//	@Tags			products
//	@Accept			json
//	@Produce		json
//	@Param			body	body		request	true	"Product data"
//	@Success		200		{string}	string	"Product ID"
//	@Failure		400		{array}		response.ErrorResponse
//	@Failure		500
//	@Router			/products [post]
func (h *ProductHandler) createProduct(w http.ResponseWriter, r *http.Request) {
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

	id, err := h.repo.Create(r.Context(), product.Product{
		ID:          store.GenerateID(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		Amount:      req.Amount,
		AddedAt:     store.OnlyDate(req.AddedAt),
	})

	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	render.PlainText(w, r, id)
}

//	@Summary		Delete product
//	@Description	delete product
//	@Tags			products
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Product ID"
//	@Success		200
//	@Failure		404	{string}	string
//	@Failure		500
//	@Router			/products/{id} [delete]
func (h *ProductHandler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.repo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, product.ErrNotFound) {
			response.NotFound(w, r, err)
			return
		}

		response.InternalServerError(w, r, err)
		return
	}
}

//	@Summary		Get product
//	@Description	get product
//	@Tags			products
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Product ID"
//	@Success		200	{object}	product.Product
//	@Failure		404	{string}	string
//	@Failure		500
//	@Router			/products/{id} [get]
func (h *ProductHandler) getProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	p, err := h.repo.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, product.ErrNotFound) {
			response.NotFound(w, r, err)
			return
		}

		response.InternalServerError(w, r, err)
		return
	}

	render.JSON(w, r, p)
}

//	@Summary		Search product
//	@Description	search product
//	@Tags			products
//	@Accept			json
//	@Produce		json
//	@Param			name		query		string	false	"Product name"
//	@Param			category	query		string	false	"Product category"
//	@Success		200			{array}		product.Product
//	@Failure		404			{string}	string
//	@Failure		500
//	@Router			/products/search [get]
func (h *ProductHandler) searchProduct(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	category := r.URL.Query().Get("category")

	if name != "" {
		p, err := h.repo.Search(r.Context(), "name", name)
		if err != nil {
			if errors.Is(err, product.ErrNotFound) {
				response.NotFound(w, r, err)
				return
			}

			response.InternalServerError(w, r, err)
			return
		}

		render.JSON(w, r, p)
		return
	}

	if category != "" {
		p, err := h.repo.Search(r.Context(), "category", category)
		if err != nil {
			if errors.Is(err, product.ErrNotFound) {
				response.NotFound(w, r, err)
				return
			}

			response.InternalServerError(w, r, err)
			return
		}

		render.JSON(w, r, p)
		return
	}
}

//	@Summary		List products
//	@Description	list all products
//	@Tags			products
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	product.Product
//	@Failure		500
//	@Router			/products [get]
func (h *ProductHandler) listProducts(w http.ResponseWriter, r *http.Request) {
	p, err := h.repo.List(r.Context())
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	render.JSON(w, r, p)
}
