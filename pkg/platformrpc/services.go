package platformrpc

import (
	"context"

	"google.golang.org/grpc"
)

type HealthCheckRequest struct{}

type HealthCheckResponse struct {
	Service   string `json:"service"`
	Status    string `json:"status"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

type NearbyDriversRequest struct {
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	KRing       int32   `json:"k_ring"`
	VehicleType string  `json:"vehicle_type"`
}

type DriverSnapshot struct {
	DriverID    string  `json:"driver_id"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	H3Cell      string  `json:"h3_cell"`
	Rating      float64 `json:"rating"`
	TotalTrips  int32   `json:"total_trips"`
	VehicleType string  `json:"vehicle_type"`
	UpdatedAt   string  `json:"updated_at"`
}

type NearbyDriversResponse struct {
	Drivers []*DriverSnapshot `json:"drivers"`
}

type DriverLocationRequest struct {
	DriverID string `json:"driver_id"`
}

type DriverLocationResponse struct {
	Driver *DriverSnapshot `json:"driver"`
}

type PaymentStatusRequest struct {
	TransactionID string `json:"transaction_id"`
}

type PaymentStatusResponse struct {
	TransactionID string `json:"transaction_id"`
	OrderID       string `json:"order_id"`
	CustomerID    string `json:"customer_id"`
	AmountMinor   int64  `json:"amount_minor"`
	Currency      string `json:"currency"`
	Provider      string `json:"provider"`
	Status        string `json:"status"`
}

type HealthServiceClient interface {
	Check(ctx context.Context, in *HealthCheckRequest, opts ...grpc.CallOption) (*HealthCheckResponse, error)
}

type MobilityServiceClient interface {
	NearbyDrivers(ctx context.Context, in *NearbyDriversRequest, opts ...grpc.CallOption) (*NearbyDriversResponse, error)
	DriverLocation(ctx context.Context, in *DriverLocationRequest, opts ...grpc.CallOption) (*DriverLocationResponse, error)
}

type PaymentServiceClient interface {
	PaymentStatus(ctx context.Context, in *PaymentStatusRequest, opts ...grpc.CallOption) (*PaymentStatusResponse, error)
}

type HealthServiceServer interface {
	Check(context.Context, *HealthCheckRequest) (*HealthCheckResponse, error)
}

type MobilityServiceServer interface {
	NearbyDrivers(context.Context, *NearbyDriversRequest) (*NearbyDriversResponse, error)
	DriverLocation(context.Context, *DriverLocationRequest) (*DriverLocationResponse, error)
}

type PaymentServiceServer interface {
	PaymentStatus(context.Context, *PaymentStatusRequest) (*PaymentStatusResponse, error)
}

type healthServiceClient struct{ cc grpc.ClientConnInterface }

func NewHealthServiceClient(cc grpc.ClientConnInterface) HealthServiceClient {
	return &healthServiceClient{cc: cc}
}

func (c *healthServiceClient) Check(ctx context.Context, in *HealthCheckRequest, opts ...grpc.CallOption) (*HealthCheckResponse, error) {
	out := new(HealthCheckResponse)
	err := c.cc.Invoke(ctx, "/platform.v1.HealthService/Check", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type mobilityServiceClient struct{ cc grpc.ClientConnInterface }

func NewMobilityServiceClient(cc grpc.ClientConnInterface) MobilityServiceClient {
	return &mobilityServiceClient{cc: cc}
}

func (c *mobilityServiceClient) NearbyDrivers(ctx context.Context, in *NearbyDriversRequest, opts ...grpc.CallOption) (*NearbyDriversResponse, error) {
	out := new(NearbyDriversResponse)
	err := c.cc.Invoke(ctx, "/platform.v1.MobilityService/NearbyDrivers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mobilityServiceClient) DriverLocation(ctx context.Context, in *DriverLocationRequest, opts ...grpc.CallOption) (*DriverLocationResponse, error) {
	out := new(DriverLocationResponse)
	err := c.cc.Invoke(ctx, "/platform.v1.MobilityService/DriverLocation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type paymentServiceClient struct{ cc grpc.ClientConnInterface }

func NewPaymentServiceClient(cc grpc.ClientConnInterface) PaymentServiceClient {
	return &paymentServiceClient{cc: cc}
}

func (c *paymentServiceClient) PaymentStatus(ctx context.Context, in *PaymentStatusRequest, opts ...grpc.CallOption) (*PaymentStatusResponse, error) {
	out := new(PaymentStatusResponse)
	err := c.cc.Invoke(ctx, "/platform.v1.PaymentService/PaymentStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func RegisterHealthServiceServer(s grpc.ServiceRegistrar, srv HealthServiceServer) {
	s.RegisterService(&HealthService_ServiceDesc, srv)
}

func RegisterMobilityServiceServer(s grpc.ServiceRegistrar, srv MobilityServiceServer) {
	s.RegisterService(&MobilityService_ServiceDesc, srv)
}

func RegisterPaymentServiceServer(s grpc.ServiceRegistrar, srv PaymentServiceServer) {
	s.RegisterService(&PaymentService_ServiceDesc, srv)
}

var HealthService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "platform.v1.HealthService",
	HandlerType: (*HealthServiceServer)(nil),
	Methods: []grpc.MethodDesc{{
		MethodName: "Check",
		Handler:    _HealthService_Check_Handler,
	}},
}

var MobilityService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "platform.v1.MobilityService",
	HandlerType: (*MobilityServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "NearbyDrivers", Handler: _MobilityService_NearbyDrivers_Handler},
		{MethodName: "DriverLocation", Handler: _MobilityService_DriverLocation_Handler},
	},
}

var PaymentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "platform.v1.PaymentService",
	HandlerType: (*PaymentServiceServer)(nil),
	Methods: []grpc.MethodDesc{{
		MethodName: "PaymentStatus",
		Handler:    _PaymentService_PaymentStatus_Handler,
	}},
}

func _HealthService_Check_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(HealthCheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HealthServiceServer).Check(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/platform.v1.HealthService/Check"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(HealthServiceServer).Check(ctx, req.(*HealthCheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MobilityService_NearbyDrivers_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(NearbyDriversRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MobilityServiceServer).NearbyDrivers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/platform.v1.MobilityService/NearbyDrivers"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(MobilityServiceServer).NearbyDrivers(ctx, req.(*NearbyDriversRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MobilityService_DriverLocation_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(DriverLocationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MobilityServiceServer).DriverLocation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/platform.v1.MobilityService/DriverLocation"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(MobilityServiceServer).DriverLocation(ctx, req.(*DriverLocationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PaymentService_PaymentStatus_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(PaymentStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).PaymentStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/platform.v1.PaymentService/PaymentStatus"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(PaymentServiceServer).PaymentStatus(ctx, req.(*PaymentStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}
