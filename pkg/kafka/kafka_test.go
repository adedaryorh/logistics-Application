package kafka

import "testing"

func TestNewProducer(t *testing.T) {
	t.Parallel()

	producer := NewProducer([]string{"localhost:9092"})
	if producer == nil {
		t.Fatal("expected producer")
	}
	if producer.w == nil {
		t.Fatal("expected writer on producer")
	}
}

func TestNewConsumer(t *testing.T) {
	t.Parallel()

	consumer := NewConsumer([]string{"localhost:9092"})
	if consumer == nil {
		t.Fatal("expected consumer")
	}
	if len(consumer.brokers) != 1 {
		t.Fatalf("expected 1 broker, got %d", len(consumer.brokers))
	}
}
