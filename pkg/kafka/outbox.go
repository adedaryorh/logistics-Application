package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/adedaryorh/logistics-platform/pkg/observability"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/segmentio/kafka-go"
)

// OutboxRelay relays messages from the outbox table to Kafka
type OutboxRelay struct {
	db           *sql.DB
	producer     messageWriter
	pollInterval time.Duration
	batchSize    int
	tableName    string
	topicPrefix  string
	producerName string
}

type messageWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

// NewOutboxRelay creates a new outbox relay
func NewOutboxRelay(db *sql.DB, producer *Producer, pollInterval time.Duration, batchSize int) *OutboxRelay {
	return &OutboxRelay{
		db:           db,
		producer:     producer,
		pollInterval: pollInterval,
		batchSize:    batchSize,
		tableName:    "outbox",
		topicPrefix:  "",
		producerName: "service",
	}
}

func (r *OutboxRelay) WithTableName(tableName string) *OutboxRelay {
	r.tableName = tableName
	return r
}

func (r *OutboxRelay) WithTopicPrefix(prefix string) *OutboxRelay {
	r.topicPrefix = prefix
	return r
}

func (r *OutboxRelay) WithProducerName(name string) *OutboxRelay {
	r.producerName = name
	return r
}

// Start begins polling the outbox table and publishing messages to Kafka
func (r *OutboxRelay) Start(ctx context.Context) error {
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := r.processBatch(ctx); err != nil {
				return fmt.Errorf("process outbox batch: %w", err)
			}
		}
	}
}

// processBatch fetches unpublished messages and publishes them
func (r *OutboxRelay) processBatch(ctx context.Context) error {
	start := time.Now()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		observability.IncCounter("outbox_batches_total", 1, map[string]string{"table": r.tableName, "status": "begin_failed"})
		return fmt.Errorf("begin outbox transaction: %w", err)
	}
	defer tx.Rollback()

	var gotLock bool
	if err := tx.QueryRowContext(ctx, "SELECT pg_try_advisory_xact_lock($1)", advisoryLockKey(r.tableName)).Scan(&gotLock); err != nil {
		observability.IncCounter("outbox_batches_total", 1, map[string]string{"table": r.tableName, "status": "lock_failed"})
		return fmt.Errorf("acquire outbox advisory lock: %w", err)
	}
	if !gotLock {
		observability.IncCounter("outbox_batches_total", 1, map[string]string{"table": r.tableName, "status": "skipped_locked"})
		return nil
	}

	query := `
		SELECT id, aggregate_id, event_type, payload
		FROM ` + r.tableName + `
		WHERE published_at IS NULL
		ORDER BY created_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`
	rows, err := tx.QueryContext(ctx, query, r.batchSize)
	if err != nil {
		return fmt.Errorf("query outbox rows: %w", err)
	}
	defer rows.Close()

	var ids []interface{}
	for rows.Next() {
		var (
			id          string
			aggregateID string
			eventType   string
			payload     []byte
		)
		if err := rows.Scan(&id, &aggregateID, &eventType, &payload); err != nil {
			return fmt.Errorf("scan outbox row: %w", err)
		}

		envelope, err := NewEventWithPayload(r.producerName, aggregateID, eventType, payload)
		if err != nil {
			return fmt.Errorf("build outbox event envelope: %w", err)
		}
		value, err := json.Marshal(envelope)
		if err != nil {
			return fmt.Errorf("marshal outbox event envelope: %w", err)
		}

		msg := kafka.Message{
			Topic: r.topicForEvent(eventType),
			Key:   []byte(aggregateID),
			Value: value,
			Time:  time.Now().UTC(),
		}
		if err := r.producer.WriteMessages(ctx, msg); err != nil {
			observability.IncCounter("outbox_events_total", 1, map[string]string{"table": r.tableName, "topic": msg.Topic, "status": "publish_failed"})
			return fmt.Errorf("publish outbox event: %w", err)
		}
		observability.IncCounter("outbox_events_total", 1, map[string]string{"table": r.tableName, "topic": msg.Topic, "status": "published"})
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate outbox rows: %w", err)
	}

	if len(ids) > 0 {
		updateQuery := `
			UPDATE ` + r.tableName + `
			SET published_at = NOW()
			WHERE id = ANY($1)
		`
		if _, err := tx.ExecContext(ctx, updateQuery, ids); err != nil {
			return fmt.Errorf("mark outbox rows published: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		observability.IncCounter("outbox_batches_total", 1, map[string]string{"table": r.tableName, "status": "commit_failed"})
		return fmt.Errorf("commit outbox transaction: %w", err)
	}
	observability.IncCounter("outbox_batches_total", 1, map[string]string{"table": r.tableName, "status": "success"})
	observability.SetGauge("outbox_batch_size", float64(len(ids)), map[string]string{"table": r.tableName})
	observability.ObserveHistogram("outbox_batch_duration_seconds", time.Since(start).Seconds(), map[string]string{"table": r.tableName})

	return nil
}

// Close closes the producer
func (r *OutboxRelay) Close() error {
	return r.producer.Close()
}

func advisoryLockKey(name string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	return int64(h.Sum64())
}

func (r *OutboxRelay) topicForEvent(eventType string) string {
	if r.topicPrefix == "" {
		return eventType
	}
	if strings.HasPrefix(eventType, r.topicPrefix) {
		return eventType
	}
	return r.topicPrefix + eventType
}
