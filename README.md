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
```

The schema uses JSON Schema draft-07 with `if/then` conditional dependencies (e.g. `type=service` requires `listen_port` and `protocol`).

## Examples

| File | Skill Type | Highlights |
|------|-----------|------------|
| `examples/weather-lookup.skill.json` | `claw` | AI readability fields, permissions, i18n, interop, disclosure |
| `examples/code-review.skill.json` | `service` | listen_port, protocol, health_check, schedule, depends_on |
| `examples/database-query.skill.json` | `mcp` | MCP primitives (tools/resources/prompts), mcp_mode, ontology |

## Industry Compatibility

Skill Manifest is designed to interoperate with five major industry standards:

| Standard | Description | Integration |
|----------|-------------|-------------|
| **Agent Skills** (Anthropic) | SKILL.md portable skill format | `interop.agent_skills` + `disclosure` three-level progressive disclosure |
| **MCP** (Model Context Protocol) | Agent-tool integration protocol | `type=mcp` native support, `mcp` field declares tools/resources/prompts |
| **A2A** (Google Agent-to-Agent) | Agent communication protocol | `interop.a2a_card` maps to A2A Agent Card |
| **ClawHub** | OpenClaw skill marketplace ecosystem | `interop.clawhub` maps to claw.json format |
| **SkillNet** (ZJU/OpenKG) | 200,000+ skill knowledge graph | `ontology` field references three-level skill ontology |

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
├── spec/                  # JSON Schema definition
├── pkg/manifest/          # Go parsing library
├── examples/              # Example skill.json files
├── rfcs/                  # Standard evolution RFCs
├── CONTRIBUTING.md
└── LICENSE                # Apache 2.0
```

## License

Apache License 2.0 — see [LICENSE](LICENSE).
