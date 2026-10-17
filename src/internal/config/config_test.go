package config

import (
	"os"
	"testing"
)

func TestLoadServerConfig(t *testing.T) {
	clearEnvironment(t, "PORT", "OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_METRICS_OTLP_ENDPOINT")
	got, err := LoadServerConfig()
	if err != nil {
		t.Fatalf("LoadServerConfig() error = %v", err)
	}
	if got.Port != "8020" || got.OTLPEndpoint != "localhost:4317" || got.OTLPMetricsEndpoint != "localhost:4318" {
		t.Errorf("LoadServerConfig() = %+v, want documented defaults", got)
	}
}

func TestLoadClientConfigRejectsInvalidPort(t *testing.T) {
	clearEnvironment(t, "HTTP_PORT", "GRPC_SERVER_ADDR", "OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_METRICS_OTLP_ENDPOINT")
	t.Setenv("HTTP_PORT", "65536")
	if _, err := LoadClientConfig(); err == nil {
		t.Fatal("LoadClientConfig() accepted an out-of-range port")
	}
}

func TestValidateRejectsEmptyEndpoint(t *testing.T) {
	loadedConfig := ServerConfig{
		Port:                "8020",
		OTLPEndpoint:        "",
		OTLPMetricsEndpoint: "localhost:4318",
	}
	if err := Validate(loadedConfig); err == nil {
		t.Fatal("Validate() accepted an empty endpoint")
	}
}

func clearEnvironment(t *testing.T, names ...string) {
	t.Helper()
	previous := make(map[string]string)
	present := make(map[string]bool)
	for _, name := range names {
		if value, ok := os.LookupEnv(name); ok {
			previous[name] = value
			present[name] = true
		}
		_ = os.Unsetenv(name)
	}
	t.Cleanup(func() {
		for _, name := range names {
			if present[name] {
				_ = os.Setenv(name, previous[name])
				continue
			}
			_ = os.Unsetenv(name)
		}
	})
}
