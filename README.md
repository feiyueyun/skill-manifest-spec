# Skill Manifest Specification

**Skill Manifest (skill.json)** is the core standard file format for the FYY distributed AI workflow platform, describing AI Agent skill metadata, capability declarations, permission requirements, protocol configuration, industry interoperability, and commercialization information.

The first specification to unify **knowledge description + executable services + network interconnection** in a single skill descriptor.

## Scope

- JSON Schema v1 full definition (draft-07, with conditional dependencies)
- Go parsing library `pkg/manifest/` (Parse / Validate / Marshal / MarshalIndent)
- Property-based tests (round-trip consistency, crash resistance, idempotency — 10 properties)
- Industry standard compatibility mappings (Agent Skills / MCP / A2A / ClawHub / SkillNet)
- Complete example files for all three skill types

## Quick Start

### Install

```bash
go get github.com/feiyueyun/skill-manifest-spec/pkg/manifest
```

### Parse a skill.json

```go
package main

import (
	"fmt"
	"os"

	"github.com/feiyueyun/skill-manifest-spec/pkg/manifest"
)

func main() {
	data, _ := os.ReadFile("examples/weather-lookup.skill.json")

	m, err := manifest.Parse(data)
	if err != nil {
		fmt.Println("parse failed:", err)
		return
	}
	fmt.Printf("Skill: %s v%s (%s)\n", m.Name, m.Version, m.Type)
}
```

### Validate

```go
errs := manifest.Validate(m)
for _, e := range errs {
	fmt.Printf("[%s] %s: %s\n", e.Level, e.Field, e.Message)
}
```

`Validate` returns all validation errors (does not stop at the first error). Each `ValidationError` contains:
- `Field` — JSON field path (e.g. `mcp_remote.endpoint`)
- `Message` — error description
- `Level` — `"error"` or `"warn"`

### Serialize

```go
// Compact JSON
data, err := manifest.Marshal(m)

// Indented JSON
pretty, err := manifest.MarshalIndent(m, "  ")
```

## Schema

```
spec/skill-manifest-v1.schema.json
spec/categories.json              # Recommended category taxonomy (advisory, not enforced)
```

The schema uses JSON Schema draft-07 with `if/then` conditional dependencies (e.g. `type=service` requires `listen_port` and `protocol`).

Key field groups:
- Core: `name`, `version`, `description`, `type` (required)
- Service config: `listen_port`, `protocol`, `mcp`, `mcp_mode`, `mcp_remote`
- Permissions: `permissions`, `network_endpoints`, `risk_level`, `visibility`
- AI readability: `when_to_use`, `category` (20 values), `capabilities`, `input_examples`, `tags`
- Commercialization: `pricing` (6 models incl. `per_output`), `sla`, `resellable`
- Quality: `output_format`, `output_schema`, `quality_indicators` (verified, security_audit, quality_score, source_type)
- Interoperability: `interop`, `disclosure`, `ontology`, `instructions_path`
- Runtime: `schedule`, `memory_access`, `depends_on`, `health_check`, `resource_limits`
- Remote exposure: `remote` (enabled, visibility, rate_limit, billing — structured JSON over tsnet)
- System skill lifecycle: `system_skill`, `auto_install`

## Examples

| File | Skill Type | Highlights |
|------|-----------|------------|
| `examples/weather-lookup.skill.json` | `claw` | AI readability fields, permissions, i18n, interop, disclosure |
| `examples/code-review.skill.json` | `service` | listen_port, protocol, health_check, schedule, depends_on |
| `examples/database-query.skill.json` | `mcp` | MCP primitives (tools/resources/prompts), mcp_mode, ontology |

### Official Skills

| Directory | Skill Type | Highlights |
|-----------|-----------|------------|
| `official-skills/fyy-messenger/` | `claw` | system_skill, auto_install, remote disabled, owner escalation |

## Industry Compatibility

Skill Manifest is designed to interoperate with five major industry standards:

| Standard | Description | Integration |
|----------|-------------|-------------|
| **Agent Skills** (Anthropic) | SKILL.md portable skill format | `interop.agent_skills` + `disclosure` three-level progressive disclosure |
| **MCP** (Model Context Protocol) | Agent-tool integration protocol | `type=mcp` native support, `mcp` field declares tools/resources/prompts |
| **A2A** (Google Agent-to-Agent) | Agent communication protocol | `interop.a2a_card` maps to A2A Agent Card |
| **ClawHub** | OpenClaw skill marketplace ecosystem | `interop.clawhub` maps to claw.json format |
| **SkillNet** (ZJU/OpenKG) | 200,000+ skill knowledge graph | `ontology` field references three-level skill ontology |

## Remote Skill Exposure

The `remote` field controls whether and how a skill can be invoked by other devices over the tsnet mesh network. Remote invocation uses a **structured JSON protocol over tsnet** (WireGuard encrypted) — skills are NOT converted to MCP Servers. This preserves both layers of skill value:

- **Layer A (Reasoning)**: skill.md contains decision logic and domain knowledge, executed by Agent/LLM on the owner's device
- **Layer B (Tools)**: scripts/ contains deterministic functions, also executed on the owner's device

### `remote` field structure

```json
{
  "remote": {
    "enabled": true,
    "visibility": "private",
    "max_concurrency": 10,
    "timeout_s": 60,
    "rate_limit": {
      "max_per_minute": 60,
      "max_per_hour": 1000
    },
    "input_schema": {
      "type": "object",
      "properties": {
        "region": { "type": "string" },
        "industry": { "type": "string" }
      }
    },
    "output_formats": ["markdown", "json"],
    "billing": {
      "model": "per_call",
      "unit_price": 0.5,
      "currency": "credit"
    }
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | boolean | `false` | Whether this skill can be exposed for remote invocation |
| `visibility` | enum | `"private"` | Default visibility: `private` (grant holders only), `follows` (followers + grants), `public` (all) |
| `max_concurrency` | integer | `10` | Maximum concurrent remote invocations (1-1000) |
| `timeout_s` | integer | `60` | Invocation timeout in seconds (1-3600) |
| `rate_limit` | object | — | Per-caller rate limiting |
| `input_schema` | object | — | JSON Schema for structured input validation |
| `output_formats` | string[] | — | Supported output formats |
| `billing` | object | — | Billing model for remote callers |

A skill with `remote.enabled: false` (or no `remote` field) cannot be exposed via `fyy skill expose`.

Format conversion tools are in a separate repository: [`skill-interop-tools`](https://github.com/feiyueyun/skill-interop-tools).

## Forward Compatibility

The `schema_version` field identifies the current schema version (`"1.0.0"`), following Semver.

v1.x compatibility guarantees:

- ✅ Add optional fields
- ✅ Extend enum values
- ✅ Relax constraints
- ❌ No field removal
- ❌ No constraint tightening
- ❌ No semantic changes to existing fields

The parser uses a lenient strategy for unknown fields — they are preserved in `extra_metadata` without errors. A v1.0 parser can safely parse skill.json files containing fields added in v1.x.

## Project Structure

```
skill-manifest-spec/
├── spec/
│   ├── skill-manifest-v1.schema.json   # JSON Schema v1 (draft-07)
│   └── categories.json                 # Recommended category taxonomy
├── pkg/manifest/                       # Go parsing library
├── examples/                           # Example skill.json files
├── official-skills/                    # Platform-provided system skills
│   └── fyy-messenger/                  # Agent-to-Agent messaging skill
├── rfcs/                               # Standard evolution RFCs
├── CONTRIBUTING.md
└── LICENSE                             # Apache 2.0
```

## License

Apache License 2.0 — see [LICENSE](LICENSE).
