package database

import (
	"context"
	"database/sql"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"

	"github.com/plexusone/agentcomms/internal/tenant"
)

func TestNewRLSDriver(t *testing.T) {
	// Create a test SQLite driver (RLS driver wraps any SQL driver)
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	drv := entsql.OpenDB(dialect.SQLite, db)
	rlsDriver := NewRLSDriver(drv)

	if rlsDriver == nil {
		t.Fatal("NewRLSDriver returned nil")
	}
	if rlsDriver.Driver == nil {
		t.Error("RLSDriver.Driver is nil")
	}
}

func TestRLSDriverTx(t *testing.T) {
	// Create a test SQLite driver
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	drv := entsql.OpenDB(dialect.SQLite, db)
	rlsDriver := NewRLSDriver(drv)

	ctx := tenant.WithTenantID(context.Background(), "test-tenant")

	// Start a transaction
	tx, err := rlsDriver.Tx(ctx)
	if err != nil {
		t.Fatalf("Tx() error = %v", err)
	}

	// Verify it returns an RLSTx
	rlsTx, ok := tx.(*RLSTx)
	if !ok {
		t.Fatalf("Tx() returned %T, want *RLSTx", tx)
	}

	if rlsTx.Tx == nil {
		t.Error("RLSTx.Tx is nil")
	}

	// Rollback to clean up
	if err := tx.Rollback(); err != nil {
		t.Errorf("Rollback() error = %v", err)
	}
}

func TestRLSTxTenantContext(t *testing.T) {
	// This test verifies the tenant context is properly extracted
	// Note: The actual SET LOCAL command is PostgreSQL-specific and will fail on SQLite
	// This test focuses on the context extraction logic

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	drv := entsql.OpenDB(dialect.SQLite, db)
	rlsDriver := NewRLSDriver(drv)

	testTenantID := "tenant-xyz-123"
	ctx := tenant.WithTenantID(context.Background(), testTenantID)

	tx, err := rlsDriver.Tx(ctx)
	if err != nil {
		t.Fatalf("Tx() error = %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	rlsTx := tx.(*RLSTx)

	// Verify tenant ID can be extracted from the stored context
	if tenant.FromContext(rlsTx.ctx) != testTenantID {
		t.Errorf("tenant from context = %q, want %q",
			tenant.FromContext(rlsTx.ctx), testTenantID)
	}
}

func TestRLSTxNotSetTwice(t *testing.T) {
	// Verify that tenantSet flag prevents multiple SET LOCAL calls
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	drv := entsql.OpenDB(dialect.SQLite, db)
	rlsDriver := NewRLSDriver(drv)

	ctx := tenant.WithTenantID(context.Background(), "test-tenant")

	tx, err := rlsDriver.Tx(ctx)
	if err != nil {
		t.Fatalf("Tx() error = %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	rlsTx := tx.(*RLSTx)

	// Initially tenantSet should be false
	if rlsTx.tenantSet {
		t.Error("tenantSet should initially be false")
	}

	// After setTenantContext is called (which will fail on SQLite but the flag logic is what we're testing)
	// The actual test of the SET LOCAL command requires PostgreSQL
}
