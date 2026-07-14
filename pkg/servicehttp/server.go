package servicehttp

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/adedaryorh/logistics-platform/pkg/config"
	"github.com/adedaryorh/logistics-platform/pkg/logger"
	"github.com/adedaryorh/logistics-platform/pkg/middleware"
	platformotel "github.com/adedaryorh/logistics-platform/pkg/otel"
	"github.com/adedaryorh/logistics-platform/pkg/platformrpc"
)

type Options struct {
	ServiceName  string
	GRPCPort     string
	Register     func(*gin.Engine, *config.Config)
	NewApp       func(*config.Config) App
	RegisterGRPC func(*grpc.Server, *config.Config)
}

// App captures the lifecycle shared by the platform's stateful services.
// Keeping this wiring here leaves service entrypoints responsible only for
// naming the service, selecting its ports, and constructing the application.
type App interface {
	RegisterRoutes(*gin.Engine)
	RegisterGRPC(*grpc.Server, *config.Config)
	StartBackground(context.Context, *config.Config) error
}

func Run(opts Options) error {
	if err := opts.validate(); err != nil {
		return err
	}

	cfg, err := config.LoadConfig("")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if err := config.ValidateSecurity(cfg); err != nil {
		return fmt.Errorf("validate security config: %w", err)
	}

	log, err := logger.New(opts.ServiceName, cfg.Server.Environment, cfg.Server.Version)
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}
	defer log.Sync()

	shutdownTracer, err := platformotel.InitTracer(opts.ServiceName, cfg.Telemetry.JaegerEndpoint)
	if err != nil {
		return fmt.Errorf("initialize tracer: %w", err)
	}
	defer shutdownTracer()

	var (
		registerRoutes func(*gin.Engine)
		background     func(context.Context) error
	)

	if opts.NewApp != nil {
		app := opts.NewApp(cfg)
		if app == nil {
			return fmt.Errorf("setup service: app factory returned nil")
		}
		registerRoutes = app.RegisterRoutes
		background = func(ctx context.Context) error {
			return app.StartBackground(ctx, cfg)
		}
		opts.RegisterGRPC = app.RegisterGRPC
	} else if opts.Register != nil {
		registerRoutes = func(router *gin.Engine) {
			opts.Register(router, cfg)
		}
	}

	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	platformrpc.RegisterCodec()
	router := gin.New()
	metrics := middleware.NewMetricsCollector()
	router.Use(middleware.Recovery(log))
	router.Use(middleware.RequestID())
	router.Use(metrics.Middleware())
	router.Use(middleware.AbuseProtection(cfg))
	router.Use(middleware.Logger(log))
	router.Use(middleware.Tracing(nil))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"status":  "ok",
				"service": opts.ServiceName,
			},
			"error": nil,
			"meta": gin.H{
				"request_id": c.GetString("request_id"),
				"timestamp":  time.Now().UTC(),
				"version":    cfg.Server.Version,
			},
		})
	})

	router.GET("/metrics", metrics.Endpoint(opts.ServiceName))

	if registerRoutes != nil {
		registerRoutes(router)
	}

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}
	grpcPort := cfg.GRPC.Port
	if opts.GRPCPort != "" {
		grpcPort = opts.GRPCPort
	}
	var (
		grpcServer   *grpc.Server
		grpcListener net.Listener
	)
	if opts.RegisterGRPC != nil {
		grpcServer = grpc.NewServer()
		opts.RegisterGRPC(grpcServer, cfg)
		grpcListener, err = net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			return fmt.Errorf("listen grpc: %w", err)
		}
	}
	if cfg.Server.TLSCertFile != "" && cfg.Server.TLSKeyFile != "" {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		if cfg.Security.MTLSRequired {
			clientCA, err := os.ReadFile(cfg.Server.ClientCAFile)
			if err != nil {
				return fmt.Errorf("read client ca file: %w", err)
			}
			clientPool := x509.NewCertPool()
			if !clientPool.AppendCertsFromPEM(clientCA) {
				return fmt.Errorf("append client ca certs")
			}
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
			tlsConfig.ClientCAs = clientPool
		}
		server.TLSConfig = tlsConfig
	}

	// HTTP, gRPC, and the background worker can fail independently. Buffer one
	// result per component so a second failure never leaves a goroutine blocked
	// while shutdown is already in progress.
	errCh := make(chan error, 3)
	bgDone := make(chan struct{})
	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()

	if background != nil {
		go func() {
			defer close(bgDone)
			if err := background(bgCtx); err != nil && !errors.Is(err, context.Canceled) {
				errCh <- fmt.Errorf("run background worker: %w", err)
			}
		}()
	} else {
		close(bgDone)
	}

	go func() {
		log.Info("starting service",
			zap.String("address", server.Addr),
			zap.String("environment", cfg.Server.Environment),
		)
		var err error
		if cfg.Server.TLSCertFile != "" && cfg.Server.TLSKeyFile != "" {
			err = server.ListenAndServeTLS(cfg.Server.TLSCertFile, cfg.Server.TLSKeyFile)
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("serve http: %w", err)
		}
	}()
	if grpcServer != nil {
		go func() {
			log.Info("starting grpc server", zap.String("address", ":"+grpcPort))
			if err := grpcServer.Serve(grpcListener); err != nil {
				errCh <- fmt.Errorf("serve grpc: %w", err)
			}
		}()
	}

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var runErr error
	select {
	case runErr = <-errCh:
	case <-stopCtx.Done():
	}

	log.Info("shutting down service")
	bgCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var shutdownErr error
	if err := server.Shutdown(ctx); err != nil {
		shutdownErr = fmt.Errorf("shutdown http server: %w", err)
	}
	if grpcServer != nil {
		stopped := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
		case <-ctx.Done():
			grpcServer.Stop()
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown grpc server: %w", ctx.Err()))
		}
	}
	select {
	case <-bgDone:
	case <-ctx.Done():
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown background worker: %w", ctx.Err()))
	}

	log.Info("service stopped")
	return errors.Join(runErr, shutdownErr)
}

func (opts Options) validate() error {
	if opts.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}
	if opts.NewApp != nil && (opts.Register != nil || opts.RegisterGRPC != nil) {
		return fmt.Errorf("new app cannot be combined with route or grpc registration callbacks")
	}
	return nil
}
