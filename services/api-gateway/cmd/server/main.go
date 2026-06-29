package main

import (
	"log"

	"github.com/adedaryorh/logistics-platform/pkg/servicehttp"
	gatewayhttp "github.com/adedaryorh/logistics-platform/services/api-gateway/internal/http"
)

func main() {
	if err := servicehttp.Run(servicehttp.Options{
		ServiceName: "api-gateway",
		Register:    gatewayhttp.RegisterRoutes,
	}); err != nil {
		log.Fatal(err)
	}
}
