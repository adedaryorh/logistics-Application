package kafka

import (
	"context"
	"testing"

	segmentkafka "github.com/segmentio/kafka-go"
)

type fakeWriter struct {
	msgs []segmentkafka.Message
}

func (f *fakeWriter) WriteMessages(ctx context.Context, msgs ...segmentkafka.Message) error {
	f.msgs = append(f.msgs, msgs...)
	return nil
}

func (f *fakeWriter) Close() error { return nil }

func TestTopicForEvent(t *testing.T) {
	relay := (&OutboxRelay{topicPrefix: "logistics."})
	if got := relay.topicForEvent("order.created"); got != "logistics.order.created" {
		t.Fatalf("expected prefixed topic, got %q", got)
	}
	if got := relay.topicForEvent("logistics.order.created"); got != "logistics.order.created" {
		t.Fatalf("expected already-prefixed topic unchanged, got %q", got)
	}
}
