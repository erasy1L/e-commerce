package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erazr/ecommerce-microservices/internal/common/server/response"
	"github.com/erazr/ecommerce-microservices/internal/common/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type UserHandler struct {
	repo repository
}

func newUserHandler(repo repository) *UserHandler {
	if repo == nil {
		panic("repo is required")
	}

	return &UserHandler{
		repo: repo,
	}
}

func (h *UserHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.create)

	r.Get("/search", h.search)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
	})

	return r
}

// @Summary		Get users
// @Description	list all users
// @Tags			users
// @Accept			json
// @Produce		json
// @Success		200	{array}	User
// @Failure		500
// @Router			/users [get]
func (h *UserHandler) list(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.List(r.Context())
	if err != nil {
		response.InternalServerError(w, r, err)
	}

	response.OK(w, r, users)
}

// @Summary		Create user
// @Description	create a new user
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			body	body		request	true	"User data"
// @Success		200		{string}	string	"user id"
// @Failure		400		{array}		response.ErrorResponse
// @Failure		500
// @Router			/users [post]
func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {
	req := request{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, []response.ErrorResponse{{
			Message: err.Error(),
			Field:   "body",
		}})
		return
	}

	if errs := req.validate(); errs != nil {
		response.BadRequest(w, r, errs)
		return
	}

	u := User{
		ID:               store.GenerateID(),
		Name:             req.Name,
		Email:            req.Email,
		RegistrationDate: store.OnlyDate(req.RegistrationDate),
		Role:             req.Role,
	}

	id, err := h.repo.Create(r.Context(), u)
	if err != nil {
		response.InternalServerError(w, r, err)
	}

	render.PlainText(w, r, id)
}

// @Summary		Update user
// @Description	update user
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			id		path	string	true	"User ID"
// @Param			body	body	request	true	"User data"
// @Success		200
// @Failure		400	{array}	response.ErrorResponse
// @Failure		500
// @Failure		404	{string}	string
// @Router			/users/{id} [put]
func (h *UserHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := request{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, []response.ErrorResponse{{
			Message: err.Error(),
			Field:   "body",
		}})
		return
	}

	if errs := req.validate(); errs != nil {
		response.BadRequest(w, r, errs)
		return
	}

	u := User{
		Name:             req.Name,
		Email:            req.Email,
		RegistrationDate: store.OnlyDate(req.RegistrationDate),
		Role:             req.Role,
	}

	err := h.repo.Update(r.Context(), id, u)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, r, err)
			return
		}

		response.InternalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary		Get user
// @Description	get user by ID
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			id	path		string	true	"User ID"
// @Success		200	{object}	User
// @Failure		404	{string}	string
// @Failure		500
// @Router			/users/{id} [get]
func (h *UserHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	u, err := h.repo.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, r, err)
			return
		}

		response.InternalServerError(w, r, err)
		return
	}

	response.OK(w, r, u)
}

// @Summary		Delete user
// @Description	delete user by ID
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			id	path	string	true	"User ID"
// @Success		200
// @Failure		500
// @Failure		404	{string}	string
// @Router			/users/{id} [delete]
func (h *UserHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.repo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, r, err)
			return
		}

		response.InternalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary		Search users
// @Description	search users by name or email
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			name	query		string	false	"User name"
// @Param			email	query		string	false	"User email"
// @Success		200		{array}		User
// @Failure		404		{string}	string
// @Failure		500
// @Failure		400	{array}	response.ErrorResponse
func (h *UserHandler) search(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	email := r.URL.Query().Get("email")

	if name == "" && email == "" {
		response.BadRequest(w, r, []response.ErrorResponse{{
			Message: "empty query",
			Field:   "query",
		}})
		return
	}

	if name != "" {
		users, err := h.repo.Search(r.Context(), "name", name)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				response.NotFound(w, r, err)
				return
			}

			response.InternalServerError(w, r, err)
			return
		}

		response.OK(w, r, users)
	}

	if email != "" {
		users, err := h.repo.Search(r.Context(), "email", email)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				response.NotFound(w, r, err)
				return
			}

			response.InternalServerError(w, r, err)
			return
		}

		response.OK(w, r, users)
	}

}
