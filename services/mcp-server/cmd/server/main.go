package main

import (
	"log"

	"github.com/adedaryorh/logistics-platform/pkg/servicehttp"
	mcphttp "github.com/adedaryorh/logistics-platform/services/mcp-server/internal/http"
)

func main() {
	if err := servicehttp.Run(servicehttp.Options{
		ServiceName: "mcp-server",
		Register:    mcphttp.RegisterRoutes,
	}); err != nil {
		log.Fatal(err)
	}
}
