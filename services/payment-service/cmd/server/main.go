package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"github.com/adedaryorh/logistics-platform/pkg/config"
	"github.com/adedaryorh/logistics-platform/pkg/servicehttp"
	paymenthttp "github.com/adedaryorh/logistics-platform/services/payment-service/internal/http"
)

func main() {
	var app *paymenthttp.App
	if err := servicehttp.Run(servicehttp.Options{
		ServiceName: "payment-service",
		GRPCPort:    "9084",
		Setup: func(cfg *config.Config) (func(*gin.Engine), func(context.Context) error, error) {
			app = paymenthttp.NewApp(cfg)
			return app.RegisterRoutes, func(ctx context.Context) error {
				return app.StartBackground(ctx, cfg)
			}, nil
		},
		RegisterGRPC: func(server *grpc.Server, cfg *config.Config) {
			if app == nil {
				app = paymenthttp.NewApp(cfg)
			}
			app.RegisterGRPC(server, cfg)
		},
	}); err != nil {
		log.Fatal(err)
	}
}
