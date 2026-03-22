package database

import (
	"strings"
	"testing"
)

func TestRLSPolicySQL(t *testing.T) {
	// Test that RLS policy SQL constants are properly defined
	tests := []struct {
		name     string
		sql      string
		contains []string
	}{
		{
			name: "EnableRLSAgents",
			sql:  EnableRLSAgents,
			contains: []string{
				"ALTER TABLE",
				"agents",
				"ENABLE ROW LEVEL SECURITY",
			},
		},
		{
			name: "CreatePolicyAgents",
			sql:  CreatePolicyAgents,
			contains: []string{
				"CREATE POLICY",
				"tenant_isolation_agents",
				"agents",
				"tenant_id",
				"current_setting",
				"app.current_tenant",
			},
		},
		{
			name: "EnableRLSEvents",
			sql:  EnableRLSEvents,
			contains: []string{
				"ALTER TABLE",
				"events",
				"ENABLE ROW LEVEL SECURITY",
			},
		},
		{
			name: "CreatePolicyEvents",
			sql:  CreatePolicyEvents,
			contains: []string{
				"CREATE POLICY",
				"tenant_isolation_events",
				"events",
				"tenant_id",
				"current_setting",
				"app.current_tenant",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, substr := range tt.contains {
				if !strings.Contains(tt.sql, substr) {
					t.Errorf("%s SQL should contain %q, got:\n%s", tt.name, substr, tt.sql)
				}
			}
		})
	}
}

func TestRLSPolicyUsesWithCheck(t *testing.T) {
	// Verify that CREATE POLICY includes WITH CHECK for insert/update operations
	policies := []string{CreatePolicyAgents, CreatePolicyEvents}

	for _, policy := range policies {
		if !strings.Contains(policy, "WITH CHECK") {
			t.Errorf("Policy should include WITH CHECK clause:\n%s", policy)
		}
	}
}

func TestRLSPolicyUsesCurrentSetting(t *testing.T) {
	// Verify that policies use current_setting with true for missing_ok
	// This prevents errors when the session variable is not set
	policies := []string{CreatePolicyAgents, CreatePolicyEvents}

	for _, policy := range policies {
		if !strings.Contains(policy, "current_setting('app.current_tenant', true)") {
			t.Errorf("Policy should use current_setting with missing_ok=true:\n%s", policy)
		}
	}
}
