package main

import (
	"log"

	"github.com/adedaryorh/logistics-platform/pkg/config"
	"github.com/adedaryorh/logistics-platform/pkg/servicehttp"
	mobilityhttp "github.com/adedaryorh/logistics-platform/services/mobility-service/internal/http"
)

func main() {
	if err := servicehttp.Run(servicehttp.Options{
		ServiceName: "mobility-service",
		GRPCPort:    "9083",
		NewApp:      func(cfg *config.Config) servicehttp.App { return mobilityhttp.NewApp(cfg) },
	}); err != nil {
		log.Fatal(err)
	}
}
