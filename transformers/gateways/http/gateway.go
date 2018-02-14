package http

import (
	"net/http"

	"net"

	"context"

	"time"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// Middleware represents HTTP middleware that can be attached to middleware chain.
type Middleware func(http.Handler) http.Handler

// Controller represents a group of related HTTP handlers.
type Controller interface {
	// Register sets up controller routes in the router.
	Register(router chi.Router)
}

// NewGateway creates a new instance if Gateway.
func NewGateway(listener net.Listener, options ...GatewayOption) (gtw *Gateway) {

	// Default values
	gtw = &Gateway{
		Listener:   listener,
		Logger:     zap.NewNop(),
		httpServer: &http.Server{},
	}

	// Apply custom options
	for _, option := range options {
		option(gtw)
	}

	return
}

// Gateway is a Gateway interface implementation that is using HTTP as a transport layer.
type Gateway struct {
	Listener net.Listener // listener to bind too
	Logger   *zap.Logger  // logger instance

	middlewares []Middleware // middlewares to use
	controllers []Controller // controllers to use
	httpServer  *http.Server // HTTP server instance
}

// Start tells gateway to start listening for incoming requests.
func (gtw *Gateway) Start() (err error) {

	router := chi.NewRouter()

	// Register middlewares
	for _, middleware := range gtw.middlewares {
		router.Use(middleware)
	}

	// Register controllers
	for _, controller := range gtw.controllers {
		controller.Register(router)
	}

	// Start server
	gtw.httpServer.Handler = router
	go func() {
		err = gtw.httpServer.Serve(gtw.Listener)
	}()
	time.Sleep(time.Second)

	return
}

// Stop tells gateway to stop listening for incoming requests and exit gracefully.
func (gtw *Gateway) Stop() {
	err := gtw.httpServer.Shutdown(context.Background())
	if err != nil {
		gtw.Logger.Warn("HTTP server shutdown failed.", zap.Error(err))
	}
}

// GatewayOption represents a custom gateway option.
type GatewayOption func(gtw *Gateway)

// GatewayWithLogger sets a custom logger instance to use by this gateway.
func GatewayWithLogger(logger *zap.Logger) (option GatewayOption) {
	return func(gtw *Gateway) {
		gtw.Logger = logger
	}
}

// GatewayWithMiddlewares sets global middlewares that will be used by all endpoints.
func GatewayWithMiddlewares(middlewares ...Middleware) (option GatewayOption) {
	return func(gtw *Gateway) {
		gtw.middlewares = append(gtw.middlewares, middlewares...)
	}
}

// GatewayWithControllers sets controllers to use by this gateway.
func GatewayWithControllers(controllers ...Controller) (option GatewayOption) {
	return func(gtw *Gateway) {
		gtw.controllers = append(gtw.controllers, controllers...)
	}
}
