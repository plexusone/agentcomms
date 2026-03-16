package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/plexusone/agentcomms/internal/daemon"
)

var (
	// Flags
	flagLimit   int
	flagReason  string
	flagAgentID string
)

func init() {
	// Add commands
	rootCmd.AddCommand(sendCmd)
	rootCmd.AddCommand(interruptCmd)
	rootCmd.AddCommand(agentsCmd)
	rootCmd.AddCommand(eventsCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(replyCmd)
	rootCmd.AddCommand(channelsCmd)
	rootCmd.AddCommand(configCmd)

	// Config subcommands
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configShowCmd)

	// Events flags
	eventsCmd.Flags().IntVarP(&flagLimit, "limit", "n", 20, "Number of events to show")

	// Interrupt flags
	interruptCmd.Flags().StringVarP(&flagReason, "reason", "r", "", "Reason for interrupt")

	// Reply flags
	replyCmd.Flags().StringVarP(&flagAgentID, "agent", "a", "", "Agent ID (optional, for event tracking)")
}

var sendCmd = &cobra.Command{
	Use:   "send <agent-id> <message>",
	Short: "Send a message to an agent",
	Long:  `Sends a message to the specified agent via the daemon.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID := args[0]
		message := args[1]
		return runSend(agentID, message)
	},
}

var interruptCmd = &cobra.Command{
	Use:   "interrupt <agent-id>",
	Short: "Send an interrupt signal to an agent",
	Long:  `Sends Ctrl-C to the specified agent to interrupt the current operation.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID := args[0]
		return runInterrupt(agentID, flagReason)
	},
}

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "List registered agents",
	Long:  `Lists all agents registered with the daemon.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAgents()
	},
}

var eventsCmd = &cobra.Command{
	Use:   "events <agent-id>",
	Short: "List recent events for an agent",
	Long:  `Lists recent events (messages, interrupts) for the specified agent.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID := args[0]
		return runEvents(agentID, flagLimit)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status",
	Long:  `Shows the status of the agentcomms daemon.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus()
	},
}

var replyCmd = &cobra.Command{
	Use:   "reply <channel-id> <message>",
	Short: "Send a reply to a chat channel",
	Long: `Sends a message from an agent to a chat channel (Discord, Telegram, WhatsApp).

