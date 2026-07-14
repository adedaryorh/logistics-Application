package main

import (
	"log"

	"github.com/adedaryorh/logistics-platform/pkg/config"
	"github.com/adedaryorh/logistics-platform/pkg/servicehttp"
	identityhttp "github.com/adedaryorh/logistics-platform/services/identity-service/internal/http"
)

func main() {
	if err := servicehttp.Run(servicehttp.Options{
		ServiceName: "identity-service",
		GRPCPort:    "9081",
		NewApp:      func(cfg *config.Config) servicehttp.App { return identityhttp.NewApp(cfg) },
	}); err != nil {
		log.Fatal(err)
	}
}
