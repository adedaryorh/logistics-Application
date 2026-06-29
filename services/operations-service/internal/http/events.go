package http

import (
	"context"
	"encoding/json"
	"fmt"

	platformkafka "github.com/adedaryorh/logistics-platform/pkg/kafka"
	operationsservice "github.com/adedaryorh/logistics-platform/services/operations-service/internal/service"
)

func handleUserRegisteredEvent(ctx context.Context, service *operationsservice.Service, value []byte) error {
	event, err := platformkafka.DecodeEvent(value)
	if err != nil {
		return err
	}
	var payload struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode user registered payload: %w", err)
	}
	return service.HandleUserRegistered(ctx, payload.UserID, payload.Email)
}

func handleOrderCreatedEvent(ctx context.Context, service *operationsservice.Service, value []byte) error {
	event, err := platformkafka.DecodeEvent(value)
	if err != nil {
		return err
	}
	var payload struct {
		OrderID    string `json:"order_id"`
		CustomerID string `json:"customer_id"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode order created payload: %w", err)
	}
	return service.HandleOrderCreated(ctx, payload.OrderID, payload.CustomerID)
}

func handleOrderAssignedEvent(ctx context.Context, service *operationsservice.Service, value []byte) error {
	event, err := platformkafka.DecodeEvent(value)
	if err != nil {
		return err
	}
	var payload struct {
		OrderID  string `json:"order_id"`
		DriverID string `json:"driver_id"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode order assigned payload: %w", err)
	}
	return service.HandleOrderAssigned(ctx, payload.OrderID, payload.DriverID)
}

func handleDispatchOfferedEvent(ctx context.Context, service *operationsservice.Service, value []byte) error {
	event, err := platformkafka.DecodeEvent(value)
	if err != nil {
		return err
	}
	var payload struct {
		OrderID       string `json:"order_id"`
		DriverID      string `json:"driver_id"`
		AttemptNumber int    `json:"attempt_number"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode dispatch offered payload: %w", err)
	}
	return service.HandleDispatchOffered(ctx, payload.OrderID, payload.DriverID, payload.AttemptNumber)
}

func handlePaymentFailedEvent(ctx context.Context, service *operationsservice.Service, value []byte) error {
	event, err := platformkafka.DecodeEvent(value)
	if err != nil {
		return err
	}
	var payload struct {
		OrderID string `json:"order_id"`
		Reason  string `json:"reason"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode payment failed payload: %w", err)
	}
	return service.HandlePaymentFailed(ctx, payload.OrderID, payload.Reason)
}

func handlePaymentCompletedEvent(ctx context.Context, service *operationsservice.Service, value []byte) error {
	event, err := platformkafka.DecodeEvent(value)
	if err != nil {
		return err
	}
	var payload struct {
		OrderID string `json:"order_id"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode payment completed payload: %w", err)
	}
	return service.HandlePaymentCompleted(ctx, payload.OrderID)
}

func handleOrderCancelledEvent(ctx context.Context, service *operationsservice.Service, value []byte) error {
	event, err := platformkafka.DecodeEvent(value)
	if err != nil {
		return err
	}
	var payload struct {
		OrderID string `json:"order_id"`
		Reason  string `json:"reason"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode order cancelled payload: %w", err)
	}
	return service.HandleOrderCancelled(ctx, payload.OrderID, payload.Reason)
}
