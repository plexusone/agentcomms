package tenant

import (
	"context"
	"testing"
)

func TestWithTenantID(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
		want     string
	}{
		{
			name:     "sets tenant ID",
			tenantID: "tenant-123",
			want:     "tenant-123",
		},
		{
			name:     "sets empty tenant ID",
			tenantID: "",
			want:     DefaultTenantID, // empty string should return default
		},
		{
			name:     "sets local tenant ID",
			tenantID: "local",
			want:     "local",
		},
		{
			name:     "sets UUID tenant ID",
			tenantID: "550e8400-e29b-41d4-a716-446655440000",
			want:     "550e8400-e29b-41d4-a716-446655440000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithTenantID(context.Background(), tt.tenantID)
			got := FromContext(ctx)
			if got != tt.want {
				t.Errorf("FromContext() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFromContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "returns default for empty context",
			ctx:  context.Background(),
			want: DefaultTenantID,
		},
		{
			name: "returns default for nil value",
			ctx:  context.WithValue(context.Background(), tenantKey, nil),
			want: DefaultTenantID,
		},
		{
			name: "returns default for wrong type",
			ctx:  context.WithValue(context.Background(), tenantKey, 123),
			want: DefaultTenantID,
		},
		{
			name: "returns tenant ID when set",
			ctx:  WithTenantID(context.Background(), "my-tenant"),
			want: "my-tenant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromContext(tt.ctx)
			if got != tt.want {
				t.Errorf("FromContext() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultTenantID(t *testing.T) {
	if DefaultTenantID != "local" {
		t.Errorf("DefaultTenantID = %q, want %q", DefaultTenantID, "local")
	}
}

func TestContextChaining(t *testing.T) {
	// Test that tenant context works with other context values
	type otherKey struct{}

	ctx := context.Background()
	ctx = context.WithValue(ctx, otherKey{}, "other-value")
	ctx = WithTenantID(ctx, "tenant-abc")

	// Verify tenant ID is accessible
	if got := FromContext(ctx); got != "tenant-abc" {
		t.Errorf("FromContext() = %q, want %q", got, "tenant-abc")
	}

	// Verify other context value is still accessible
	if got := ctx.Value(otherKey{}); got != "other-value" {
		t.Errorf("other value = %v, want %q", got, "other-value")
	}
}

func TestContextOverwrite(t *testing.T) {
	// Test that setting tenant ID twice uses the latest value
	ctx := context.Background()
	ctx = WithTenantID(ctx, "first-tenant")
	ctx = WithTenantID(ctx, "second-tenant")

	if got := FromContext(ctx); got != "second-tenant" {
		t.Errorf("FromContext() = %q, want %q", got, "second-tenant")
	}
}
