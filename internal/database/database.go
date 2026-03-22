// Package database provides database abstraction for SQLite and PostgreSQL.
package database

import (
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"  // PostgreSQL driver
	_ "modernc.org/sqlite" // SQLite driver

	"github.com/plexusone/agentcomms/ent"
)

// DriverType identifies the database driver.
type DriverType string

const (
	// DriverSQLite uses modernc.org/sqlite (pure Go).
	DriverSQLite DriverType = "sqlite"
	// DriverPostgres uses lib/pq for PostgreSQL.
	DriverPostgres DriverType = "postgres"
)

// Config holds database configuration.
type Config struct {
	// Driver is the database driver type ("sqlite" or "postgres").
	Driver DriverType

	// DSN is the data source name (connection string).
	// For SQLite: file path or :memory:
	// For Postgres: postgres://user:pass@host:port/dbname?sslmode=disable
	DSN string

	// MultiTenant enables multi-tenancy mode (requires tenant_id per request).
	MultiTenant bool

	// UseRLS enables PostgreSQL Row-Level Security (PostgreSQL only).
	// When enabled, RLS policies are applied after schema creation.
	UseRLS bool
}

// OpenResult contains the result of opening a database connection.
type OpenResult struct {
	// Client is the Ent client.
	Client *ent.Client

	// DB is the underlying *sql.DB for direct database operations.
	// Used for applying RLS policies on PostgreSQL.
	DB *sql.DB
}

// Open creates an Ent client for the configured database.
func Open(cfg Config) (*OpenResult, error) {
	switch cfg.Driver {
	case DriverSQLite:
		return openSQLite(cfg)
	case DriverPostgres:
		return openPostgres(cfg)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}
}

// openSQLite opens a SQLite database.
func openSQLite(cfg Config) (*OpenResult, error) {
	// SQLite connection string with WAL mode for better concurrency
	dsn := cfg.DSN
	if dsn == "" {
		return nil, fmt.Errorf("sqlite dsn is required")
	}

	// Add pragmas if not already present
	if dsn != ":memory:" && dsn[0] != '?' {
		dsn = fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", dsn)
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

	drv := entsql.OpenDB(dialect.SQLite, db)
	return &OpenResult{
		Client: ent.NewClient(ent.Driver(drv)),
		DB:     db,
	}, nil
}

// openPostgres opens a PostgreSQL database.
func openPostgres(cfg Config) (*OpenResult, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("postgres dsn is required")
	}

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)

	// If RLS is enabled, wrap the driver to set tenant context per transaction
	if cfg.UseRLS {
		return &OpenResult{
			Client: ent.NewClient(ent.Driver(NewRLSDriver(drv))),
			DB:     db,
		}, nil
	}

	return &OpenResult{
		Client: ent.NewClient(ent.Driver(drv)),
		DB:     db,
	}, nil
}
