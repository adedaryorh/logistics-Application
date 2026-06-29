package tracking

import "sync"

type DriverLocationMsg struct {
	DriverID   string  `json:"driver_id"`
	OrderID    string  `json:"order_id"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	SpeedKMH   float64 `json:"speed_kmh"`
	HeadingDeg float64 `json:"heading_deg"`
}

type Hub struct {
	sessions map[string]map[chan DriverLocationMsg]struct{}
	mu       sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		sessions: map[string]map[chan DriverLocationMsg]struct{}{},
	}
}

func (h *Hub) BroadcastDriverLocation(orderID string, loc DriverLocationMsg) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for subscriber := range h.sessions[orderID] {
		select {
		case subscriber <- loc:
		default:
		}
	}
}

func (h *Hub) Subscribe(orderID string, ch chan DriverLocationMsg) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sessions[orderID] == nil {
		h.sessions[orderID] = map[chan DriverLocationMsg]struct{}{}
	}
	h.sessions[orderID][ch] = struct{}{}
}

func (h *Hub) Unsubscribe(orderID string, ch chan DriverLocationMsg) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.sessions[orderID], ch)
}
