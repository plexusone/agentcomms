package database

import (
	"context"
	"strings"
	"testing"

	_ "github.com/plexusone/agentcomms/ent/runtime" // Required for Ent privacy policies
)

func TestOpen_SQLite(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "opens in-memory SQLite",
			cfg: Config{
				Driver: DriverSQLite,
				DSN:    ":memory:",
			},
			wantErr: false,
		},
		{
			name: "opens file-based SQLite",
			cfg: Config{
				Driver: DriverSQLite,
				DSN:    t.TempDir() + "/test.db",
			},
			wantErr: false,
		},
		{
			name: "fails with empty DSN",
			cfg: Config{
				Driver: DriverSQLite,
				DSN:    "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Open(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if result == nil {
					t.Error("Open() returned nil result")
					return
				}
				if result.Client == nil {
					t.Error("Open() returned nil client")
				}
				if result.DB == nil {
					t.Error("Open() returned nil DB")
				}
				// Clean up
				result.Client.Close()
			}
		})
	}
}

func TestOpen_UnsupportedDriver(t *testing.T) {
	cfg := Config{
		Driver: DriverType("mysql"),
		DSN:    "localhost:3306",
	}

	_, err := Open(cfg)
	if err == nil {
		t.Error("Open() expected error for unsupported driver")
	}
	if !strings.Contains(err.Error(), "unsupported driver") {
		t.Errorf("Open() error = %v, want error containing 'unsupported driver'", err)
	}
}

func TestOpen_SQLiteWithMultiTenant(t *testing.T) {
	// Use a file-based SQLite to get proper pragma handling
	cfg := Config{
		Driver:      DriverSQLite,
		DSN:         t.TempDir() + "/multi_tenant_test.db",
		MultiTenant: true,
	}

	result, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer result.Client.Close()

	// Run schema migration with tenant context
	ctx := context.Background()
	if err := result.Client.Schema.Create(ctx); err != nil {
		t.Fatalf("Schema.Create() error = %v", err)
	}

	// Multi-tenant mode should work with SQLite (application-level filtering)
	if result.Client == nil {
		t.Error("expected client to be created")
	}
}

func TestDriverType(t *testing.T) {
	tests := []struct {
		driver DriverType
		want   string
	}{
		{DriverSQLite, "sqlite"},
		{DriverPostgres, "postgres"},
	}

	for _, tt := range tests {
		t.Run(string(tt.driver), func(t *testing.T) {
			if string(tt.driver) != tt.want {
				t.Errorf("DriverType = %q, want %q", tt.driver, tt.want)
			}
		})
	}
}

func TestConfig(t *testing.T) {
	// Test config struct initialization
	cfg := Config{
		Driver:      DriverPostgres,
		DSN:         "localhost:5432/db", // simplified DSN for test
		MultiTenant: true,
		UseRLS:      true,
	}

	if cfg.Driver != DriverPostgres {
		t.Errorf("Driver = %v, want %v", cfg.Driver, DriverPostgres)
	}
	if cfg.DSN == "" {
		t.Error("DSN should not be empty")
	}
	if !cfg.MultiTenant {
		t.Error("MultiTenant should be true")
	}
	if !cfg.UseRLS {
		t.Error("UseRLS should be true")
	}
}
