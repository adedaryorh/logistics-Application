package http

import (
	"context"
	"time"

	"github.com/adedaryorh/logistics-platform/pkg/clients/healthgrpc"
	"github.com/adedaryorh/logistics-platform/pkg/clients/mobilitygrpc"
	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	service "github.com/adedaryorh/logistics-platform/services/mobility-service/internal/service"
	"google.golang.org/grpc"
)

func (a *App) RegisterGRPC(server *grpc.Server, cfg *platformconfig.Config) {
	healthgrpc.RegisterServer(server, &healthServer{service: "mobility-service", version: cfg.Server.Version})
	mobilitygrpc.RegisterServer(server, &mobilityGRPCServer{service: a.handler.service})
}

type healthServer struct {
	service string
	version string
}

func (s *healthServer) Check(ctx context.Context, _ *healthgrpc.CheckRequest) (*healthgrpc.CheckResponse, error) {
	return &healthgrpc.CheckResponse{
		Service:   s.service,
		Status:    "ok",
		Version:   s.version,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

type mobilityGRPCServer struct {
	service *service.Service
}

func (s *mobilityGRPCServer) NearbyDrivers(ctx context.Context, req *mobilitygrpc.NearbyDriversRequest) (*mobilitygrpc.NearbyDriversResponse, error) {
	drivers, err := s.service.FindNearbyDrivers(ctx, req.Lat, req.Lng, int(req.KRing), req.VehicleType)
	if err != nil {
		return nil, err
	}
	resp := &mobilitygrpc.NearbyDriversResponse{Drivers: make([]*mobilitygrpc.DriverSnapshot, 0, len(drivers))}
	for _, driver := range drivers {
		resp.Drivers = append(resp.Drivers, &mobilitygrpc.DriverSnapshot{
			DriverID:    driver.DriverID,
			Lat:         driver.Lat,
			Lng:         driver.Lng,
			H3Cell:      driver.H3Cell,
			Rating:      driver.Rating,
			TotalTrips:  int32(driver.TotalTrips),
			VehicleType: driver.VehicleType,
			UpdatedAt:   driver.UpdatedAt.Format(time.RFC3339),
		})
	}
	return resp, nil
}

func (s *mobilityGRPCServer) DriverLocation(ctx context.Context, req *mobilitygrpc.DriverLocationRequest) (*mobilitygrpc.DriverLocationResponse, error) {
	driver, err := s.service.GetDriverLocation(ctx, req.DriverID)
	if err != nil {
		return nil, err
	}
	return &mobilitygrpc.DriverLocationResponse{
		Driver: &mobilitygrpc.DriverSnapshot{
			DriverID:    driver.DriverID,
			Lat:         driver.Lat,
			Lng:         driver.Lng,
			H3Cell:      driver.H3Cell,
			Rating:      driver.Rating,
			TotalTrips:  int32(driver.TotalTrips),
			VehicleType: driver.VehicleType,
			UpdatedAt:   driver.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}
