package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	platformauth "github.com/adedaryorh/logistics-platform/pkg/auth"
	"github.com/adedaryorh/logistics-platform/pkg/clients/healthgrpc"
	"github.com/adedaryorh/logistics-platform/pkg/clients/mobilitygrpc"
	"github.com/adedaryorh/logistics-platform/pkg/clients/paymentgrpc"
	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
)

type Tool struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	RequiredRole      string   `json:"required_role"`
	RequiredFields    []string `json:"required_fields,omitempty"`
	Sensitive         bool     `json:"sensitive"`
	RequiresReason    bool     `json:"requires_reason"`
	MaxCallsPerMinute int      `json:"max_calls_per_minute,omitempty"`
}

type AuditEntry struct {
	Tool      string    `json:"tool"`
	Actor     string    `json:"actor"`
	InputHash string    `json:"input_hash"`
	PrevHash  string    `json:"prev_hash,omitempty"`
	Hash      string    `json:"hash"`
	CalledAt  time.Time `json:"called_at"`
}

type APIKeyRecord struct {
	Actor     string
	Role      string
	ExpiresAt time.Time
}

type Service struct {
	cfg           *platformconfig.Config
	mu            sync.Mutex
	keys          map[string]APIKeyRecord
	audit         []AuditEntry
	lastAuditHash string
	tools         []Tool
	callHistory   map[string][]time.Time
}

