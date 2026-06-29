package config

import "github.com/spf13/viper"

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	GRPC      GRPCConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	Kafka     KafkaConfig
	Temporal  TemporalConfig
	Telemetry TelemetryConfig
	JWT       JWTConfig
	Money     MoneyConfig
	Security  SecurityConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         string `mapstructure:"PORT"`
	Environment  string `mapstructure:"ENV"`
	Version      string `mapstructure:"VERSION"`
	ReadTimeout  int    `mapstructure:"READ_TIMEOUT_SECONDS"`
	WriteTimeout int    `mapstructure:"WRITE_TIMEOUT_SECONDS"`
	IdleTimeout  int    `mapstructure:"IDLE_TIMEOUT_SECONDS"`
	TLSCertFile  string `mapstructure:"TLS_CERT_FILE"`
	TLSKeyFile   string `mapstructure:"TLS_KEY_FILE"`
	ClientCAFile string `mapstructure:"TLS_CLIENT_CA_FILE"`
}

type GRPCConfig struct {
	Port                  string `mapstructure:"GRPC_PORT"`
	IdentityServiceAddr   string `mapstructure:"IDENTITY_GRPC_ADDR"`
	LogisticsServiceAddr  string `mapstructure:"LOGISTICS_GRPC_ADDR"`
	MobilityServiceAddr   string `mapstructure:"MOBILITY_GRPC_ADDR"`
	PaymentServiceAddr    string `mapstructure:"PAYMENT_GRPC_ADDR"`
	OperationsServiceAddr string `mapstructure:"OPERATIONS_GRPC_ADDR"`
	MCPServiceAddr        string `mapstructure:"MCP_GRPC_ADDR"`
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	DBName   string `mapstructure:"DB_NAME"`
	SSLMode  string `mapstructure:"DB_SSLMODE"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string `mapstructure:"REDIS_ADDR"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Enabled bool     `mapstructure:"KAFKA_ENABLED"`
	Brokers []string `mapstructure:"KAFKA_BROKERS"`
	GroupID string   `mapstructure:"KAFKA_GROUP_ID"`
	// Schema Registry URL
	SchemaRegistryURL        string `mapstructure:"KAFKA_SCHEMA_REGISTRY_URL"`
	TopicPrefix              string `mapstructure:"KAFKA_TOPIC_PREFIX"`
	OutboxPollIntervalMillis int    `mapstructure:"KAFKA_OUTBOX_POLL_INTERVAL_MILLIS"`
}

// TemporalConfig holds Temporal configuration
type TemporalConfig struct {
	HostPort  string `mapstructure:"TEMPORAL_HOST_PORT"`
	Namespace string `mapstructure:"TEMPORAL_NAMESPACE"`
	TaskQueue string `mapstructure:"TEMPORAL_TASK_QUEUE"`
}

