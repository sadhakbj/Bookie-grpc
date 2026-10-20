package config

import "fmt"

// ServerConfig contains gRPC server runtime configuration.
type ServerConfig struct {
	Port                string `env:"PORT,default=8020" validate:"required,port"`
	OTLPEndpoint        string `env:"OTEL_EXPORTER_OTLP_ENDPOINT,default=localhost:4317" validate:"required,hostname_port"`
	OTLPLogsEndpoint    string `env:"OTEL_LOGS_OTLP_ENDPOINT,default=localhost:31081" validate:"required,hostname_port"`
	OTLPMetricsEndpoint string `env:"OTEL_METRICS_OTLP_ENDPOINT,default=localhost:4318" validate:"required,hostname_port"`
}

// LoadServerConfig loads and validates gRPC server configuration.
func LoadServerConfig() (ServerConfig, error) {
	var loadedConfig ServerConfig
	if err := Load(&loadedConfig); err != nil {
		return ServerConfig{}, fmt.Errorf("load server configuration: %w", err)
	}
	if err := Validate(loadedConfig); err != nil {
		return ServerConfig{}, fmt.Errorf("validate server configuration: %w", err)
	}
	return loadedConfig, nil
}
