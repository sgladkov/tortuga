package web

import (
	storage2 "github.com/sgladkov/tortuga/internal/storage"

	"github.com/go-chi/chi"
)

var storage storage2.Storage
var address string

func TortugaRouter(s storage2.Storage, a string) chi.Router {
	storage = s
	address = a
	r := chi.NewRouter()
	r.Middlewares()
	r.Use(RequestLogger)
	r.Use(GzipHandle)
	r.Use(AuthorizationHandle)
	r.Get("/", mock)
	r.Route("/api/public/", func(r chi.Router) {
		r.Get("/config", configInfo)
		r.Get("/user_list", userList)
		r.Get("/user/{id}", userInfo)
		r.Get("/user/{id}/history", userHistory)
		r.Get("/project_list", projectList)
		r.Get("/project/{id}", projectInfo)
	})
	r.Route("/api/private/", func(r chi.Router) {
		r.Post("/register", register)
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
