package http

import (
	"context"
	"time"

	"github.com/adedaryorh/logistics-platform/pkg/clients/healthgrpc"
	"github.com/adedaryorh/logistics-platform/pkg/clients/paymentgrpc"
	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	paymentservice "github.com/adedaryorh/logistics-platform/services/payment-service/internal/service"
	"google.golang.org/grpc"
)

func (a *App) RegisterGRPC(server *grpc.Server, cfg *platformconfig.Config) {
	healthgrpc.RegisterServer(server, &paymentHealthServer{service: "payment-service", version: cfg.Server.Version})
	paymentgrpc.RegisterServer(server, &paymentGRPCServer{service: a.handler.service})
}

type paymentHealthServer struct {
	service string
	version string
}

func (s *paymentHealthServer) Check(ctx context.Context, _ *healthgrpc.CheckRequest) (*healthgrpc.CheckResponse, error) {
	return &healthgrpc.CheckResponse{
		Service:   s.service,
		Status:    "ok",
		Version:   s.version,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

type paymentGRPCServer struct {
	service *paymentservice.Service
}

func (s *paymentGRPCServer) PaymentStatus(ctx context.Context, req *paymentgrpc.PaymentStatusRequest) (*paymentgrpc.PaymentStatusResponse, error) {
	tx, err := s.service.GetPayment(ctx, req.TransactionID)
	if err != nil {
		return nil, err
	}
	return &paymentgrpc.PaymentStatusResponse{
		TransactionID: tx.ID,
		OrderID:       tx.OrderID,
		CustomerID:    tx.CustomerID,
		AmountMinor:   tx.AmountMinor,
		Currency:      tx.Currency,
		Provider:      tx.Provider,
		Status:        tx.Status,
	}, nil
}
