package main

import (
	"log"

	"github.com/adedaryorh/logistics-platform/pkg/config"
	"github.com/adedaryorh/logistics-platform/pkg/servicehttp"
	paymenthttp "github.com/adedaryorh/logistics-platform/services/payment-service/internal/http"
)

func main() {
	if err := servicehttp.Run(servicehttp.Options{
		ServiceName: "payment-service",
		GRPCPort:    "9084",
		NewApp:      func(cfg *config.Config) servicehttp.App { return paymenthttp.NewApp(cfg) },
	}); err != nil {
		log.Fatal(err)
	}
}
