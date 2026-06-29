package model

import "time"

type OrderType string

const (
	OrderTypeRide   OrderType = "ride"
	OrderTypeFood   OrderType = "food"
	OrderTypeParcel OrderType = "parcel"
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
	ID                 string      `json:"id"`
	CustomerID         string      `json:"customer_id"`
	DriverID           *string     `json:"driver_id,omitempty"`
	MerchantID         *string     `json:"merchant_id,omitempty"`
	Type               OrderType   `json:"type"`
	Status             OrderStatus `json:"status"`
	Pickup             Coordinate  `json:"pickup"`
	Dropoff            Coordinate  `json:"dropoff"`
	Items              []OrderItem `json:"items,omitempty"`
	PriceMinor         int64       `json:"price_minor"`
	Currency           string      `json:"currency"`
	SurgeMultiplier    float64     `json:"surge_multiplier"`
	IdempotencyKey     string      `json:"idempotency_key"`
	TemporalWorkflowID string      `json:"temporal_workflow_id,omitempty"`
	CancellationReason *string     `json:"cancellation_reason,omitempty"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
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