Channel ID format: "provider:chatid"
Examples:
  - discord:123456789012345678
  - telegram:987654321
  - whatsapp:1234567890@s.whatsapp.net`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		channelID := args[0]
		message := args[1]
		return runReply(channelID, message, flagAgentID)
	},
}

var channelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "List mapped chat channels",
	Long:  `Lists all chat channels mapped to agents in the configuration.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChannels()
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long:  `Commands for managing and validating the daemon configuration.`,
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration file",
	Long: `Validates the daemon configuration file and checks:
  - YAML syntax
  - Required fields
  - Agent configuration (type, tmux session)
  - Chat provider configuration
  - Channel mappings
  - Tmux session existence (if tmux agents configured)`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigValidate()
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current configuration",
	Long:  `Displays the current daemon configuration loaded from the config file.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigShow()
	},
}

// runSend sends a message to an agent.
func runSend(agentID, message string) error {
	client := daemon.DefaultClient()

	if err := client.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	result, err := client.Send(ctx, agentID, message)
	if err != nil {
		return err
	}

	fmt.Printf("Message sent (event: %s)\n", result.EventID)
	return nil
}

// runInterrupt sends an interrupt to an agent.
func runInterrupt(agentID, reason string) error {
	client := daemon.DefaultClient()

	if err := client.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	result, err := client.Interrupt(ctx, agentID, reason)
	if err != nil {
		return err
	}

	fmt.Printf("Interrupt sent (event: %s)\n", result.EventID)
	return nil
}

// runAgents lists registered agents.
func runAgents() error {
	client := daemon.DefaultClient()

	if err := client.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	result, err := client.Agents(ctx)
	if err != nil {
		return err
	}

	if len(result.Agents) == 0 {
		fmt.Println("No agents registered")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTYPE\tTARGET")
	for _, agent := range result.Agents {
		fmt.Fprintf(w, "%s\t%s\t%s\n", agent.ID, agent.Type, agent.Target)
	}
	w.Flush()

	return nil
}

// runEvents lists recent events for an agent.
func runEvents(agentID string, limit int) error {
	client := daemon.DefaultClient()

	if err := client.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	result, err := client.Events(ctx, agentID, limit)
	if err != nil {
		return err
	}

	if len(result.Events) == 0 {
		fmt.Printf("No events for agent %s\n", agentID)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIME\tTYPE\tROLE\tSTATUS\tCHANNEL")
	for _, evt := range result.Events {
		ts := evt.Timestamp.Format(time.RFC3339)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			ts, evt.Type, evt.Role, evt.Status, evt.ChannelID)
	}
	w.Flush()

	return nil
}

// runStatus shows daemon status.
func runStatus() error {
	client := daemon.DefaultClient()

	if err := client.Connect(); err != nil {
		fmt.Println("Daemon status: stopped")
		return nil
	}
	defer client.Close()

	ctx := context.Background()
	result, err := client.Status(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Daemon status: running")
	fmt.Printf("  Started: %s\n", result.StartedAt.Format(time.RFC3339))
	fmt.Printf("  Agents:  %d\n", result.Agents)
	if len(result.Providers) > 0 {
		fmt.Printf("  Providers: %v\n", result.Providers)
	}

	return nil
}

// runReply sends a message to a chat channel.
func runReply(channelID, message, agentID string) error {
	client := daemon.DefaultClient()

	if err := client.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	result, err := client.Reply(ctx, channelID, message, agentID)
	if err != nil {
		return err
	}

	fmt.Printf("Reply sent (event: %s)\n", result.EventID)
	return nil
}

// runChannels lists mapped chat channels.
func runChannels() error {
	client := daemon.DefaultClient()

	if err := client.Connect(); err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	result, err := client.Channels(ctx)
	if err != nil {
		return err
	}

	if len(result.Channels) == 0 {
		fmt.Println("No channels configured")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CHANNEL\tPROVIDER\tAGENT")
	for _, ch := range result.Channels {
		fmt.Fprintf(w, "%s\t%s\t%s\n", ch.ChannelID, ch.Provider, ch.AgentID)
	}
	w.Flush()

	return nil
}

// runConfigValidate validates the configuration file.
func runConfigValidate() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	dataDir := filepath.Join(homeDir, ".agentcomms")
	configPath := filepath.Join(dataDir, "config.yaml")

	fmt.Printf("Validating configuration: %s\n\n", configPath)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("Status: No configuration file found")
		fmt.Println("\nTo create a configuration file, copy the example:")
		fmt.Printf("  mkdir -p %s\n", dataDir)
		fmt.Printf("  cp examples/config.yaml %s\n", configPath)
		return nil
	}

	// Load configuration
	cfg, err := daemon.LoadDaemonConfig(dataDir)
	if err != nil {
		fmt.Println("Status: INVALID")
		fmt.Printf("\nError: %v\n", err)
		return fmt.Errorf("configuration error: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Println("Status: INVALID")
		fmt.Printf("\nValidation error: %v\n", err)
		return fmt.Errorf("validation error: %w", err)
	}

	var warnings []string

	// Check agents
	fmt.Printf("Agents: %d configured\n", len(cfg.Agents))
	for _, agent := range cfg.Agents {
		fmt.Printf("  - %s (type: %s)\n", agent.ID, agent.Type)

		// Check tmux session exists for tmux agents
		if agent.Type == "tmux" {
			if !checkTmuxSession(agent.TmuxSession) {
				warnings = append(warnings,
					fmt.Sprintf("tmux session '%s' for agent '%s' does not exist",
						agent.TmuxSession, agent.ID))
			}
		}
	}

	// Check chat providers
	if cfg.Chat != nil {
		var providers []string
		if cfg.Chat.Discord != nil {
			providers = append(providers, "discord")
			if len(cfg.Chat.Discord.Token) < 50 {
				warnings = append(warnings, "discord token appears too short")
			}
		}
		if cfg.Chat.Telegram != nil {
			providers = append(providers, "telegram")
		}
		if cfg.Chat.WhatsApp != nil {
			providers = append(providers, "whatsapp")
			// Check if WhatsApp DB path is writable
			dbDir := filepath.Dir(cfg.Chat.WhatsApp.DBPath)
			if _, err := os.Stat(dbDir); os.IsNotExist(err) {
				warnings = append(warnings,
					fmt.Sprintf("whatsapp db directory does not exist: %s", dbDir))
			}
		}

		fmt.Printf("\nChat providers: %v\n", providers)
		fmt.Printf("Channel mappings: %d\n", len(cfg.Chat.Channels))
		for _, mapping := range cfg.Chat.Channels {
			fmt.Printf("  - %s -> %s\n", mapping.ChannelID, mapping.AgentID)
		}
	} else {
		fmt.Println("\nChat: not configured")
	}

	// Print warnings
	if len(warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, w := range warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	fmt.Println("\nStatus: VALID")
	return nil
}

// runConfigShow displays the current configuration.
func runConfigShow() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	dataDir := filepath.Join(homeDir, ".agentcomms")
	configPath := filepath.Join(dataDir, "config.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("No configuration file found at: %s\n", configPath)
		return nil
	}

	// Read and print the file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	fmt.Printf("# Configuration file: %s\n\n", configPath)
	fmt.Println(string(data))
	return nil
}

// checkTmuxSession checks if a tmux session exists.
func checkTmuxSession(session string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", session) //nolint:gosec
	return cmd.Run() == nil
}
