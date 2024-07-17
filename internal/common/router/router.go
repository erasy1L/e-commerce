package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)

	r.Use(middleware.RealIP)

	r.Use(middleware.Logger)

	r.Use(middleware.Recoverer)

	r.Use(middleware.CleanPath)

	r.Use(middleware.Heartbeat("/healthcheck"))

	// r.Use(cors.Handler(cors.Options{
	// 	Debug:            true,
	// 	AllowCredentials: true,
	// }))

	return r
}
