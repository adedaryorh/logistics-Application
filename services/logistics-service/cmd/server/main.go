package main

import (
	"log"

	"github.com/adedaryorh/logistics-platform/pkg/config"
	"github.com/adedaryorh/logistics-platform/pkg/servicehttp"
	logisticshttp "github.com/adedaryorh/logistics-platform/services/logistics-service/internal/http"
)

func main() {
	if err := servicehttp.Run(servicehttp.Options{
		ServiceName: "logistics-service",
		GRPCPort:    "9082",
		NewApp:      func(cfg *config.Config) servicehttp.App { return logisticshttp.NewApp(cfg) },
	}); err != nil {
		log.Fatal(err)
	}
}
