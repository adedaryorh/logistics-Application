package otel

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func InitTracer(serviceName, jaegerEndpoint string) (func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if serviceName == "" {
		serviceName = "unknown-service"
	}

	trimmedEndpoint := strings.TrimPrefix(strings.TrimPrefix(jaegerEndpoint, "http://"), "https://")
	trimmedEndpoint = strings.TrimSuffix(trimmedEndpoint, "/")
	trimmedEndpoint = strings.TrimSuffix(trimmedEndpoint, "/v1/traces")

	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(trimmedEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	exp, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resourceNew(serviceName)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagationConfig())

	return func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = tp.Shutdown(shutdownCtx)
	}, nil
}

func resourceNew(serviceName string) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		attribute.String("environment", getEnv("ENVIRONMENT", "development")),
		attribute.String("host.id", getEnv("HOSTNAME", "")),
	)
}

func SpanFromContext(ctx context.Context) oteltrace.Span {
	return oteltrace.SpanFromContext(ctx)
}

func propagationConfig() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
