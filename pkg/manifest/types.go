// Copyright (c) 2025 Feiyueyun Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package manifest

import (
	"encoding/json"
	"fmt"
)

// SchemaVersion is the current Skill Manifest standard version.
const SchemaVersion = "1.0.0"

// ValidationError represents a single validation error or warning.
type ValidationError struct {
	Field   string // JSON field path, e.g. "mcp_remote.endpoint"
	Message string // Error description
	Level   string // "error" or "warn"
}

// Manifest represents a complete skill.json file.
type Manifest struct {
	// Core required fields
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Type        string `json:"type"`

	// Schema metadata
	SchemaVersion *string `json:"schema_version,omitempty"`

	// Service type specific fields
	EntryPoint         *string    `json:"entry_point,omitempty"`
	AutoStart          *bool      `json:"auto_start,omitempty"`
	ListenPort         *int       `json:"listen_port,omitempty"`
	Protocol           *string    `json:"protocol,omitempty"`
	MCP                *MCPConfig `json:"mcp,omitempty"`
	MCPMode            *string    `json:"mcp_mode,omitempty"`
	MCPRemote          *MCPRemote `json:"mcp_remote,omitempty"`
	OpenClawCompatible *bool      `json:"openclaw_compatible,omitempty"`

	// Permissions and network declarations
	Permissions      []string `json:"permissions,omitempty"`
	NetworkEndpoints []string `json:"network_endpoints,omitempty"`
	RiskLevel        *string  `json:"risk_level,omitempty"`
	SandboxOverride  *string  `json:"sandbox_override,omitempty"`
	SeccompProfile   *string  `json:"seccomp_profile,omitempty"`
	AllowedNetworks  []string `json:"allowed_networks,omitempty"`
	Visibility       *string  `json:"visibility,omitempty"`

	// Capability declarations and AI readability
	Capabilities   []string `json:"capabilities,omitempty"`
	WhenToUse      *string  `json:"when_to_use,omitempty"`
	Category       *string  `json:"category,omitempty"`
	Models         []string `json:"models,omitempty"`
	InputExamples  []string `json:"input_examples,omitempty"`
	OutputExamples []string `json:"output_examples,omitempty"`
	RelatedSkills  []string `json:"related_skills,omitempty"`
	RoleAffinity   []string `json:"role_affinity,omitempty"`
	Tags           []string `json:"tags,omitempty"`

	// Commercialization fields
	Pricing          *Pricing `json:"pricing,omitempty"`
	SLA              *SLA     `json:"sla,omitempty"`
	Resellable       *bool    `json:"resellable,omitempty"`
	ResaleCommission *float64 `json:"resale_commission,omitempty"`

	// Memory access, scheduling, and remote config
	MemoryAccess *MemoryAccess `json:"memory_access,omitempty"`
	Schedule     *Schedule     `json:"schedule,omitempty"`
	RemoteConfig *bool         `json:"remote_config,omitempty"`

	// Dependency management
	DependsOn          []Dependency `json:"depends_on,omitempty"`
	ConflictsWith      []string     `json:"conflicts_with,omitempty"`
	MinCLIVersion      *string      `json:"min_cli_version,omitempty"`
	MinPlatformVersion *string      `json:"min_platform_version,omitempty"`

	// Internationalization
	Locale  *string           `json:"locale,omitempty"`
	Locales map[string]Locale `json:"locales,omitempty"`

	// Runtime configuration
	Requirements   *Requirements   `json:"requirements,omitempty"`
	ResourceLimits *ResourceLimits `json:"resource_limits,omitempty"`
	DataDir        *string         `json:"data_dir,omitempty"`
	Timeout        *int            `json:"timeout,omitempty"`
	HealthCheck    *HealthCheck    `json:"health_check,omitempty"`

	// Interoperability
	Interop          *Interop       `json:"interop,omitempty"`
	Disclosure       *Disclosure    `json:"disclosure,omitempty"`
	Ontology         *Ontology      `json:"ontology,omitempty"`
	InstructionsPath *string        `json:"instructions_path,omitempty"`
	ExtraMetadata    map[string]any `json:"extra_metadata,omitempty"`
}

// --- Nested types ---

