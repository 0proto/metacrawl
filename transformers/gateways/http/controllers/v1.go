package controllers

import (
	"github.com/0proto/metacrawl/services"
	"github.com/go-chi/chi"
)

// NewV1 creates a new instance of V1.
func NewV1(metaCrawlService services.MetaCrawlSvc) (ctrl *V1) {
	ctrl = &V1{
		metaCrawlSvc: metaCrawlService,
	}
	return
}

// V1 is a HTTP gateway controller that is responsible for processing calls to Data Loader v1 endpoints
type V1 struct {
	metaCrawlSvc services.MetaCrawlSvc
}

// Register sets up controller routes in the router.
func (ctrl *V1) Register(router chi.Router) {
	router.Get("/tasks/{taskID}/", ctrl.GetTask)
	router.Post("/tasks/", ctrl.PostTask)
	//router.Get("/status/{region}/{summonerName}", ctrl.StatusSummoner)
}

// httpError represents an HTTP error.
type httpError struct {
	Error error `json:"error"`
}
