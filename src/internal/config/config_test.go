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

func TestLoadDatabaseConfig(t *testing.T) {
	clearEnvironment(t, "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME")
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "root")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "bookie")
	got, err := LoadDatabaseConfig()
	if err != nil {
		t.Fatalf("LoadDatabaseConfig() error = %v", err)
	}
	if got.Host != "127.0.0.1" || got.Port != "3306" || got.User != "root" || got.Name != "bookie" {
		t.Errorf("LoadDatabaseConfig() = %+v, want env values", got)
	}
}

func TestLoadDatabaseConfigRejectsInvalidPort(t *testing.T) {
	clearEnvironment(t, "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME")
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_PORT", "65536")
	t.Setenv("DB_USER", "root")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "bookie")
	if _, err := LoadDatabaseConfig(); err == nil {
		t.Fatal("LoadDatabaseConfig() accepted an out-of-range port")
	}
}

func TestLoadDatabaseConfigRejectsMissingHost(t *testing.T) {
	clearEnvironment(t, "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "root")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "bookie")
	if _, err := LoadDatabaseConfig(); err == nil {
		t.Fatal("LoadDatabaseConfig() accepted missing DB_HOST")
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
