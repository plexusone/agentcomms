package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandEnvVars(t *testing.T) {
	// Set test env vars
	os.Setenv("TEST_VAR", "test_value")
	os.Setenv("ANOTHER_VAR", "another_value")
	defer os.Unsetenv("TEST_VAR")
	defer os.Unsetenv("ANOTHER_VAR")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no vars",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "single ${VAR}",
			input:    "value is ${TEST_VAR}",
			expected: "value is test_value",
		},
		{
			name:     "single $VAR",
			input:    "value is $TEST_VAR",
			expected: "value is test_value",
		},
		{
			name:     "multiple vars",
			input:    "${TEST_VAR} and ${ANOTHER_VAR}",
			expected: "test_value and another_value",
		},
		{
			name:     "undefined var",
			input:    "${UNDEFINED_VAR}",
			expected: "",
		},
		{
			name:     "mixed vars",
			input:    "${TEST_VAR} $ANOTHER_VAR ${UNDEFINED_VAR}",
			expected: "test_value another_value ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("expandEnvVars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoadUnifiedConfig(t *testing.T) {
	// Set env vars for substitution
	os.Setenv("TEST_TOKEN", "secret_token")
	defer os.Unsetenv("TEST_TOKEN")

	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configContent := `{
		"version": "1",
		"server": {
			"port": 4444
		},
		"agents": [
			{
				"id": "test-agent",
				"type": "tmux",
				"tmux_session": "test-session"
			}
		],
		"chat": {
			"discord": {
				"enabled": true,
				"token": "${TEST_TOKEN}"
			}
		}
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := LoadUnifiedConfig(configPath)
	if err != nil {
		t.Fatalf("LoadUnifiedConfig failed: %v", err)
	}

	// Check values
	if cfg.Server.Port != 4444 {
		t.Errorf("Server.Port = %d, want 4444", cfg.Server.Port)
	}

	if len(cfg.Agents) != 1 {
		t.Fatalf("len(Agents) = %d, want 1", len(cfg.Agents))
	}
	if cfg.Agents[0].ID != "test-agent" {
		t.Errorf("Agents[0].ID = %q, want %q", cfg.Agents[0].ID, "test-agent")
	}

	// Check env var substitution
	if cfg.Chat.Discord.Token != "secret_token" {
		t.Errorf("Chat.Discord.Token = %q, want %q", cfg.Chat.Discord.Token, "secret_token")
	}
}

func TestUnifiedConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *UnifiedConfig
		wantError bool
		errMsg    string
	}{
		{
			name:      "empty config is valid",
			config:    DefaultUnifiedConfig(),
			wantError: false,
		},
		{
			name: "valid agent",
			config: &UnifiedConfig{
				Agents: []AgentConfig{
					{ID: "test", Type: "tmux", TmuxSession: "session"},
				},
			},
			wantError: false,
		},
		{
			name: "agent missing ID",
			config: &UnifiedConfig{
				Agents: []AgentConfig{
					{Type: "tmux", TmuxSession: "session"},
				},
			},
			wantError: true,
			errMsg:    "agent ID is required",
		},
		{
			name: "agent missing type",
			config: &UnifiedConfig{
				Agents: []AgentConfig{
					{ID: "test", TmuxSession: "session"},
				},
			},
			wantError: true,
			errMsg:    "type is required",
		},
		{
			name: "tmux agent missing session",
			config: &UnifiedConfig{
				Agents: []AgentConfig{
					{ID: "test", Type: "tmux"},
				},
			},
			wantError: true,
			errMsg:    "tmux_session is required",
		},
		{
			name: "duplicate agent ID",
			config: &UnifiedConfig{
				Agents: []AgentConfig{
					{ID: "test", Type: "tmux", TmuxSession: "s1"},
					{ID: "test", Type: "tmux", TmuxSession: "s2"},
				},
			},
			wantError: true,
			errMsg:    "duplicate agent ID",
		},
		{
			name: "discord enabled without token",
			config: &UnifiedConfig{
				Chat: &ChatConfig{
					Discord: &DiscordConfig{Enabled: true},
				},
			},
			wantError: true,
			errMsg:    "chat.discord.token is required when enabled",
		},
		{
			name: "discord disabled without token is ok",
			config: &UnifiedConfig{
				Chat: &ChatConfig{
					Discord: &DiscordConfig{Enabled: false},
				},
			},
			wantError: false,
		},
		{
			name: "channel references unknown agent",
			config: &UnifiedConfig{
				Agents: []AgentConfig{
					{ID: "agent1", Type: "tmux", TmuxSession: "s1"},
				},
				Chat: &ChatConfig{
					Channels: []ChannelMapping{
						{ChannelID: "discord:123", AgentID: "unknown"},
					},
				},
			},
			wantError: true,
			errMsg:    "unknown agent_id",
		},
		{
			name: "invalid TTS provider",
			config: &UnifiedConfig{
				Voice: &VoiceConfig{
					Phone: PhoneConfig{
						AccountSID: "sid",
						AuthToken:  "token",
						Number:     "+1234",
						UserNumber: "+5678",
					},
					TTS: TTSConfig{
						Provider: "invalid",
						APIKey:   "key",
					},
					STT: STTConfig{
						APIKey: "key",
					},
					Ngrok: NgrokConfig{
						AuthToken: "token",
					},
				},
			},
			wantError: true,
			errMsg:    "invalid TTS provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				if err == nil {
					t.Error("Validate() returned nil, want error")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() returned error: %v", err)
				}
			}
		})
	}
}

func TestUnifiedConfigToLegacyConfig(t *testing.T) {
	cfg := &UnifiedConfig{
		Server: ServerConfig{Port: 4444},
		Voice: &VoiceConfig{
			Phone: PhoneConfig{
				Provider:   "twilio",
				AccountSID: "sid123",
				AuthToken:  "token456",
				Number:     "+15551234567",
				UserNumber: "+15559876543",
			},
			TTS: TTSConfig{
				Provider: "elevenlabs",
				APIKey:   "eleven_key",
				Voice:    "Rachel",
				Model:    "eleven_turbo_v2_5",
			},
			STT: STTConfig{
				Provider: "deepgram",
				APIKey:   "dg_key",
				Model:    "nova-2",
				Language: "en-US",
			},
			Ngrok: NgrokConfig{
				AuthToken: "ngrok_token",
			},
		},
		Chat: &ChatConfig{
			Discord: &DiscordConfig{
				Enabled: true,
				Token:   "discord_token",
				GuildID: "guild123",
			},
		},
	}

	legacy := cfg.ToLegacyConfig()

	// Check server
	if legacy.Port != 4444 {
		t.Errorf("Port = %d, want 4444", legacy.Port)
	}

	// Check phone
	if legacy.PhoneProvider != "twilio" {
		t.Errorf("PhoneProvider = %q, want twilio", legacy.PhoneProvider)
	}
	if legacy.PhoneAccountSID != "sid123" {
		t.Errorf("PhoneAccountSID = %q, want sid123", legacy.PhoneAccountSID)
	}

	// Check TTS
	if legacy.TTSProvider != "elevenlabs" {
		t.Errorf("TTSProvider = %q, want elevenlabs", legacy.TTSProvider)
	}
	if legacy.ElevenLabsAPIKey != "eleven_key" {
		t.Errorf("ElevenLabsAPIKey = %q, want eleven_key", legacy.ElevenLabsAPIKey)
	}

	// Check STT
	if legacy.STTProvider != "deepgram" {
		t.Errorf("STTProvider = %q, want deepgram", legacy.STTProvider)
	}
	if legacy.DeepgramAPIKey != "dg_key" {
		t.Errorf("DeepgramAPIKey = %q, want dg_key", legacy.DeepgramAPIKey)
	}

	// Check Discord
	if !legacy.DiscordEnabled {
		t.Error("DiscordEnabled = false, want true")
	}
	if legacy.DiscordToken != "discord_token" {
		t.Errorf("DiscordToken = %q, want discord_token", legacy.DiscordToken)
	}
}

func TestUnifiedConfigFindAgentByChannel(t *testing.T) {
	cfg := &UnifiedConfig{
		Agents: []AgentConfig{
			{ID: "agent1", Type: "tmux", TmuxSession: "s1"},
			{ID: "agent2", Type: "tmux", TmuxSession: "s2"},
		},
		Chat: &ChatConfig{
			Channels: []ChannelMapping{
				{ChannelID: "discord:123", AgentID: "agent1"},
				{ChannelID: "telegram:456", AgentID: "agent2"},
			},
		},
	}

	tests := []struct {
		channelID string
		wantAgent string
		wantFound bool
	}{
		{"discord:123", "agent1", true},
		{"telegram:456", "agent2", true},
		{"discord:999", "", false},
		{"whatsapp:123", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.channelID, func(t *testing.T) {
			agent, found := cfg.FindAgentByChannel(tt.channelID)
			if found != tt.wantFound {
				t.Errorf("FindAgentByChannel(%q) found = %v, want %v", tt.channelID, found, tt.wantFound)
			}
			if agent != tt.wantAgent {
				t.Errorf("FindAgentByChannel(%q) = %q, want %q", tt.channelID, agent, tt.wantAgent)
			}
		})
	}
}

func TestUnifiedConfigGetAgent(t *testing.T) {
	cfg := &UnifiedConfig{
		Agents: []AgentConfig{
			{ID: "agent1", Type: "tmux", TmuxSession: "s1"},
			{ID: "agent2", Type: "tmux", TmuxSession: "s2"},
		},
	}

	agent, found := cfg.GetAgent("agent1")
	if !found {
		t.Error("GetAgent(agent1) not found")
	}
	if agent.TmuxSession != "s1" {
		t.Errorf("agent.TmuxSession = %q, want s1", agent.TmuxSession)
	}

	_, found = cfg.GetAgent("unknown")
	if found {
		t.Error("GetAgent(unknown) should not be found")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
