package mobilitygrpc

import (
	"context"

	"github.com/adedaryorh/logistics-platform/pkg/platformrpc"
	"google.golang.org/grpc"
)

type NearbyDriversRequest = platformrpc.NearbyDriversRequest
type NearbyDriversResponse = platformrpc.NearbyDriversResponse
type DriverLocationRequest = platformrpc.DriverLocationRequest
type DriverLocationResponse = platformrpc.DriverLocationResponse
type DriverSnapshot = platformrpc.DriverSnapshot
type Server = platformrpc.MobilityServiceServer

type Client interface {
	NearbyDrivers(ctx context.Context, req *NearbyDriversRequest) (*NearbyDriversResponse, error)
	DriverLocation(ctx context.Context, req *DriverLocationRequest) (*DriverLocationResponse, error)
}

type client struct {
	inner platformrpc.MobilityServiceClient
}

func NewFromPlatform(inner platformrpc.MobilityServiceClient) Client {
	return &client{inner: inner}
}

func NewFromConn(conn grpc.ClientConnInterface) Client {
	return NewFromPlatform(platformrpc.NewMobilityServiceClient(conn))
}

func Dial(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return platformrpc.DialContext(ctx, target, opts...)
}

func DialClient(ctx context.Context, target string, opts ...grpc.DialOption) (Client, *grpc.ClientConn, error) {
	conn, err := Dial(ctx, target, opts...)
	if err != nil {
		return nil, nil, err
	}
	return NewFromConn(conn), conn, nil
}

func RegisterServer(server grpc.ServiceRegistrar, srv Server) {
	platformrpc.RegisterMobilityServiceServer(server, srv)
}

func (c *client) NearbyDrivers(ctx context.Context, req *NearbyDriversRequest) (*NearbyDriversResponse, error) {
	return c.inner.NearbyDrivers(ctx, req)
}

func (c *client) DriverLocation(ctx context.Context, req *DriverLocationRequest) (*DriverLocationResponse, error) {
	return c.inner.DriverLocation(ctx, req)
}