type TelemetryConfig struct {
	JaegerEndpoint string `mapstructure:"JAEGER_ENDPOINT"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	PrivateKey    string `mapstructure:"JWT_PRIVATE_KEY"`
	PublicKey     string `mapstructure:"JWT_PUBLIC_KEY"`
	RefreshSecret string `mapstructure:"JWT_REFRESH_SECRET"`
	AccessTTL     int    `mapstructure:"JWT_ACCESS_TTL_SECONDS"`  // 15 minutes
	RefreshTTL    int    `mapstructure:"JWT_REFRESH_TTL_SECONDS"` // 7 days
}

// MoneyConfig holds money-related configuration
type MoneyConfig struct {
	DefaultCurrency string `mapstructure:"DEFAULT_CURRENCY"`
}

type SecurityConfig struct {
	InternalToken        string   `mapstructure:"INTERNAL_AUTH_TOKEN"`
	CORSAllowedOrigins   []string `mapstructure:"CORS_ALLOWED_ORIGINS"`
	TrustedProxyCIDRs    []string `mapstructure:"TRUSTED_PROXY_CIDRS"`
	MTLSRequired         bool     `mapstructure:"INTERNAL_MTLS_REQUIRED"`
	MTLSClientHeader     string   `mapstructure:"INTERNAL_MTLS_CLIENT_HEADER"`
	RequireStrongSecrets bool     `mapstructure:"REQUIRE_STRONG_SECRETS"`
	AnomalyThreshold     int      `mapstructure:"ANOMALY_THRESHOLD"`
	AnomalyWindowSeconds int      `mapstructure:"ANOMALY_WINDOW_SECONDS"`
}

// LoadConfig loads configuration from environment variables and optional config file
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Allow environment variables
	v.AutomaticEnv()

	// If config path provided, read from file
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	// Set defaults
	setDefaults(v)

	cfg := &Config{
		Server: ServerConfig{
			Port:         v.GetString("PORT"),
			Environment:  v.GetString("ENV"),
			Version:      v.GetString("VERSION"),
			ReadTimeout:  v.GetInt("READ_TIMEOUT_SECONDS"),
			WriteTimeout: v.GetInt("WRITE_TIMEOUT_SECONDS"),
			IdleTimeout:  v.GetInt("IDLE_TIMEOUT_SECONDS"),
			TLSCertFile:  v.GetString("TLS_CERT_FILE"),
			TLSKeyFile:   v.GetString("TLS_KEY_FILE"),
			ClientCAFile: v.GetString("TLS_CLIENT_CA_FILE"),
		},
		GRPC: GRPCConfig{
			Port:                  v.GetString("GRPC_PORT"),
			IdentityServiceAddr:   v.GetString("IDENTITY_GRPC_ADDR"),
			LogisticsServiceAddr:  v.GetString("LOGISTICS_GRPC_ADDR"),
			MobilityServiceAddr:   v.GetString("MOBILITY_GRPC_ADDR"),
			PaymentServiceAddr:    v.GetString("PAYMENT_GRPC_ADDR"),
			OperationsServiceAddr: v.GetString("OPERATIONS_GRPC_ADDR"),
			MCPServiceAddr:        v.GetString("MCP_GRPC_ADDR"),
		},
		Database: DatabaseConfig{
			Host:     v.GetString("DB_HOST"),
			Port:     v.GetString("DB_PORT"),
			User:     v.GetString("DB_USER"),
			Password: v.GetString("DB_PASSWORD"),
			DBName:   v.GetString("DB_NAME"),
			SSLMode:  v.GetString("DB_SSLMODE"),
		},
		Redis: RedisConfig{
			Addr:     v.GetString("REDIS_ADDR"),
			Password: v.GetString("REDIS_PASSWORD"),
			DB:       v.GetInt("REDIS_DB"),
		},
		Kafka: KafkaConfig{
			Enabled:                  v.GetBool("KAFKA_ENABLED"),
			Brokers:                  v.GetStringSlice("KAFKA_BROKERS"),
			GroupID:                  v.GetString("KAFKA_GROUP_ID"),
			SchemaRegistryURL:        v.GetString("KAFKA_SCHEMA_REGISTRY_URL"),
			TopicPrefix:              v.GetString("KAFKA_TOPIC_PREFIX"),
			OutboxPollIntervalMillis: v.GetInt("KAFKA_OUTBOX_POLL_INTERVAL_MILLIS"),
		},
		Temporal: TemporalConfig{
			HostPort:  v.GetString("TEMPORAL_HOST_PORT"),
			Namespace: v.GetString("TEMPORAL_NAMESPACE"),
			TaskQueue: v.GetString("TEMPORAL_TASK_QUEUE"),
		},
		Telemetry: TelemetryConfig{
			JaegerEndpoint: v.GetString("JAEGER_ENDPOINT"),
		},
		JWT: JWTConfig{
			PrivateKey:    v.GetString("JWT_PRIVATE_KEY"),
			PublicKey:     v.GetString("JWT_PUBLIC_KEY"),
			RefreshSecret: v.GetString("JWT_REFRESH_SECRET"),
			AccessTTL:     v.GetInt("JWT_ACCESS_TTL_SECONDS"),
			RefreshTTL:    v.GetInt("JWT_REFRESH_TTL_SECONDS"),
		},
		Money: MoneyConfig{
			DefaultCurrency: v.GetString("DEFAULT_CURRENCY"),
		},
		Security: SecurityConfig{
			InternalToken:        v.GetString("INTERNAL_AUTH_TOKEN"),
			CORSAllowedOrigins:   v.GetStringSlice("CORS_ALLOWED_ORIGINS"),
			TrustedProxyCIDRs:    v.GetStringSlice("TRUSTED_PROXY_CIDRS"),
			MTLSRequired:         v.GetBool("INTERNAL_MTLS_REQUIRED"),
			MTLSClientHeader:     v.GetString("INTERNAL_MTLS_CLIENT_HEADER"),
			RequireStrongSecrets: v.GetBool("REQUIRE_STRONG_SECRETS"),
			AnomalyThreshold:     v.GetInt("ANOMALY_THRESHOLD"),
			AnomalyWindowSeconds: v.GetInt("ANOMALY_WINDOW_SECONDS"),
		},
	}

	return cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("PORT", "8080")
	v.SetDefault("ENV", "development")
	v.SetDefault("VERSION", "v1.0.0")
	v.SetDefault("READ_TIMEOUT_SECONDS", 15)
	v.SetDefault("WRITE_TIMEOUT_SECONDS", 15)
	v.SetDefault("IDLE_TIMEOUT_SECONDS", 60)
	v.SetDefault("TLS_CERT_FILE", "")
	v.SetDefault("TLS_KEY_FILE", "")
	v.SetDefault("TLS_CLIENT_CA_FILE", "")
	v.SetDefault("GRPC_PORT", "9080")
	v.SetDefault("IDENTITY_GRPC_ADDR", "identity-service:9081")
	v.SetDefault("LOGISTICS_GRPC_ADDR", "logistics-service:9082")
	v.SetDefault("MOBILITY_GRPC_ADDR", "mobility-service:9083")
	v.SetDefault("PAYMENT_GRPC_ADDR", "payment-service:9084")
	v.SetDefault("OPERATIONS_GRPC_ADDR", "operations-service:9085")
	v.SetDefault("MCP_GRPC_ADDR", "mcp-server:9090")

	// Database defaults
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_PASSWORD", "postgres")
	v.SetDefault("DB_NAME", "logistics")
	v.SetDefault("DB_SSLMODE", "disable")

	// Redis defaults
	v.SetDefault("REDIS_ADDR", "localhost:6379")
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)

	// Kafka defaults
	v.SetDefault("KAFKA_ENABLED", false)
	v.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	v.SetDefault("KAFKA_GROUP_ID", "logistics-platform")
	v.SetDefault("KAFKA_SCHEMA_REGISTRY_URL", "http://localhost:8081")
	v.SetDefault("KAFKA_TOPIC_PREFIX", "platform.")
	v.SetDefault("KAFKA_OUTBOX_POLL_INTERVAL_MILLIS", 1000)

	// Temporal defaults
	v.SetDefault("TEMPORAL_HOST_PORT", "localhost:7233")
	v.SetDefault("TEMPORAL_NAMESPACE", "default")
	v.SetDefault("TEMPORAL_TASK_QUEUE", "logistics-task-queue")

	// Telemetry defaults
	v.SetDefault("JAEGER_ENDPOINT", "localhost:4317")

	// JWT defaults (these should be overridden in production)
	v.SetDefault("JWT_PRIVATE_KEY", "")
	v.SetDefault("JWT_PUBLIC_KEY", "")
	v.SetDefault("JWT_REFRESH_SECRET", "")
	v.SetDefault("JWT_ACCESS_TTL_SECONDS", 900)     // 15 minutes
	v.SetDefault("JWT_REFRESH_TTL_SECONDS", 604800) // 7 days

	// Money defaults
	v.SetDefault("DEFAULT_CURRENCY", "NGN")

	// Security defaults
	v.SetDefault("INTERNAL_AUTH_TOKEN", "")
	v.SetDefault("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:19006"})
	v.SetDefault("TRUSTED_PROXY_CIDRS", []string{})
	v.SetDefault("INTERNAL_MTLS_REQUIRED", false)
	v.SetDefault("INTERNAL_MTLS_CLIENT_HEADER", "X-Client-Cert-Verified")
	v.SetDefault("REQUIRE_STRONG_SECRETS", false)
	v.SetDefault("ANOMALY_THRESHOLD", 8)
	v.SetDefault("ANOMALY_WINDOW_SECONDS", 300)
}
