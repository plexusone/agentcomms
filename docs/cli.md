# CLI Commands

AgentComms provides a command-line interface for managing the daemon and interacting with agents.

## Main Commands

### serve

Run the MCP server for outbound communication (AI → Human).

```bash
agentcomms serve
```

This starts the MCP server that AI assistants connect to for voice calls and chat messaging.

### daemon

Run the background daemon for inbound communication (Human → AI).

```bash
agentcomms daemon
```

The daemon:

- Connects to Discord/Telegram/WhatsApp to receive messages
- Routes messages to agents via tmux
- Stores all events in SQLite
- Exposes a Unix socket API for CLI commands

## Daemon Interaction Commands

These commands interact with a running daemon.

### status

Show daemon status.

```bash
agentcomms status
```

Output:

```
Daemon status: running
  Started: 2024-01-15T09:00:00Z
  Agents:  1
  Providers: [discord telegram]
```

### agents

List registered agents.

```bash
agentcomms agents
```

Output:

```
ID       TYPE   TARGET
claude   tmux   tmux:claude-code:0
```

### send

Send a message to an agent.

```bash
agentcomms send <agent-id> <message>
```

Example:

```bash
agentcomms send claude "Can you also add unit tests?"
```

The message appears in the agent's tmux pane.

### interrupt

Send an interrupt signal (Ctrl-C) to an agent.

```bash
agentcomms interrupt <agent-id> [--reason <reason>]
```

Example:

```bash
agentcomms interrupt claude --reason "Need to change approach"
```

### events

List recent events for an agent.

```bash
agentcomms events <agent-id> [--limit <n>]
```

Example:

```bash
agentcomms events claude --limit 10
```

Output:

```
TIME                          TYPE            ROLE    STATUS     CHANNEL
2024-01-15T10:30:00Z          human_message   human   delivered  discord:123456
2024-01-15T10:25:00Z          agent_message   agent   delivered  discord:123456
```

### reply

Send a message to a chat channel (outbound from agent).

```bash
agentcomms reply <channel-id> <message> [--agent <agent-id>]
```

Example:

```bash
agentcomms reply discord:123456789 "Task completed!" --agent claude
```

Channel ID format: `provider:chatid`

- `discord:123456789012345678`
- `telegram:987654321`
- `whatsapp:1234567890@s.whatsapp.net`

### channels

List mapped chat channels.

```bash
agentcomms channels
```

Output:

```
CHANNEL                   PROVIDER   AGENT
discord:123456789         discord    claude
telegram:987654321        telegram   claude
```

## Configuration Commands

### config init

Generate a new configuration file.

```bash
# Full config with voice settings
agentcomms config init

# Minimal config (chat only)
agentcomms config init --minimal

# Generate to specific path
agentcomms config init -o /path/to/config.json
```

Output:

```
Configuration file created: /Users/you/.agentcomms/config.json

Next steps:
  1. Edit the configuration file to set your values
  2. Set environment variables for secrets (DISCORD_TOKEN, etc.)
  3. Validate with: agentcomms config validate
  4. Start the daemon: agentcomms daemon
```

### config validate

Validate the configuration file.

```bash
agentcomms config validate
```

Checks:

- JSON syntax
- Required fields
- Agent configuration
- Voice provider configuration
- Tmux session existence
- Chat provider configuration
- Channel mappings

Output:

```
Validating configuration: /Users/you/.agentcomms/config.json

Server port: 3333
Agents: 1 configured
  - claude (type: tmux)

Voice: enabled
  Phone provider: twilio
  TTS provider: elevenlabs
  STT provider: deepgram

Chat providers: [discord]
Channel mappings: 1
  - discord:123456789 -> claude

Status: VALID
```

### config show

Display the current configuration file.

```bash
agentcomms config show
```

## Data Storage

The daemon stores data in `~/.agentcomms/`:

| File | Description |
|------|-------------|
| `config.json` | Configuration file (unified JSON) |
| `data.db` | SQLite database (events, agents) |
| `daemon.sock` | Unix socket for CLI/API |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (see error message) |
