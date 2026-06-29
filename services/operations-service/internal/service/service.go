package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/adedaryorh/logistics-platform/pkg/clients/healthgrpc"
	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	"github.com/adedaryorh/logistics-platform/pkg/observability"
	"github.com/adedaryorh/logistics-platform/services/operations-service/internal/email"
	notificationtemplate "github.com/adedaryorh/logistics-platform/services/operations-service/internal/template"
)

type FeatureFlag struct {
	Name        string    `json:"name"`
	Enabled     bool      `json:"enabled"`
	RolloutPct  int       `json:"rollout_pct"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BackgroundJob struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Payload     map[string]any `json:"payload"`
	Status      string         `json:"status"`
	Attempts    int            `json:"attempts"`
	LastError   string         `json:"last_error,omitempty"`
	RunAt       *time.Time     `json:"run_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type Notification struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	Type              string    `json:"type"`
	TemplateName      string    `json:"template_name"`
	Recipient         string    `json:"recipient"`
	Status            string    `json:"status"`
	Provider          string    `json:"provider"`
	ProviderMessageID string    `json:"provider_message_id,omitempty"`
	SentAt            time.Time `json:"sent_at"`
}

type Incident struct {
	ID          string     `json:"id"`
	Category    string     `json:"category"`
	Severity    string     `json:"severity"`
	Status      string     `json:"status"`
	OrderID     string     `json:"order_id,omitempty"`
	Summary     string     `json:"summary"`
	Runbook     string     `json:"runbook"`
	CreatedAt   time.Time  `json:"created_at"`
	AckedAt     *time.Time `json:"acked_at,omitempty"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	LastEventAt time.Time  `json:"last_event_at"`
}

type WorkflowRun struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	ReferenceID string         `json:"reference_id"`
	Status      string         `json:"status"`
	Attempts    int            `json:"attempts"`
	LastError   string         `json:"last_error,omitempty"`
	Context     map[string]any `json:"context,omitempty"`
	StartedAt   time.Time      `json:"started_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type OrderOpsState struct {
	OrderID         string         `json:"order_id"`
	CustomerID      string         `json:"customer_id,omitempty"`
	DriverID        string         `json:"driver_id,omitempty"`
	PaymentStatus   string         `json:"payment_status"`
	DispatchStatus  string         `json:"dispatch_status"`
	TrackingStatus  string         `json:"tracking_status"`
	CurrentIncident string         `json:"current_incident,omitempty"`
	LastEvent       string         `json:"last_event"`
	LastEventAt     time.Time      `json:"last_event_at"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	Events          []string       `json:"events"`
}

type Service struct {
	cfg           *platformconfig.Config
	emailProvider email.Provider
	mu            sync.RWMutex
	featureFlags  map[string]*FeatureFlag
	jobs          map[string]*BackgroundJob
	notifications []Notification
	incidents     map[string]*Incident
	workflows     map[string]*WorkflowRun
	orderOps      map[string]*OrderOpsState
}

func New(cfg *platformconfig.Config) *Service {
	var provider email.Provider = email.MailtrapProvider{}
	if cfg != nil && cfg.Server.Environment == "production" {
		provider = email.ResendProvider{}
	}
	now := time.Now().UTC()
	return &Service{
		cfg:           cfg,
		emailProvider: provider,
		featureFlags: map[string]*FeatureFlag{
			"surge_pricing":    {Name: "surge_pricing", Enabled: true, RolloutPct: 100, Description: "Enable surge multiplier", UpdatedAt: now},
			"magic_link_login": {Name: "magic_link_login", Enabled: true, RolloutPct: 100, Description: "Allow passwordless auth", UpdatedAt: now},
		},
		jobs: map[string]*BackgroundJob{
			"job-1": {ID: "job-1", Type: "send_email", Payload: map[string]any{"template": "welcome"}, Status: "pending", Attempts: 0, CreatedAt: now},
		},
		notifications: []Notification{},
		incidents:     map[string]*Incident{},
		workflows:     map[string]*WorkflowRun{},
		orderOps:      map[string]*OrderOpsState{},
	}
}

func (s *Service) Health(ctx context.Context) map[string]string {
	if s.cfg == nil {
		return map[string]string{
			"identity-service":   "ok",
			"logistics-service":  "ok",
			"mobility-service":   "ok",
			"payment-service":    "ok",
			"operations-service": "ok",
		}
	}
	return map[string]string{
		"identity-service":   s.grpcHealthStatus(ctx, s.cfg.GRPC.IdentityServiceAddr),
		"logistics-service":  s.grpcHealthStatus(ctx, s.cfg.GRPC.LogisticsServiceAddr),
		"mobility-service":   s.grpcHealthStatus(ctx, s.cfg.GRPC.MobilityServiceAddr),
		"payment-service":    s.grpcHealthStatus(ctx, s.cfg.GRPC.PaymentServiceAddr),
		"operations-service": s.grpcHealthStatus(ctx, s.cfg.GRPC.OperationsServiceAddr),
	}
}

func (s *Service) ListFeatureFlags(ctx context.Context) []FeatureFlag {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]FeatureFlag, 0, len(s.featureFlags))
	for _, flag := range s.featureFlags {
		result = append(result, *flag)
	}
	return result
}

func (s *Service) ToggleFeatureFlag(ctx context.Context, name string) (*FeatureFlag, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	flag, ok := s.featureFlags[name]
	if !ok {
		return nil, platformerrors.ErrNotFound
	}
	flag.Enabled = !flag.Enabled
	flag.UpdatedAt = time.Now().UTC()
	return flag, nil
}

func (s *Service) ListJobs(ctx context.Context) []BackgroundJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]BackgroundJob, 0, len(s.jobs))
	for _, job := range s.jobs {
		result = append(result, *job)
	}
	return result
}

func (s *Service) RetryJob(ctx context.Context, id string) (*BackgroundJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	if !ok {
		return nil, platformerrors.ErrNotFound
	}
	job.Status = "pending"
	job.Attempts++
	job.LastError = ""
	return job, nil
}

func (s *Service) ListNotifications(ctx context.Context) []Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Notification(nil), s.notifications...)
}

func (s *Service) SendTemplateEmail(ctx context.Context, userID, recipient, templateName string, data any) (*Notification, error) {
	subject, htmlBody, textBody, err := notificationtemplate.Render(templateName, data)
	if err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}
	response, err := s.emailProvider.Send(ctx, email.Request{
		To:        recipient,
		Subject:   subject,
		BodyHTML:  htmlBody,
		BodyText:  textBody,
		FromName:  "Logistics Platform",
		FromEmail: "no-reply@example.com",
	})
	if err != nil {
		return nil, fmt.Errorf("send email: %w", err)
	}
	notification := Notification{
		ID:                fmt.Sprintf("notif-%d", len(s.notifications)+1),
		UserID:            userID,
		Type:              "email",
		TemplateName:      templateName,
		Recipient:         recipient,
		Status:            response.Status,
		Provider:          s.emailProvider.Name(),
		ProviderMessageID: response.ProviderMessageID,
		SentAt:            time.Now().UTC(),
	}
	s.mu.Lock()
	s.notifications = append(s.notifications, notification)
	s.mu.Unlock()
	observability.IncCounter("notifications_sent_total", 1, map[string]string{"provider": notification.Provider, "template": templateName, "status": notification.Status})
	return &notification, nil
}

func (s *Service) RecordNotification(userID, recipient, templateName, status string) Notification {
	s.mu.Lock()
	defer s.mu.Unlock()
	notification := Notification{
		ID:           fmt.Sprintf("notif-%d", len(s.notifications)+1),
		UserID:       userID,
		Type:         "event",
		TemplateName: templateName,
		Recipient:    recipient,
		Status:       status,
		Provider:     "kafka",
		SentAt:       time.Now().UTC(),
	}
	s.notifications = append(s.notifications, notification)
	observability.IncCounter("notifications_recorded_total", 1, map[string]string{"template": templateName, "status": status})
	return notification
}

func (s *Service) RecordBackgroundJob(jobType string, payload map[string]any) *BackgroundJob {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := fmt.Sprintf("job-%d", len(s.jobs)+1)
	job := &BackgroundJob{
		ID:        id,
		Type:      jobType,
		Payload:   payload,
		Status:    "pending",
		Attempts:  0,
		CreatedAt: time.Now().UTC(),
	}
	s.jobs[id] = job
	observability.IncCounter("background_jobs_created_total", 1, map[string]string{"type": jobType})
	observability.SetGauge("background_jobs_total", float64(len(s.jobs)), nil)
	return job
}

func (s *Service) ListIncidents(ctx context.Context) []Incident {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Incident, 0, len(s.incidents))
	for _, incident := range s.incidents {
		result = append(result, *incident)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result
}

func (s *Service) AcknowledgeIncident(ctx context.Context, id string) (*Incident, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	incident, ok := s.incidents[id]
	if !ok {
		return nil, platformerrors.ErrNotFound
	}
	now := time.Now().UTC()
	incident.Status = "acknowledged"
	incident.AckedAt = &now
	incident.LastEventAt = now
	copyIncident := *incident
	return &copyIncident, nil
}

func (s *Service) ResolveIncident(ctx context.Context, id string) (*Incident, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	incident, ok := s.incidents[id]
	if !ok {
		return nil, platformerrors.ErrNotFound
	}
	now := time.Now().UTC()
	incident.Status = "resolved"
	incident.ResolvedAt = &now
	incident.LastEventAt = now
	copyIncident := *incident
	return &copyIncident, nil
}

func (s *Service) ListWorkflows(ctx context.Context) []WorkflowRun {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]WorkflowRun, 0, len(s.workflows))
	for _, workflow := range s.workflows {
		result = append(result, *workflow)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})
	return result
}

func (s *Service) RetryWorkflow(ctx context.Context, id string) (*WorkflowRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	workflow, ok := s.workflows[id]
	if !ok {
		return nil, platformerrors.ErrNotFound
	}
	workflow.Attempts++
	workflow.Status = "retrying"
	workflow.LastError = ""
	workflow.UpdatedAt = time.Now().UTC()
	copyWorkflow := *workflow
	return &copyWorkflow, nil
}

func (s *Service) GetOrderTrace(ctx context.Context, orderID string) (*OrderOpsState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	trace, ok := s.orderOps[orderID]
	if !ok {
		return nil, platformerrors.ErrNotFound
	}
	copyTrace := *trace
	copyTrace.Events = append([]string(nil), trace.Events...)
	return &copyTrace, nil
}

func (s *Service) recordWorkflowLocked(runType, referenceID, status string, context map[string]any) *WorkflowRun {
	id := fmt.Sprintf("wf-%d", len(s.workflows)+1)
	now := time.Now().UTC()
	run := &WorkflowRun{
		ID:          id,
		Type:        runType,
		ReferenceID: referenceID,
		Status:      status,
		Attempts:    1,
		Context:     context,
		StartedAt:   now,
		UpdatedAt:   now,
	}
	s.workflows[id] = run
	return run
}

func (s *Service) getOrCreateOrderOpsLocked(orderID string) *OrderOpsState {
	state, ok := s.orderOps[orderID]
	if !ok {
		state = &OrderOpsState{
			OrderID:        orderID,
			PaymentStatus:  "pending",
			DispatchStatus: "pending",
			TrackingStatus: "inactive",
			Metadata:       map[string]any{},
			Events:         []string{},
		}
		s.orderOps[orderID] = state
	}
	return state
}

func (s *Service) appendOrderEventLocked(orderID, event string, metadata map[string]any) *OrderOpsState {
	state := s.getOrCreateOrderOpsLocked(orderID)
	state.LastEvent = event
	state.LastEventAt = time.Now().UTC()
	state.Events = append(state.Events, fmt.Sprintf("%s:%s", state.LastEventAt.Format(time.RFC3339), event))
	if metadata != nil {
		for key, value := range metadata {
			state.Metadata[key] = value
		}
	}
	return state
}

func (s *Service) createIncidentLocked(category, severity, orderID, summary, runbook string) *Incident {
	id := fmt.Sprintf("inc-%d", len(s.incidents)+1)
	now := time.Now().UTC()
	incident := &Incident{
		ID:          id,
		Category:    category,
		Severity:    severity,
		Status:      "open",
		OrderID:     orderID,
		Summary:     summary,
		Runbook:     runbook,
		CreatedAt:   now,
		LastEventAt: now,
	}
	s.incidents[id] = incident
	return incident
}

func (s *Service) grpcHealthStatus(ctx context.Context, target string) string {
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
