package main

import (
	"net"
	"os"
	"os/signal"

	"github.com/0proto/metacrawl/services"
	httpGtw "github.com/0proto/metacrawl/transformers/gateways/http"
	httpGtwCtrls "github.com/0proto/metacrawl/transformers/gateways/http/controllers"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("failed to initialize logger")
	}

	metaCrawlSvc := services.NewMetaCrawl(logger)
	listener, err := net.Listen(
		"tcp",
		"0.0.0.0:8080",
	)

	httpGtw := httpGtw.NewGateway(
		listener,
		httpGtw.GatewayWithControllers(
			httpGtwCtrls.NewV1(metaCrawlSvc),
		),
	)

	err = httpGtw.Start()
	if err != nil {
		panic("can't start http gateway")
	}

	logger.Info("application started.")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	logger.Info("termination signal received, shutting down gracefully...")
	httpGtw.Stop()
	return
}
