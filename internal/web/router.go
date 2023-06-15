package web

import (
	"github.com/go-chi/chi"
)

func TortugaRouter() chi.Router {
	r := chi.NewRouter()
	r.Middlewares()
	r.Use(RequestLogger)
	r.Use(GzipHandle)
	r.Get("/", mock)
	r.Route("/api/public/", func(r chi.Router) {
		r.Get("/config", mock)
		r.Get("/user_list", mock)
		r.Get("/user/{id}", mock)
		r.Get("/user/{id}/history", mock)
		r.Get("/project_list", mock)
		r.Get("/project/{id}", mock)
	})
	r.Route("/api/private/", func(r chi.Router) {
		r.Post("/register", mock)
		r.Post("/create_project", mock)
		r.Get("/project/{id}", mock)
		r.Post("/project/{id}/accept_bid", mock)
		r.Post("/project/{id}/update", mock)
		r.Post("/project/{id}/delete", mock)
		r.Post("/project/{id}/cancel_work", mock)
		r.Post("/project/{id}/rate_work", mock)
		r.Post("/project/{id}/rate_owner", mock)
		r.Post("/user/{id}/request_mentorship", mock)
		r.Post("/user/{id}/accept_mentorship", mock)
		r.Post("/user/{id}/cancel_mentorship", mock)
		r.Post("/user/{id}/rate_mentor", mock)
		r.Post("/user/{id}/rate_student", mock)
		r.Post("/user/{id}/account", mock)
		r.Post("/user/{id}/account/withdraw", mock)
		r.Get("/user/{id}/account/history", mock)
	})
	return r
}
