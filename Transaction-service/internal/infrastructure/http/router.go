package http

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func NewRouter(accountHandler http.Handler, transactionHandler http.Handler) http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/accounts", accountHandler)
		r.Mount("/transactions", transactionHandler)
	})

	return r
} 