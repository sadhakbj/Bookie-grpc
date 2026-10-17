package config

import "fmt"

// ClientConfig contains HTTP client runtime configuration.
type ClientConfig struct {
	HTTPPort            string `env:"HTTP_PORT,default=8080" validate:"required,port"`
	GRPCServerAddr      string `env:"GRPC_SERVER_ADDR,default=localhost:8020" validate:"required,hostname_port"`
	OTLPEndpoint        string `env:"OTEL_EXPORTER_OTLP_ENDPOINT,default=localhost:4317" validate:"required,hostname_port"`
	OTLPMetricsEndpoint string `env:"OTEL_METRICS_OTLP_ENDPOINT,default=localhost:4318" validate:"required,hostname_port"`
}

// LoadClientConfig loads and validates HTTP client configuration.
func LoadClientConfig() (ClientConfig, error) {
	var loadedConfig ClientConfig
	if err := Load(&loadedConfig); err != nil {
		return ClientConfig{}, fmt.Errorf("load client configuration: %w", err)
	}
	if err := Validate(loadedConfig); err != nil {
		return ClientConfig{}, fmt.Errorf("validate client configuration: %w", err)
	}
	return loadedConfig, nil
}
