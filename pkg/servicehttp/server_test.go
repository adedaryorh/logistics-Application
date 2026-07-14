package servicehttp

import (
	"testing"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"github.com/adedaryorh/logistics-platform/pkg/config"
)

func TestOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{
			name:    "service name is required",
			opts:    Options{},
			wantErr: true,
		},
		{
			name: "simple route registration",
			opts: Options{
				ServiceName: "gateway",
				Register:    func(*gin.Engine, *config.Config) {},
			},
		},
		{
			name: "app factory",
			opts: Options{
				ServiceName: "orders",
				NewApp:      func(*config.Config) App { return nil },
			},
		},
		{
			name: "app factory and route callback are ambiguous",
			opts: Options{
				ServiceName: "orders",
				NewApp:      func(*config.Config) App { return nil },
				Register:    func(*gin.Engine, *config.Config) {},
			},
			wantErr: true,
		},
		{
			name: "app factory and grpc callback are ambiguous",
			opts: Options{
				ServiceName:  "orders",
				NewApp:       func(*config.Config) App { return nil },
				RegisterGRPC: func(*grpc.Server, *config.Config) {},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
