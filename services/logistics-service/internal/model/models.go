package model

import "time"

type OrderType string

const (
	OrderTypeRide         OrderType = "ride"
	OrderTypeFood         OrderType = "food"
	OrderTypeParcel       OrderType = "parcel"
	OrderTypeAgricultural OrderType = "agricultural"
)

type OrderStatus string

const (
	OrderStatusPending         OrderStatus = "pending"
	OrderStatusAwaitingPayment OrderStatus = "awaiting_payment"
	OrderStatusPaid            OrderStatus = "paid"
	OrderStatusDispatching     OrderStatus = "dispatching"
	OrderStatusAssigned        OrderStatus = "assigned"
	OrderStatusPickedUp        OrderStatus = "picked_up"
	OrderStatusDelivered       OrderStatus = "delivered"
	OrderStatusCancelled       OrderStatus = "cancelled"
	OrderStatusFailed          OrderStatus = "failed"
)

type AssignmentStatus string

const (
	AssignmentStatusOffered  AssignmentStatus = "offered"
	AssignmentStatusAccepted AssignmentStatus = "accepted"
	AssignmentStatusRejected AssignmentStatus = "rejected"
	AssignmentStatusTimedOut AssignmentStatus = "timed_out"
)

type Coordinate struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address,omitempty"`
	H3Cell  string  `json:"h3_cell,omitempty"`
}

type OrderItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	PriceMinor int64  `json:"price_minor"`
	Currency   string `json:"currency"`
}

type Order struct {
	ID                    string                `json:"id"`
	CustomerID            string                `json:"customer_id"`
	DriverID              *string               `json:"driver_id,omitempty"`
	MerchantID            *string               `json:"merchant_id,omitempty"`
	Type                  OrderType             `json:"type"`
	Status                OrderStatus           `json:"status"`
	Pickup                Coordinate            `json:"pickup"`
	Dropoff               Coordinate            `json:"dropoff"`
	Items                 []OrderItem           `json:"items,omitempty"`
	PriceMinor            int64                 `json:"price_minor"`
	Currency              string                `json:"currency"`
	SurgeMultiplier       float64               `json:"surge_multiplier"`
	IdempotencyKey        string                `json:"idempotency_key"`
	TemporalWorkflowID    string                `json:"temporal_workflow_id,omitempty"`
	CancellationReason    *string               `json:"cancellation_reason,omitempty"`
	CreatedAt             time.Time             `json:"created_at"`
	UpdatedAt             time.Time             `json:"updated_at"`
	PlatformUserID        string                `json:"platform_user_id,omitempty"`
	SourcePlatformService string                `json:"source_platform_service,omitempty"`
	MarketplaceRequestID  string                `json:"marketplace_request_id,omitempty"`
	FarmSenseRequestID    string                `json:"farmsense_request_id,omitempty"`
	AgriculturalShipment  *AgriculturalShipment `json:"agricultural_shipment,omitempty"`
	Proofs                []DeliveryProof       `json:"proofs,omitempty"`
}

type TimeWindow struct {
	StartAt time.Time `json:"start_at"`
	EndAt   time.Time `json:"end_at"`
}
type AgriculturalShipment struct {
	ProduceType           string     `json:"produce_type"`
	Quantity              float64    `json:"quantity"`
	QuantityUnit          string     `json:"quantity_unit"`
	WeightKG              *float64   `json:"weight_kg,omitempty"`
	Packaging             string     `json:"packaging"`
	RequiresRefrigeration bool       `json:"requires_refrigeration"`
	ColdChainMinC         *float64   `json:"cold_chain_min_c,omitempty"`
	ColdChainMaxC         *float64   `json:"cold_chain_max_c,omitempty"`
	PickupWindow          TimeWindow `json:"pickup_window"`
	DeliveryWindow        TimeWindow `json:"delivery_window"`
	HandlingNotes         string     `json:"handling_notes,omitempty"`
	LoadingNotes          string     `json:"loading_notes,omitempty"`
}
type AgriculturalQuote struct {
	ID                   string               `json:"id"`
	PlatformService      string               `json:"platform_service"`
	PlatformUserID       string               `json:"platform_user_id"`
	MarketplaceRequestID string               `json:"marketplace_request_id"`
	FarmSenseRequestID   string               `json:"farmsense_request_id"`
	Pickup               Coordinate           `json:"pickup"`
	Dropoff              Coordinate           `json:"dropoff"`
	Shipment             AgriculturalShipment `json:"shipment"`
	PriceMinor           int64                `json:"price_minor"`
	Currency             string               `json:"currency"`
	ExpiresAt            time.Time            `json:"expires_at"`
	CreatedAt            time.Time            `json:"created_at"`
}
type DeliveryProof struct {
	ID            string     `json:"id"`
	OrderID       string     `json:"order_id"`
	Type          string     `json:"type"`
	EvidenceURL   string     `json:"evidence_url"`
	Notes         string     `json:"notes,omitempty"`
	RecipientName string     `json:"recipient_name,omitempty"`
	Coordinate    Coordinate `json:"coordinate"`
	CapturedAt    time.Time  `json:"captured_at"`
	DriverID      string     `json:"driver_id"`
}

type OrderStatusHistory struct {
	ID         string       `json:"id"`
	OrderID    string       `json:"order_id"`
	FromStatus *OrderStatus `json:"from_status,omitempty"`
	ToStatus   OrderStatus  `json:"to_status"`
	Reason     *string      `json:"reason,omitempty"`
	ActorID    *string      `json:"actor_id,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
}

type Driver struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	FullName   string    `json:"full_name"`
	Phone      string    `json:"phone"`
	Type       string    `json:"type"`
	Lat        float64   `json:"lat"`
	Lng        float64   `json:"lng"`
	H3Cell     string    `json:"h3_cell"`
	Rating     float64   `json:"rating"`
	TotalTrips int       `json:"total_trips"`
	Online     bool      `json:"online"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Merchant struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Address   string     `json:"address"`
	Lat       float64    `json:"lat"`
	Lng       float64    `json:"lng"`
	H3Cell    string     `json:"h3_cell"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	MenuItems []MenuItem `json:"menu_items,omitempty"`
}

type MenuItem struct {
	ID          string    `json:"id"`
	MerchantID  string    `json:"merchant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	PriceMinor  int64     `json:"price_minor"`
	Currency    string    `json:"currency"`
	IsAvailable bool      `json:"is_available"`
	CreatedAt   time.Time `json:"created_at"`
}

type Assignment struct {
	ID            string           `json:"id"`
	OrderID       string           `json:"order_id"`
	DriverID      string           `json:"driver_id"`
	AttemptNumber int              `json:"attempt_number"`
	Status        AssignmentStatus `json:"status"`
	OfferedAt     time.Time        `json:"offered_at"`
	RespondedAt   *time.Time       `json:"responded_at,omitempty"`
	TimeoutAt     time.Time        `json:"timeout_at"`
}

type Rating struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	RaterID   string    `json:"rater_id"`
	RateeID   string    `json:"ratee_id"`
	RateeType string    `json:"ratee_type"`
	Score     int       `json:"score"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type OutboxEvent struct {
	ID          string         `json:"id"`
	AggregateID string         `json:"aggregate_id"`
	EventType   string         `json:"event_type"`
	Payload     map[string]any `json:"payload"`
	CreatedAt   time.Time      `json:"created_at"`
}