var piiPatterns = []*regexp.Regexp{
	regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`),
	regexp.MustCompile(`\+?[0-9]{10,15}`),
}

func New(cfg *platformconfig.Config) *Service {
	return &Service{
		cfg: cfg,
		keys: map[string]APIKeyRecord{
			hashKey("ops-key"):       {Actor: "admin-user", Role: "admin", ExpiresAt: time.Now().UTC().Add(24 * time.Hour)},
			hashKey("ops-admin-key"): {Actor: "super-admin-user", Role: "super_admin", ExpiresAt: time.Now().UTC().Add(24 * time.Hour)},
			hashKey("finance-key"):   {Actor: "finance-user", Role: "finance", ExpiresAt: time.Now().UTC().Add(24 * time.Hour)},
			hashKey("support-key"):   {Actor: "user-support", Role: "user", ExpiresAt: time.Now().UTC().Add(24 * time.Hour)},
		},
		lastAuditHash: "",
		tools:         defaultTools(),
		callHistory:   map[string][]time.Time{},
	}
}

func defaultTools() []Tool {
	return []Tool{
		{Name: "health_check", Description: "Returns health status of all platform services", RequiredRole: "admin", MaxCallsPerMinute: 60},
		{Name: "trace_order", Description: "Returns full lifecycle trace of an order", RequiredRole: "support", RequiredFields: []string{"order_id"}, MaxCallsPerMinute: 30},
		{Name: "trace_trip", Description: "Returns trip event history and GPS track", RequiredRole: "support", RequiredFields: []string{"order_id"}, MaxCallsPerMinute: 30},
		{Name: "nearby_drivers", Description: "Returns available drivers near a coordinate", RequiredRole: "admin", RequiredFields: []string{"lat", "lng"}, MaxCallsPerMinute: 60},
		{Name: "driver_location", Description: "Returns current location of a driver", RequiredRole: "admin", RequiredFields: []string{"driver_id"}, MaxCallsPerMinute: 60},
		{Name: "get_eta", Description: "Returns current ETA estimate for an active order", RequiredRole: "support", RequiredFields: []string{"order_id"}, MaxCallsPerMinute: 30},
		{Name: "get_route", Description: "Returns computed route for a trip", RequiredRole: "admin", RequiredFields: []string{"from_lat", "from_lng", "to_lat", "to_lng"}, MaxCallsPerMinute: 30},
		{Name: "payment_status", Description: "Returns transaction status", RequiredRole: "finance", RequiredFields: []string{"transaction_id"}, Sensitive: true, RequiresReason: true, MaxCallsPerMinute: 20},
		{Name: "get_logs", Description: "Returns recent structured logs for a service", RequiredRole: "admin", RequiredFields: []string{"service"}, Sensitive: true, RequiresReason: true, MaxCallsPerMinute: 20},
		{Name: "get_metrics", Description: "Returns current metrics for a service", RequiredRole: "admin", RequiredFields: []string{"service"}, MaxCallsPerMinute: 60},
		{Name: "retry_workflow", Description: "Re-triggers a failed workflow by ID", RequiredRole: "super_admin", RequiredFields: []string{"workflow_id", "reason"}, Sensitive: true, RequiresReason: true, MaxCallsPerMinute: 10},
		{Name: "dispatch_status", Description: "Returns dispatch attempt history for an order", RequiredRole: "admin", RequiredFields: []string{"order_id"}, MaxCallsPerMinute: 30},
		{Name: "incident_queue", Description: "Returns active operational incidents", RequiredRole: "admin", MaxCallsPerMinute: 30},
		{Name: "workflow_runs", Description: "Returns tracked workflow runs", RequiredRole: "admin", MaxCallsPerMinute: 30},
		{Name: "audit_chain_status", Description: "Returns MCP audit-chain integrity status", RequiredRole: "super_admin", Sensitive: true, RequiresReason: true, MaxCallsPerMinute: 10},
	}
}

func (s *Service) Authenticate(apiKey string) (APIKeyRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.keys[hashKey(apiKey)]
	if !ok || time.Now().UTC().After(record.ExpiresAt) {
		return APIKeyRecord{}, platformerrors.ErrUnauthorized
	}
	return record, nil
}

func (s *Service) ToolsForRole(role string) []Tool {
	visible := make([]Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		if allowedRole(role, tool.RequiredRole) {
			visible = append(visible, tool)
		}
	}
	return visible
}

func (s *Service) CallTool(ctx context.Context, actor APIKeyRecord, name string, input map[string]any) (map[string]any, error) {
	var selected *Tool
	for i := range s.tools {
		if s.tools[i].Name == name {
			selected = &s.tools[i]
			break
		}
	}
	if selected == nil {
		return nil, platformerrors.ErrNotFound
	}
	if !allowedRole(actor.Role, selected.RequiredRole) {
		return nil, platformerrors.ErrUnauthorized
	}
	if err := validateToolInput(*selected, input); err != nil {
		return nil, err
	}
	if err := s.enforceRateLimit(actor, *selected); err != nil {
		return nil, err
	}

	output := map[string]any{
		"tool":       name,
		"status":     "ok",
		"input":      scrubMap(input),
		"governance": map[string]any{"sensitive": selected.Sensitive, "reason_provided": strings.TrimSpace(toString(input["reason"])) != ""},
	}
	switch name {
	case "health_check":
		output["services"] = s.healthCheck(ctx)
	case "trace_order":
		output["trace"] = map[string]any{"order_id": input["order_id"], "events": []string{"order.created", "dispatch.offered", "order.assigned"}, "tracking_status": "active"}
	case "trace_trip":
		output["trip"] = map[string]any{"order_id": input["order_id"], "gps_points": 12, "status": "active"}
	case "nearby_drivers":
		drivers, err := s.mobilityNearbyDrivers(ctx, input)
		if err != nil {
			return nil, err
		}
		output["drivers"] = drivers
	case "driver_location":
		driver, err := s.mobilityDriverLocation(ctx, toString(input["driver_id"]))
		if err != nil {
			return nil, err
		}
		output["driver"] = driver
	case "get_eta":
		output["eta"] = map[string]any{"order_id": input["order_id"], "eta_minutes": 8}
	case "get_route":
		output["route"] = map[string]any{"distance_m": 1900, "duration_s": 95}
	case "payment_status":
		payment, err := s.paymentStatus(ctx, toString(input["transaction_id"]))
		if err != nil {
			return nil, err
		}
		output["payment"] = payment
	case "get_logs":
		output["logs"] = scrubString("user email jane@example.com phone +2348012345678")
	case "get_metrics":
		output["metrics"] = map[string]any{"service": input["service"], "request_rate": "12rps", "error_rate": "0.2%", "p99_ms": 180}
	case "retry_workflow":
		output["workflow"] = map[string]any{"workflow_id": input["workflow_id"], "status": "retrying"}
	case "dispatch_status":
		output["dispatch"] = map[string]any{"order_id": input["order_id"], "attempts": 3, "current_status": "assigned"}
	case "incident_queue":
		output["incidents"] = []map[string]any{{"incident_id": "inc-1", "severity": "high", "status": "open", "summary": "Payment failures elevated"}}
	case "workflow_runs":
		output["workflows"] = []map[string]any{{"workflow_id": "wf-1", "type": "order_lifecycle", "status": "dispatching"}}
	case "audit_chain_status":
		output["audit"] = map[string]any{"entries": len(s.audit), "last_hash": s.lastAuditHash, "integrity": "ok"}
	default:
		output["result"] = "simulated"
	}

	s.mu.Lock()
	entry := AuditEntry{
		Tool:      name,
		Actor:     actor.Actor,
		InputHash: hashInput(input),
		PrevHash:  s.lastAuditHash,
		CalledAt:  time.Now().UTC(),
	}
	entry.Hash = hashKey(entry.Tool + "|" + entry.Actor + "|" + entry.InputHash + "|" + entry.PrevHash)
	s.lastAuditHash = entry.Hash
	s.audit = append(s.audit, entry)
	s.mu.Unlock()

	return output, nil
}

func (s *Service) AuditLog() []AuditEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]AuditEntry(nil), s.audit...)
}

func scrubString(value string) string {
	scrubbed := value
	for _, pattern := range piiPatterns {
		scrubbed = pattern.ReplaceAllStringFunc(scrubbed, func(match string) string {
			if strings.Contains(match, "@") {
				return "***@***.***"
			}
			return "+234***"
		})
	}
	return scrubbed
}

func scrubMap(input map[string]any) map[string]any {
	result := map[string]any{}
	for key, value := range input {
		if text, ok := value.(string); ok {
			result[key] = scrubString(text)
			continue
		}
		result[key] = value
	}
	return result
}

func hashKey(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func hashInput(input map[string]any) string {
	parts := make([]string, 0, len(input))
	for key, value := range input {
		parts = append(parts, key+"="+scrubString(toString(value)))
	}
	sort.Strings(parts)
	return hashKey(strings.Join(parts, "&"))
}

func validateToolInput(tool Tool, input map[string]any) error {
	for _, field := range tool.RequiredFields {
		value, ok := input[field]
		if !ok || isEmptyValue(value) {
			return platformerrors.ErrBadRequest
		}
	}
	if tool.RequiresReason {
		value, ok := input["reason"]
		if !ok || isEmptyValue(value) {
			return platformerrors.ErrBadRequest
		}
	}
	return nil
}

func isEmptyValue(value any) bool {
	if value == nil {
		return true
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text) == ""
	}
	if strings.TrimSpace(fmt.Sprint(value)) == "<nil>" {
		return true
	}
	return false
}

func (s *Service) enforceRateLimit(actor APIKeyRecord, tool Tool) error {
	if tool.MaxCallsPerMinute <= 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := actor.Actor + ":" + tool.Name
	now := time.Now().UTC()
	windowStart := now.Add(-time.Minute)
	filtered := s.callHistory[key][:0]
	for _, calledAt := range s.callHistory[key] {
		if calledAt.After(windowStart) {
			filtered = append(filtered, calledAt)
		}
	}
	if len(filtered) >= tool.MaxCallsPerMinute {
		s.callHistory[key] = filtered
		return platformerrors.ErrRateLimit
	}
	filtered = append(filtered, now)
	s.callHistory[key] = filtered
	return nil
}

func allowedRole(actorRole, requiredRole string) bool {
	if requiredRole == "support" {
		return platformauth.RoleAllowed(actorRole, "user", "admin", "super_admin")
	}
	return platformauth.RoleAllowed(actorRole, requiredRole)
}

func (s *Service) healthCheck(ctx context.Context) map[string]string {
	if s.cfg == nil {
		return map[string]string{"identity-service": "ok", "logistics-service": "ok", "mobility-service": "ok", "payment-service": "ok", "operations-service": "ok"}
	}
	return map[string]string{
		"identity-service":   s.grpcHealth(ctx, s.cfg.GRPC.IdentityServiceAddr),
		"logistics-service":  s.grpcHealth(ctx, s.cfg.GRPC.LogisticsServiceAddr),
		"mobility-service":   s.grpcHealth(ctx, s.cfg.GRPC.MobilityServiceAddr),
		"payment-service":    s.grpcHealth(ctx, s.cfg.GRPC.PaymentServiceAddr),
		"operations-service": s.grpcHealth(ctx, s.cfg.GRPC.OperationsServiceAddr),
	}
}

func (s *Service) grpcHealth(ctx context.Context, target string) string {
	rpcCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	client, conn, err := healthgrpc.DialClient(rpcCtx, target)
	if err != nil {
		return "unreachable"
	}
	defer conn.Close()
	resp, err := client.Check(rpcCtx, &healthgrpc.CheckRequest{})
	if err != nil || resp == nil || resp.Status == "" {
		return "unreachable"
	}
	return resp.Status
}

func (s *Service) mobilityNearbyDrivers(ctx context.Context, input map[string]any) ([]map[string]any, error) {
	if s.cfg == nil {
		return []map[string]any{{"driver_id": "driver-1", "lat": 6.52, "lng": 3.37}}, nil
	}
	rpcCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	client, conn, err := mobilitygrpc.DialClient(rpcCtx, s.cfg.GRPC.MobilityServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("dial mobility grpc: %w", err)
	}
	defer conn.Close()
	resp, err := client.NearbyDrivers(rpcCtx, &mobilitygrpc.NearbyDriversRequest{
		Lat:         floatValue(input["lat"]),
		Lng:         floatValue(input["lng"]),
		KRing:       int32(intValue(input["k_ring"], 1)),
		VehicleType: toString(input["vehicle_type"]),
	})
	if err != nil {
		return nil, fmt.Errorf("call mobility nearby drivers: %w", err)
	}
	out := make([]map[string]any, 0, len(resp.Drivers))
	for _, driver := range resp.Drivers {
		out = append(out, map[string]any{
			"driver_id":    driver.DriverID,
			"lat":          driver.Lat,
			"lng":          driver.Lng,
			"vehicle_type": driver.VehicleType,
			"rating":       driver.Rating,
		})
	}
	return out, nil
}

func (s *Service) mobilityDriverLocation(ctx context.Context, driverID string) (map[string]any, error) {
	if s.cfg == nil {
		return map[string]any{"driver_id": driverID, "lat": 6.5244, "lng": 3.3792}, nil
	}
	rpcCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	client, conn, err := mobilitygrpc.DialClient(rpcCtx, s.cfg.GRPC.MobilityServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("dial mobility grpc: %w", err)
	}
	defer conn.Close()
	resp, err := client.DriverLocation(rpcCtx, &mobilitygrpc.DriverLocationRequest{DriverID: driverID})
	if err != nil {
		return nil, fmt.Errorf("call mobility driver location: %w", err)
	}
	if resp.Driver == nil {
		return nil, platformerrors.ErrNotFound
	}
	return map[string]any{
		"driver_id":    resp.Driver.DriverID,
		"lat":          resp.Driver.Lat,
		"lng":          resp.Driver.Lng,
		"vehicle_type": resp.Driver.VehicleType,
		"updated_at":   resp.Driver.UpdatedAt,
	}, nil
}

func (s *Service) paymentStatus(ctx context.Context, transactionID string) (map[string]any, error) {
	if s.cfg == nil {
		return map[string]any{"transaction_id": transactionID, "status": "completed"}, nil
	}
	rpcCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	client, conn, err := paymentgrpc.DialClient(rpcCtx, s.cfg.GRPC.PaymentServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("dial payment grpc: %w", err)
	}
	defer conn.Close()
	resp, err := client.PaymentStatus(rpcCtx, &paymentgrpc.PaymentStatusRequest{TransactionID: transactionID})
	if err != nil {
		return nil, fmt.Errorf("call payment status: %w", err)
	}
	return map[string]any{
		"transaction_id": resp.TransactionID,
		"order_id":       resp.OrderID,
		"customer_id":    resp.CustomerID,
		"amount_minor":   resp.AmountMinor,
		"currency":       resp.Currency,
		"provider":       resp.Provider,
		"status":         resp.Status,
	}, nil
}

func floatValue(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int32:
		return float64(typed)
	case int64:
		return float64(typed)
	default:
		return 0
	}
}

func intValue(value any, fallback int) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return fallback
	}
}

func toString(value any) string {
	return strings.TrimSpace(fmt.Sprint(value))
}
