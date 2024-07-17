package response

import (
	"net/http"

	"github.com/go-chi/render"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Field   string `json:"field"`
} // @name ErrorResponse

type Object struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func OK(w http.ResponseWriter, r *http.Request, data any) {
	render.Status(r, http.StatusOK)

	render.JSON(w, r, data)
}

func BadRequest(w http.ResponseWriter, r *http.Request, errs []ErrorResponse) {
	render.Status(r, http.StatusBadRequest)

	render.JSON(w, r, errs)
}

func NotFound(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, http.StatusNotFound)

	render.PlainText(w, r, err.Error())
}

func InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, http.StatusInternalServerError)

	render.PlainText(w, r, err.Error())
}
