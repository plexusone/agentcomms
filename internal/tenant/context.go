// Package tenant provides tenant context management for multi-tenancy.
package tenant

import "context"

// contextKey is a custom type for context keys to avoid collisions.
type contextKey int

const tenantKey contextKey = iota

// DefaultTenantID is the default tenant ID for single-tenant mode.
const DefaultTenantID = "local"

// WithTenantID returns a new context with the tenant ID set.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantKey, tenantID)
}

// FromContext returns the tenant ID from the context.
// Returns DefaultTenantID ("local") if not set.
func FromContext(ctx context.Context) string {
	if v := ctx.Value(tenantKey); v != nil {
		if id, ok := v.(string); ok && id != "" {
			return id
		}
	}
	return DefaultTenantID
}
