package web

import (
	"github.com/sgladkov/tortuga/internal/service"
	"net/http"

	"github.com/go-chi/chi"
)

var marketplace *service.Marketplace
var address string

func mock(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func TortugaRouter(m *service.Marketplace, a string) chi.Router {
	marketplace = m
	address = a
	r := chi.NewRouter()
	r.Middlewares()
	r.Use(RequestLogger)
	r.Use(GzipHandle)
	r.Get("/", mock)
	r.Route("/api/public/", func(r chi.Router) {
		r.Get("/config", configInfo)
		r.Get("/user_list", userList)
		r.Get("/user/{id}", userInfo)
		r.Get("/user/{id}/history", userHistory)
		r.Get("/project_list", projectList)
		r.Get("/project/{id}", projectInfo)
		r.Get("/project/{id}/bids", projectBids)
		r.Get("/bid/{id}", bidInfo)
	})
	r.Route("/api/", func(r chi.Router) {
		r.Use(RegistrationHandle)
		r.Post("/register", register)
	})
	r.Route("/api/private/", func(r chi.Router) {
		r.Use(AuthorizationHandle)
		r.Post("/create_project", createProject)
		r.Post("/project/{id}/create_bid", createBid)
		r.Post("/bid/{id}/accept_bid", acceptBid)
		r.Post("/bid/{id}/update_bid", updateBid)
		r.Post("/bid/{id}/delete_bid", deleteBid)
		r.Post("/project/{id}/update", updateProject)
		r.Post("/project/{id}/delete", deleteProject)
		r.Post("/project/{id}/cancel", cancelProject)
		r.Post("/project/{id}/ready", readyProject)
		r.Post("/project/{id}/accept", acceptProject)
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
