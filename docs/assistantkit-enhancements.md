# Proposed assistantkit Enhancements

Enhancements to make tool-specific code unnecessary in projects like agentcomms.

## 1. Unified Plugin Bundle Type

Create a `Bundle` type that contains all components:

```go
// assistantkit/bundle/bundle.go
package bundle

type Bundle struct {
    Plugin   *plugins.Plugin
    Skills   []*skills.Skill
    Commands []*commands.Command
    Hooks    *hooks.Config
    Agents   []*agents.Agent
    Context  *context.Context
}

func New(name, version, description string) *Bundle
func (b *Bundle) AddSkill(skill *skills.Skill)
func (b *Bundle) AddCommand(cmd *commands.Command)
func (b *Bundle) SetHooks(cfg *hooks.Config)
func (b *Bundle) AddAgent(agent *agents.Agent)
```

## 2. Single Generate Function

One function to generate for any tool:

```go
// assistantkit/bundle/generate.go
package bundle

// Generate outputs the bundle for a specific tool
func (b *Bundle) Generate(tool, outputDir string) error

// GenerateAll outputs for all supported tools
func (b *Bundle) GenerateAll(outputDir string) error

// Example usage:
bundle := bundle.New("agentcomms", "0.2.0", "Voice calls and chat messaging for AI assistants")
bundle.Plugin.Author = "plexusone"
bundle.AddSkill(phoneSkill)
bundle.AddCommand(callCommand)
bundle.SetHooks(hooksConfig)

// Generate for Claude - no Claude-specific code needed!
bundle.Generate("claude", ".")

// Or generate for all tools
bundle.GenerateAll("./plugins")
```

## 3. MCP Server in Plugin Definition

Ensure MCP servers are properly written to manifests:

```go
// Currently in agentcomms:
plugin.AddMCPServer("agentcomms", plugins.MCPServer{
    Command: "./agentcomms",
    Args:    []string{},
    Env:     map[string]string{...},
})

// Enhancement: Verify this writes correctly to:
// - Claude: .claude-plugin/plugin.json (mcpServers field)
// - VS Code: .vscode/settings.json (mcp field)
// - Cursor: .cursor/mcp.json
```

## 4. Tool-Agnostic Path Resolution

assistantkit should handle paths internally:

```go
// Instead of:
pluginDir := filepath.Join(outputDir, ".claude-plugin")

// assistantkit should resolve:
adapter.OutputDir()  // Returns tool-specific directory
adapter.PluginPath() // Returns full path for plugin manifest
adapter.SkillsDir()  // Returns skills directory
```

## 5. Declarative Bundle Definition

Support YAML/JSON bundle definition:

```yaml
# bundle.yaml
name: agentcomms
version: 0.2.0
description: Voice calls and chat messaging for AI assistants
author: plexusone

mcp_servers:
  agentcomms:
    command: ./agentcomms
    env:
      NGROK_AUTHTOKEN: ${NGROK_AUTHTOKEN}

skills:
  - path: skills/phone-input.yaml
  - path: skills/chat-messaging.yaml

commands:
  - path: commands/call.yaml
  - path: commands/message.yaml

hooks:
  on_stop:
    - type: prompt
      prompt: "Consider calling or messaging the user..."
```

Then generate with CLI:

```bash
assistantkit generate --bundle bundle.yaml --tool claude --output .
assistantkit generate --bundle bundle.yaml --all --output ./plugins
```

## 6. Tool Detection

Auto-detect which tools are configured in current project:

```go
// assistantkit/detect/detect.go
package detect

func DetectedTools(dir string) []string
// Returns: ["claude", "cursor", "vscode"] based on config files found

func PrimaryTool(dir string) string
// Returns most likely primary tool based on config
```

## 7. Default Tool Configuration

Allow setting a default tool:

```go
assistantkit.SetDefaultTool("claude")

// Then no tool specification needed:
bundle.Generate(".")  // Uses default tool
```

## 8. Template-Based Generation

For complex skills/commands, support templates:

```go
skill := skills.NewSkillFromTemplate("phone-input", &SkillTemplateData{
    Tools: []ToolDoc{
        {Name: "initiate_call", Description: "...", Example: "..."},
        {Name: "continue_call", Description: "...", Example: "..."},
    },
    BestPractices: []string{"Be conversational", "Be concise"},
})
```

## Implementation Priority

| Enhancement | Priority | Impact |
|-------------|----------|--------|
| Unified Bundle Type | High | Eliminates multi-package imports |
| Single Generate Function | High | Eliminates adapter selection code |
| MCP Server Support | High | Critical for MCP plugins |
| Tool-Agnostic Paths | Medium | Eliminates path knowledge |
| Declarative Definition | Medium | Enables CLI-based generation |
| Tool Detection | Low | Nice-to-have |
| Templates | Low | Reduces boilerplate |

## Resulting agentcomms Code

With these enhancements, agentcomms's generator becomes:

```go
package main

import (
    "log"
    "github.com/plexusone/assistantkit/bundle"
)

func main() {
    b := bundle.New("agentcomms", "0.2.0", "Voice calls and chat messaging for AI assistants")
    b.Plugin.Author = "plexusone"
    b.Plugin.Repository = "https://github.com/plexusone/agentcomms"

    b.Plugin.AddMCPServer("agentcomms", bundle.MCPServer{
        Command: "./agentcomms",
        Env: map[string]string{
            "NGROK_AUTHTOKEN": "${NGROK_AUTHTOKEN}",
        },
    })

    b.AddSkill(createPhoneSkill())
    b.AddSkill(createChatSkill())
    b.AddCommand(createCallCommand())
    b.AddCommand(createMessageCommand())
    b.SetHooks(createHooks())

    // No Claude-specific code!
    if err := b.Generate("claude", "."); err != nil {
        log.Fatal(err)
    }
}
```

Or even simpler with declarative:

```bash
assistantkit generate --bundle agentcomms.yaml --tool claude
```
