package database

import (
	"context"
	"database/sql"
	"fmt"
)

// RLS policy SQL statements for PostgreSQL.
const (
	// EnableRLSAgents enables RLS on the agents table.
	EnableRLSAgents = `ALTER TABLE agents ENABLE ROW LEVEL SECURITY`

	// CreatePolicyAgents creates a tenant isolation policy on agents.
	CreatePolicyAgents = `
CREATE POLICY IF NOT EXISTS tenant_isolation_agents ON agents
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true))`

	// EnableRLSEvents enables RLS on the events table.
	EnableRLSEvents = `ALTER TABLE events ENABLE ROW LEVEL SECURITY`

	// CreatePolicyEvents creates a tenant isolation policy on events.
	CreatePolicyEvents = `
CREATE POLICY IF NOT EXISTS tenant_isolation_events ON events
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true))`
)

// ApplyRLSPolicies applies Row-Level Security policies to PostgreSQL tables.
// This should be called after schema migration.
// The db parameter should be the underlying *sql.DB from the Ent client.
func ApplyRLSPolicies(ctx context.Context, db *sql.DB) error {
	policies := []struct {
		name string
		sql  string
	}{
		{"enable RLS on agents", EnableRLSAgents},
		{"create policy on agents", CreatePolicyAgents},
		{"enable RLS on events", EnableRLSEvents},
		{"create policy on events", CreatePolicyEvents},
	}

	for _, p := range policies {
		if _, err := db.ExecContext(ctx, p.sql); err != nil {
			return fmt.Errorf("failed to %s: %w", p.name, err)
		}
	}

	return nil
}

// DropRLSPolicies removes Row-Level Security policies from PostgreSQL tables.
// This can be useful for testing or migration rollback.
func DropRLSPolicies(ctx context.Context, db *sql.DB) error {
	statements := []string{
		"DROP POLICY IF EXISTS tenant_isolation_agents ON agents",
		"DROP POLICY IF EXISTS tenant_isolation_events ON events",
		"ALTER TABLE agents DISABLE ROW LEVEL SECURITY",
		"ALTER TABLE events DISABLE ROW LEVEL SECURITY",
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute: %s: %w", stmt, err)
		}
	}

	return nil
}
