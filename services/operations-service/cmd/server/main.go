package main

import (
	"log"

	"github.com/adedaryorh/logistics-platform/pkg/config"
	"github.com/adedaryorh/logistics-platform/pkg/servicehttp"
	operationshttp "github.com/adedaryorh/logistics-platform/services/operations-service/internal/http"
)

func main() {
	if err := servicehttp.Run(servicehttp.Options{
		ServiceName: "operations-service",
		GRPCPort:    "9085",
		NewApp:      func(cfg *config.Config) servicehttp.App { return operationshttp.NewApp(cfg) },
	}); err != nil {
		log.Fatal(err)
	}
}
