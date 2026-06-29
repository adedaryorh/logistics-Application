package paymentgrpc

import (
	"context"

	"github.com/adedaryorh/logistics-platform/pkg/platformrpc"
	"google.golang.org/grpc"
)

type PaymentStatusRequest = platformrpc.PaymentStatusRequest
type PaymentStatusResponse = platformrpc.PaymentStatusResponse
type Server = platformrpc.PaymentServiceServer

type Client interface {
	PaymentStatus(ctx context.Context, req *PaymentStatusRequest) (*PaymentStatusResponse, error)
}

type client struct {
	inner platformrpc.PaymentServiceClient
}

func NewFromPlatform(inner platformrpc.PaymentServiceClient) Client {
	return &client{inner: inner}
}

func NewFromConn(conn grpc.ClientConnInterface) Client {
	return NewFromPlatform(platformrpc.NewPaymentServiceClient(conn))
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
	platformrpc.RegisterPaymentServiceServer(server, srv)
}

func (c *client) PaymentStatus(ctx context.Context, req *PaymentStatusRequest) (*PaymentStatusResponse, error) {
	return c.inner.PaymentStatus(ctx, req)
}
