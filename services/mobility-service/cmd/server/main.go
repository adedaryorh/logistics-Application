package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"github.com/adedaryorh/logistics-platform/pkg/config"
	"github.com/adedaryorh/logistics-platform/pkg/servicehttp"
	mobilityhttp "github.com/adedaryorh/logistics-platform/services/mobility-service/internal/http"
)

func main() {
	var app *mobilityhttp.App
	if err := servicehttp.Run(servicehttp.Options{
		ServiceName: "mobility-service",
		GRPCPort:    "9083",
		Setup: func(cfg *config.Config) (func(*gin.Engine), func(context.Context) error, error) {
			app = mobilityhttp.NewApp(cfg)
			return app.RegisterRoutes, func(ctx context.Context) error {
				return app.StartBackground(ctx, cfg)
			}, nil
		},
		RegisterGRPC: func(server *grpc.Server, cfg *config.Config) {
			if app == nil {
				app = mobilityhttp.NewApp(cfg)
			}
			app.RegisterGRPC(server, cfg)
		},
	}); err != nil {
		log.Fatal(err)
	}
}
