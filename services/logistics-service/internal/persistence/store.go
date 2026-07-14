package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/adedaryorh/logistics-platform/services/logistics-service/internal/model"
	_ "github.com/lib/pq"
)

type Store interface {
	Ping(context.Context) error
	SaveQuote(context.Context, *model.AgriculturalQuote, string) error
	SaveBooking(context.Context, *model.Order, string, string) error
	EnqueueWebhook(context.Context, string, string, []byte) error
	ClaimWebhooks(context.Context, int) ([]PendingWebhook, error)
	CompleteWebhook(context.Context, string, error, int) error
	Close() error
}

type MemoryStore struct{}

func (MemoryStore) Ping(context.Context) error { return nil }

type PendingWebhook struct {
	ID, EventType string
	Payload       []byte
	Attempts      int
}

func (MemoryStore) SaveQuote(context.Context, *model.AgriculturalQuote, string) error { return nil }
func (MemoryStore) SaveBooking(context.Context, *model.Order, string, string) error   { return nil }
func (MemoryStore) EnqueueWebhook(context.Context, string, string, []byte) error      { return nil }
func (MemoryStore) ClaimWebhooks(context.Context, int) ([]PendingWebhook, error)      { return nil, nil }
func (MemoryStore) CompleteWebhook(context.Context, string, error, int) error         { return nil }
func (MemoryStore) Close() error                                                      { return nil }

type SQLStore struct{ db *sql.DB }

func NewSQLStore(dsn string) (*SQLStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	return &SQLStore{db}, nil
}
func (s *SQLStore) Close() error                   { return s.db.Close() }
func (s *SQLStore) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
func (s *SQLStore) SaveQuote(ctx context.Context, q *model.AgriculturalQuote, key string) error {
	pickup, _ := json.Marshal(q.Pickup)
	dropoff, _ := json.Marshal(q.Dropoff)
	shipment, _ := json.Marshal(q.Shipment)
	_, err := s.db.ExecContext(ctx, `INSERT INTO logistics_.agricultural_quotes(id,platform_service,platform_user_id,farmsense_request_id,marketplace_request_id,idempotency_key,pickup,dropoff,shipment,price_minor,currency,expires_at,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) ON CONFLICT(platform_service,idempotency_key) DO NOTHING`, q.ID, q.PlatformService, q.PlatformUserID, q.FarmSenseRequestID, q.MarketplaceRequestID, key, pickup, dropoff, shipment, q.PriceMinor, q.Currency, q.ExpiresAt, q.CreatedAt)
	return err
}
func (s *SQLStore) SaveBooking(ctx context.Context, o *model.Order, service, key string) error {
	shipment, _ := json.Marshal(o.AgriculturalShipment)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO logistics_.orders(id,customer_id,type,status,pickup_lat,pickup_lng,pickup_h3,dropoff_lat,dropoff_lng,dropoff_h3,pickup_address,dropoff_address,price_minor,currency,surge_multiplier,idempotency_key,temporal_workflow_id,platform_user_id,source_platform_service,farmsense_request_id,marketplace_request_id,agricultural_shipment,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24) ON CONFLICT(idempotency_key) DO NOTHING`, o.ID, o.CustomerID, o.Type, o.Status, o.Pickup.Lat, o.Pickup.Lng, o.Pickup.H3Cell, o.Dropoff.Lat, o.Dropoff.Lng, o.Dropoff.H3Cell, o.Pickup.Address, o.Dropoff.Address, o.PriceMinor, o.Currency, o.SurgeMultiplier, o.IdempotencyKey, o.TemporalWorkflowID, o.PlatformUserID, o.SourcePlatformService, o.FarmSenseRequestID, o.MarketplaceRequestID, shipment, o.CreatedAt, o.UpdatedAt)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO logistics_.agricultural_booking_idempotency(platform_service,idempotency_key,order_id) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, service, key, o.ID)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (s *SQLStore) EnqueueWebhook(ctx context.Context, orderID, eventType string, payload []byte) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO logistics_.outbound_status_webhooks(id,order_id,event_type,payload) VALUES(gen_random_uuid(),$1,$2,$3)`, orderID, eventType, payload)
	return err
}
func (s *SQLStore) ClaimWebhooks(ctx context.Context, limit int) ([]PendingWebhook, error) {
	rows, err := s.db.QueryContext(ctx, `WITH claimed AS (
        SELECT id FROM logistics_.outbound_status_webhooks
        WHERE delivered_at IS NULL AND dead_lettered_at IS NULL AND next_attempt_at<=NOW()
        ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT $1
    ) UPDATE logistics_.outbound_status_webhooks w
      SET next_attempt_at=NOW()+INTERVAL '60 seconds'
      FROM claimed WHERE w.id=claimed.id
      RETURNING w.id,w.event_type,w.payload,w.attempts`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []PendingWebhook
	for rows.Next() {
		var item PendingWebhook
		if err := rows.Scan(&item.ID, &item.EventType, &item.Payload, &item.Attempts); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (s *SQLStore) CompleteWebhook(ctx context.Context, id string, deliveryErr error, maxAttempts int) error {
	if deliveryErr == nil {
		_, err := s.db.ExecContext(ctx, `UPDATE logistics_.outbound_status_webhooks SET delivered_at=NOW(),attempts=attempts+1,last_error=NULL WHERE id=$1`, id)
		return err
	}
	var attempts int
	if err := s.db.QueryRowContext(ctx, `UPDATE logistics_.outbound_status_webhooks SET attempts=attempts+1,last_error=$2,next_attempt_at=NOW()+(INTERVAL '1 second'*LEAST(3600,POWER(2,attempts+1))) WHERE id=$1 RETURNING attempts`, id, deliveryErr.Error()).Scan(&attempts); err != nil {
		return err
	}
	if attempts >= maxAttempts {
		_, err := s.db.ExecContext(ctx, `UPDATE logistics_.outbound_status_webhooks SET dead_lettered_at=NOW() WHERE id=$1`, id)
		return err
	}
	return nil
}

func DSN(host, port, user, password, name, sslmode string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, name, sslmode)
}
