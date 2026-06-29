package healthgrpc

import (
	"context"

	"github.com/adedaryorh/logistics-platform/pkg/platformrpc"
	"google.golang.org/grpc"
)

type CheckRequest = platformrpc.HealthCheckRequest
type CheckResponse = platformrpc.HealthCheckResponse
type Server = platformrpc.HealthServiceServer

type Client interface {
	Check(ctx context.Context, req *CheckRequest) (*CheckResponse, error)
}

type client struct {
	inner platformrpc.HealthServiceClient
}

func NewFromPlatform(inner platformrpc.HealthServiceClient) Client {
	return &client{inner: inner}
}

func NewFromConn(conn grpc.ClientConnInterface) Client {
	return NewFromPlatform(platformrpc.NewHealthServiceClient(conn))
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
	platformrpc.RegisterHealthServiceServer(server, srv)
}

func (c *client) Check(ctx context.Context, req *CheckRequest) (*CheckResponse, error) {
	return c.inner.Check(ctx, req)
}
