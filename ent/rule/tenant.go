// Package rule provides Ent privacy rules for AgentComms.
package rule

import (
	"context"

	"entgo.io/ent/entql"

	"github.com/plexusone/agentcomms/ent/privacy"
	"github.com/plexusone/agentcomms/internal/tenant"
)

// TenantFilter is an interface for filtering by tenant_id.
// Both AgentQuery and EventQuery implement this through their Filter() method.
type TenantFilter interface {
	WhereTenantID(p entql.StringP)
}

// FilterTenantRule returns a privacy rule that filters queries and mutations by tenant_id.
// It extracts the tenant ID from context and adds a WHERE clause to filter by tenant_id.
// This rule works on both SQLite and PostgreSQL.
func FilterTenantRule() privacy.QueryMutationRule {
	return privacy.FilterFunc(func(ctx context.Context, f privacy.Filter) error {
		tenantID := tenant.FromContext(ctx)

		// Type assert to get the tenant-specific filter methods
		tf, ok := f.(TenantFilter)
		if !ok {
			return privacy.Denyf("unexpected filter type %T", f)
		}

		// Add WHERE tenant_id = ? clause
		tf.WhereTenantID(entql.StringEQ(tenantID))

		// Continue to next privacy rule
		return privacy.Skip
	})
}
