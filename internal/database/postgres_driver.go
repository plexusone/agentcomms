package database

import (
	"context"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/plexusone/agentcomms/internal/tenant"
)

// RLSDriver wraps an Ent SQL driver to set the tenant context on each transaction.
// This enables PostgreSQL Row-Level Security by setting app.current_tenant.
type RLSDriver struct {
	*entsql.Driver
}

// NewRLSDriver wraps an Ent SQL driver with RLS tenant context support.
func NewRLSDriver(drv *entsql.Driver) *RLSDriver {
	return &RLSDriver{Driver: drv}
}

// Tx starts a new transaction and sets the tenant context using SET LOCAL.
func (d *RLSDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	tx, err := d.Driver.Tx(ctx)
	if err != nil {
		return nil, err
	}
	return &RLSTx{Tx: tx.(*entsql.Tx), ctx: ctx}, nil
}

// RLSTx wraps an Ent transaction to set tenant context on first operation.
type RLSTx struct {
	*entsql.Tx
	ctx       context.Context
	tenantSet bool
}

// setTenantContext sets the app.current_tenant session variable.
func (tx *RLSTx) setTenantContext(ctx context.Context) error {
	if tx.tenantSet {
		return nil
	}

	tenantID := tenant.FromContext(ctx)
	// SET LOCAL scopes to the current transaction
	_, err := tx.Tx.ExecContext(ctx, "SET LOCAL app.current_tenant = $1", tenantID)
	if err != nil {
		return err
	}
	tx.tenantSet = true
	return nil
}

// Exec executes a query with tenant context.
func (tx *RLSTx) Exec(ctx context.Context, query string, args, v any) error {
	if err := tx.setTenantContext(ctx); err != nil {
		return err
	}
	return tx.Tx.Exec(ctx, query, args, v)
}

// Query executes a query with tenant context.
func (tx *RLSTx) Query(ctx context.Context, query string, args, v any) error {
	if err := tx.setTenantContext(ctx); err != nil {
		return err
	}
	return tx.Tx.Query(ctx, query, args, v)
}
