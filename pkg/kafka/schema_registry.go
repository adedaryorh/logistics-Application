package kafka

import (
	"encoding/json"
	"fmt"
	"strings"
)

type EventContract struct {
	SchemaVersion int
	Producer      string
	RequiredKeys  []string
}

var defaultContracts = map[string]EventContract{
	"user.registered":      {SchemaVersion: 1, Producer: "identity-service", RequiredKeys: []string{"user_id", "email", "role"}},
	"order.created":        {SchemaVersion: 1, Producer: "logistics-service", RequiredKeys: []string{"order_id", "customer_id", "type", "status", "price_minor", "workflow_id", "idempotency_key"}},
	"order.cancelled":      {SchemaVersion: 1, Producer: "logistics-service", RequiredKeys: []string{"order_id", "reason"}},
	"order.assigned":       {SchemaVersion: 1, Producer: "logistics-service", RequiredKeys: []string{"order_id", "driver_id"}},
	"order.payment_failed": {SchemaVersion: 1, Producer: "logistics-service", RequiredKeys: []string{"order_id", "reason"}},
	"dispatch.offered":     {SchemaVersion: 1, Producer: "logistics-service", RequiredKeys: []string{"order_id", "driver_id", "attempt_number"}},
	"dispatch.accepted":    {SchemaVersion: 1, Producer: "logistics-service", RequiredKeys: []string{"order_id", "assignment_id", "driver_id", "attempt_number"}},
	"dispatch.rejected":    {SchemaVersion: 1, Producer: "logistics-service", RequiredKeys: []string{"order_id", "assignment_id", "driver_id", "attempt_number"}},
	"driver.online":        {SchemaVersion: 1, Producer: "logistics-service", RequiredKeys: []string{"driver_id", "user_id", "online"}},
	"driver.offline":       {SchemaVersion: 1, Producer: "logistics-service", RequiredKeys: []string{"driver_id", "user_id", "online"}},
	"payment.initialized":  {SchemaVersion: 1, Producer: "payment-service", RequiredKeys: []string{"transaction_id", "order_id", "amount_minor", "currency", "provider"}},
	"payment.completed":    {SchemaVersion: 1, Producer: "payment-service", RequiredKeys: []string{"transaction_id", "order_id", "amount_minor", "currency"}},
	"payment.failed":       {SchemaVersion: 1, Producer: "payment-service", RequiredKeys: []string{"transaction_id", "order_id", "reason"}},
	"payment.refunded":     {SchemaVersion: 1, Producer: "payment-service", RequiredKeys: []string{"transaction_id", "order_id", "amount_minor"}},
	"payment.cancelled":    {SchemaVersion: 1, Producer: "payment-service", RequiredKeys: []string{"transaction_id", "order_id", "reason"}},
	"wallet.credited":      {SchemaVersion: 1, Producer: "payment-service", RequiredKeys: []string{"user_id", "amount_minor", "currency", "ref_id"}},
}

func ValidateEventContract(event Event) error {
	contract, ok := defaultContracts[event.EventType]
	if !ok {
		return fmt.Errorf("unknown event contract %q", event.EventType)
	}
	if event.SchemaVersion != contract.SchemaVersion {
		return fmt.Errorf("event %q has schema version %d, expected %d", event.EventType, event.SchemaVersion, contract.SchemaVersion)
	}
	if strings.TrimSpace(event.Producer) != contract.Producer {
		return fmt.Errorf("event %q has producer %q, expected %q", event.EventType, event.Producer, contract.Producer)
	}

	var payload map[string]any
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode event payload for contract validation: %w", err)
	}
	for _, key := range contract.RequiredKeys {
		if _, ok := payload[key]; !ok {
			return fmt.Errorf("event %q missing required payload key %q", event.EventType, key)
		}
	}
	return nil
}
