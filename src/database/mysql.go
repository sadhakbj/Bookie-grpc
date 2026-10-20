// Package database opens and tunes the MySQL connection pool used by the
// server.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	// Registers the "mysql" driver with database/sql.
	_ "github.com/go-sql-driver/mysql"

	"github.com/sadhakbj/bookie-grpc/src/internal/config"
)

// Pool limits stay modest for a single-instance demo service: ten concurrent
// connections, five kept idle, recycled every five minutes.
const (
	maxOpenConns    = 10
	maxIdleConns    = 5
	connMaxLifetime = 5 * time.Minute
	pingTimeout     = 5 * time.Second
)

// OpenMySQL opens the pool and fails fast when MySQL is unreachable.
func OpenMySQL(dbConfig config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", dbConfig.DataSourceName())
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	pingCtx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql at %s: %w", net.JoinHostPort(dbConfig.Host, dbConfig.Port), err)
	}

	return db, nil
}
