# IRC Setup Guide

This guide covers setting up IRC as a bidirectional chat transport for AgentComms.

## Overview

IRC (Internet Relay Chat) is a text-based chat protocol that has been around since 1988. It's still widely used in open source communities, developer groups, and tech organizations.

AgentComms connects to IRC servers as a client, allowing your AI agent to:

- Receive messages from IRC channels and DMs
- Send responses back to channels or users
- Join multiple channels simultaneously
- Use TLS for secure connections

## Quick Start

### 1. Choose an IRC Server

Common public IRC networks:

| Network | Server | Port (TLS) | Website |
|---------|--------|------------|---------|
| Libera.Chat | `irc.libera.chat` | 6697 | [libera.chat](https://libera.chat) |
| OFTC | `irc.oftc.net` | 6697 | [oftc.net](https://www.oftc.net) |
| EFnet | `irc.efnet.org` | 6697 | [efnet.org](http://www.efnet.org) |
| IRCnet | `open.ircnet.net` | 6697 | [ircnet.org](http://www.ircnet.org) |

You can also run your own IRC server using software like [InspIRCd](https://www.inspircd.org/) or [ngircd](https://ngircd.barton.de/).

### 2. Configure AgentComms

Add IRC configuration to your `~/.agentcomms/config.json`:

```json
{
  "chat": {
    "irc": {
      "enabled": true,
      "server": "irc.libera.chat:6697",
      "nick": "myagentbot",
      "channels": ["#mychannel", "#anotherchannel"],
      "use_tls": true
    }
  }
}
```

### 3. Environment Variables

You can also configure IRC using environment variables:

```bash
export AGENTCOMMS_IRC_ENABLED=true
export AGENTCOMMS_IRC_SERVER=irc.libera.chat:6697
export AGENTCOMMS_IRC_NICK=myagentbot
export AGENTCOMMS_IRC_PASSWORD=your_nickserv_password
export AGENTCOMMS_IRC_CHANNELS="#channel1,#channel2"
export AGENTCOMMS_IRC_USE_TLS=true
```

### 4. Start the Daemon

```bash
agentcomms daemon
```

You should see log output confirming the IRC connection:

```
IRC provider registered server=irc.libera.chat:6697
connected to IRC server server=irc.libera.chat:6697 nick=myagentbot
joining channel channel=#mychannel
```

## NickServ Authentication

Many IRC networks support NickServ for nickname registration and authentication. To use NickServ:

### 1. Register Your Nickname (Manual Step)

Connect to the IRC network manually using an IRC client and register:

```
/msg NickServ REGISTER yourpassword youremail@example.com
```

Follow the confirmation instructions from NickServ.

### 2. Configure the Password

Add the NickServ password to your config:

```json
{
  "chat": {
    "irc": {
      "enabled": true,
      "server": "irc.libera.chat:6697",
      "nick": "myagentbot",
      "password": "${IRC_PASSWORD}",
      "channels": ["#mychannel"],
      "use_tls": true
    }
  }
}
```

Set the password as an environment variable:

```bash
export IRC_PASSWORD=your_nickserv_password
```

The bot will automatically identify with NickServ on connect.

## Channel Configuration

### Joining Channels

Specify channels in the config:

```json
{
  "chat": {
    "irc": {
      "channels": ["#general", "#dev", "#support"]
    }
  }
}
```

Channels must start with `#`. The bot will join these channels automatically on connect.

### Channel Permissions

Some channels require the bot to be invited or have voice/op status to speak. Coordinate with channel operators if needed.

## TLS Configuration

TLS is enabled by default (`use_tls: true`). To disable TLS for unencrypted connections (not recommended):

```json
{
  "chat": {
    "irc": {
      "use_tls": false,
      "server": "irc.example.com:6667"
    }
  }
}
```

Standard ports:

- TLS: 6697
- Unencrypted: 6667

## Message Routing

### Chat IDs

IRC chat IDs follow this format:

- Channels: `irc:#channelname`
- Direct messages: `irc:nickname`

### Channel Mappings

Route IRC messages to specific agents:

```json
{
  "chat": {
    "irc": {
      "enabled": true,
      "server": "irc.libera.chat:6697",
      "nick": "agentbot",
      "channels": ["#dev"]
    },
    "channels": [
      {
        "channel_id": "irc:#dev",
        "agent_id": "claude"
      }
    ]
  }
}
```

## Sending Messages

Use the MCP tool to send messages to IRC:

```json
{
  "provider": "irc",
  "chat_id": "#dev",
  "message": "Hello from the agent!"
}
```

For direct messages:

```json
{
  "provider": "irc",
  "chat_id": "username",
  "message": "Hello via DM!"
}
```

## Self-Hosted IRC Server

For testing or private deployments, you can run a local IRC server.

### Using Docker (InspIRCd)

```bash
docker run -d -p 6667:6667 -p 6697:6697 inspircd/inspircd-docker
```

### Configure AgentComms for Local Server

```bash
export AGENTCOMMS_IRC_ENABLED=true
export AGENTCOMMS_IRC_SERVER=localhost:6667
export AGENTCOMMS_IRC_NICK=testbot
export AGENTCOMMS_IRC_CHANNELS="#test"
export AGENTCOMMS_IRC_USE_TLS=false
```

## Troubleshooting

### Connection Refused

- Verify the server address and port
- Check if TLS is required (most networks require TLS on port 6697)
- Ensure no firewall is blocking the connection

### Nickname Already in Use

- Choose a different nickname
- If you own the nickname, authenticate with NickServ

### Cannot Join Channel

- The channel may be invite-only (+i mode)
- The channel may require registration (+r mode)
- Contact channel operators for access

### Messages Not Received

- Verify the bot has joined the channel (`/who #channel`)
- Check channel modes that may restrict unvoiced users
- Review logs for any error messages

## Security Considerations

1. **Always use TLS** for production deployments
2. **Use environment variables** for passwords, never hardcode
3. **Register nicknames** on public networks to prevent impersonation
4. **Consider private servers** for sensitive communications

## IRC Protocol Notes

IRC has some limitations compared to modern chat platforms:

- No message editing or deletion
- No threading or replies
- Line length limits (~400 characters per message)
- No native file/media support
- No read receipts or typing indicators

AgentComms handles these limitations automatically, splitting long messages as needed.
