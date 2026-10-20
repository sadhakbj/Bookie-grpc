package config

import (
	"fmt"
	"net"

	"github.com/go-sql-driver/mysql"
)

// DatabaseConfig contains MySQL runtime configuration.
type DatabaseConfig struct {
	Host     string `env:"DB_HOST" validate:"required"`
	Port     string `env:"DB_PORT" validate:"required,port"`
	User     string `env:"DB_USER" validate:"required"`
	Password string `env:"DB_PASSWORD" validate:"required"`
	Name     string `env:"DB_NAME" validate:"required"`
}

// LoadDatabaseConfig loads and validates database configuration.
func LoadDatabaseConfig() (DatabaseConfig, error) {
	var loadedConfig DatabaseConfig
	if err := Load(&loadedConfig); err != nil {
		return DatabaseConfig{}, fmt.Errorf("load database configuration: %w", err)
	}
	if err := Validate(loadedConfig); err != nil {
		return DatabaseConfig{}, fmt.Errorf("validate database configuration: %w", err)
	}
	return loadedConfig, nil
}

// DataSourceName builds the DSN accepted by go-sql-driver/mysql.
func (c DatabaseConfig) DataSourceName() string {
	driverConfig := mysql.NewConfig()
	driverConfig.User = c.User
	driverConfig.Passwd = c.Password
	driverConfig.Net = "tcp"
	driverConfig.Addr = net.JoinHostPort(c.Host, c.Port)
	driverConfig.DBName = c.Name
	driverConfig.ParseTime = true
	return driverConfig.FormatDSN()
}
