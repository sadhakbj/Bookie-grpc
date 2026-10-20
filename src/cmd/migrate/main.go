// Command migrate applies the goose SQL migrations embedded from the
// migrations package against the configured MySQL books store.
package main

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"log"
	"log/slog"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	"github.com/sadhakbj/bookie-grpc/src/internal/config"
	"github.com/sadhakbj/bookie-grpc/src/migrations"
)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatal("Error loading .env file: ", err)
	}

	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		log.Fatal("Invalid database configuration: ", err)
	}

	db, err := sql.Open("mysql", dbConfig.DataSourceName())
	if err != nil {
		log.Fatal("Could not open MySQL connection: ", err)
	}
	defer func() { _ = db.Close() }()

	goose.SetLogger(slog.NewLogLogger(slog.Default().Handler(), slog.LevelInfo))
	goose.SetBaseFS(migrations.SQL)
	if err := goose.SetDialect("mysql"); err != nil {
		log.Fatal("Could not set MySQL dialect: ", err)
	}

	if err := goose.UpContext(context.Background(), db, "."); err != nil {
		log.Fatal("Migration failed: ", err)
	}
}