// MCPConfig declares MCP primitives.
type MCPConfig struct {
	Tools     []MCPTool     `json:"tools,omitempty"`
	Resources []MCPResource `json:"resources,omitempty"`
	Prompts   []any         `json:"prompts,omitempty"`
}

// MCPTool declares a single MCP tool.
type MCPTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

// MCPResource declares a single MCP resource.
type MCPResource struct {
	URI      string `json:"uri"`
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
}

// MCPRemote holds proxy bridge configuration.
type MCPRemote struct {
	Endpoint  string            `json:"endpoint"`
	Transport *string           `json:"transport,omitempty"`
	AuthType  *string           `json:"auth_type,omitempty"`
	AuthRef   *string           `json:"auth_ref,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	TimeoutMs *int              `json:"timeout_ms,omitempty"`
	CacheTTLs *int              `json:"cache_ttl_s,omitempty"`
}

// Pricing holds commercialization pricing fields.
type Pricing struct {
	Model     *string  `json:"model,omitempty"`
	UnitPrice *float64 `json:"unit_price,omitempty"`
	Currency  *string  `json:"currency,omitempty"`
	FreeTier  *int     `json:"free_tier,omitempty"`
}

// SLA holds service level agreement fields.
type SLA struct {
	Availability   *float64 `json:"availability,omitempty"`
	MaxResponseMs  *int     `json:"max_response_ms,omitempty"`
	MaxConcurrency *int     `json:"max_concurrency,omitempty"`
}

// MemoryAccess declares memory access requirements.
type MemoryAccess struct {
	ReadSoul    *bool `json:"read_soul,omitempty"`
	ReadMemory  *bool `json:"read_memory,omitempty"`
	WriteMemory *bool `json:"write_memory,omitempty"`
}

// Schedule holds cron-based scheduling configuration.
type Schedule struct {
	Cron     string  `json:"cron"`
	Timezone *string `json:"timezone,omitempty"`
	Enabled  *bool   `json:"enabled,omitempty"`
	Action   *string `json:"action,omitempty"`
	Params   any     `json:"params,omitempty"`
}

// Dependency declares a skill dependency.
// Supports both simple format (string) and full format (object).
// Unified into Dependency struct during parsing.
type Dependency struct {
	Name    string  `json:"name"`
	Version *string `json:"version,omitempty"`
}

// UnmarshalJSON implements custom JSON deserialization, supporting both string and object formats.
func (d *Dependency) UnmarshalJSON(data []byte) error {
	// Try string format
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		d.Name = s
		d.Version = nil
		return nil
	}
	// Try object format
	type alias Dependency
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return fmt.Errorf("depends_on element must be string or {name, version} object: %w", err)
	}
	*d = Dependency(a)
	return nil
}

// Locale holds a multilingual description entry.
type Locale struct {
	Description   *string  `json:"description,omitempty"`
	WhenToUse     *string  `json:"when_to_use,omitempty"`
	InputExamples []string `json:"input_examples,omitempty"`
}

// Requirements holds runtime dependency information.
type Requirements struct {
	Runtime      *string  `json:"runtime,omitempty"`
	MinVersion   *string  `json:"min_version,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// ResourceLimits holds resource constraint fields.
type ResourceLimits struct {
	CPU    *string `json:"cpu,omitempty"`
	Memory *string `json:"memory,omitempty"`
}

// HealthCheck holds health check configuration.
type HealthCheck struct {
	Endpoint *string `json:"endpoint,omitempty"`
	Interval *int    `json:"interval,omitempty"`
	Timeout  *int    `json:"timeout,omitempty"`
}

// Interop holds interoperability configuration.
type Interop struct {
	AgentSkills *bool `json:"agent_skills,omitempty"`
	ClawHub     *bool `json:"clawhub,omitempty"`
	A2ACard     *bool `json:"a2a_card,omitempty"`
}

// Disclosure holds progressive disclosure configuration.
type Disclosure struct {
	Level0 []string `json:"level_0,omitempty"`
	Level1 []string `json:"level_1,omitempty"`
	Level2 []string `json:"level_2,omitempty"`
}

// Ontology holds SkillNet skill ontology references.
type Ontology struct {
	Domain    *string `json:"domain,omitempty"`
	Category  *string `json:"category,omitempty"`
	SkillNode *string `json:"skill_node,omitempty"`
}
