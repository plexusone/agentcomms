// Package sms provides an SMS chat provider using the omnichat interface.
package sms

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/plexusone/omnichat/provider"
	"github.com/plexusone/omnivoice-core/callsystem"
)

// Verify interface compliance at compile time.
var _ provider.Provider = (*Provider)(nil)

// Provider implements the omnichat Provider interface for SMS.
type Provider struct {
	smsProvider callsystem.SMSProvider
	phoneNumber string // Our phone number (from number)
	logger      *slog.Logger

	msgHandler   provider.MessageHandler
	eventHandler provider.EventHandler

	mu        sync.RWMutex
	connected bool
}

// Config configures the SMS provider.
type Config struct {
	// SMSProvider is the underlying SMS provider (from omnivoice).
	SMSProvider callsystem.SMSProvider

	// PhoneNumber is our outbound phone number (E.164 format).
	PhoneNumber string

	// Logger for provider logging.
	Logger *slog.Logger
}

// New creates a new SMS provider.
func New(cfg Config) (*Provider, error) {
	if cfg.SMSProvider == nil {
		return nil, fmt.Errorf("SMS provider is required")
	}
	if cfg.PhoneNumber == "" {
		return nil, fmt.Errorf("phone number is required")
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &Provider{
		smsProvider: cfg.SMSProvider,
		phoneNumber: cfg.PhoneNumber,
		logger:      cfg.Logger,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "sms"
}

// Connect establishes connection (no-op for SMS, always "connected").
func (p *Provider) Connect(ctx context.Context) error {
	p.mu.Lock()
	p.connected = true
	p.mu.Unlock()

	p.logger.Info("SMS provider connected", "phone_number", p.phoneNumber)
	return nil
}

// Disconnect closes the connection.
func (p *Provider) Disconnect(ctx context.Context) error {
	p.mu.Lock()
	p.connected = false
	p.mu.Unlock()

	p.logger.Info("SMS provider disconnected")
	return nil
}

// Send sends an SMS message.
func (p *Provider) Send(ctx context.Context, chatID string, msg provider.OutgoingMessage) error {
	p.mu.RLock()
	if !p.connected {
		p.mu.RUnlock()
		return fmt.Errorf("provider not connected")
	}
	p.mu.RUnlock()

	// chatID is the recipient phone number
	_, err := p.smsProvider.SendSMS(ctx, chatID, msg.Content)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	p.logger.Info("sent SMS",
		"to", chatID,
		"content_length", len(msg.Content),
	)

	return nil
}

// OnMessage registers a handler for incoming messages.
func (p *Provider) OnMessage(handler provider.MessageHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.msgHandler = handler
}

// OnEvent registers a handler for platform events.
func (p *Provider) OnEvent(handler provider.EventHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.eventHandler = handler
}

// HandleIncomingSMS processes an incoming SMS from the webhook server.
// This should be called by the webhook server when an SMS is received.
func (p *Provider) HandleIncomingSMS(ctx context.Context, from, to, body, messageID string) error {
	p.mu.RLock()
	handler := p.msgHandler
	connected := p.connected
	p.mu.RUnlock()

	if !connected {
		return fmt.Errorf("provider not connected")
	}

	if handler == nil {
		p.logger.Warn("no message handler registered, dropping SMS",
			"from", from,
			"message_id", messageID,
		)
		return nil
	}

	// Create incoming message
	msg := provider.IncomingMessage{
		ID:           messageID,
		ProviderName: "sms",
		ChatID:       from, // Use sender's phone as chat ID
		ChatType:     provider.ChatTypeDM,
		SenderID:     from,
		SenderName:   from, // Phone number as name
		Content:      body,
		Timestamp:    time.Now(),
		Metadata: map[string]any{
			"to":   to,
			"from": from,
		},
	}

	p.logger.Info("received SMS",
		"from", from,
		"to", to,
		"message_id", messageID,
	)

	return handler(ctx, msg)
}
