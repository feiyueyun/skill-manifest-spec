// Copyright (c) 2025 Feiyueyun Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package manifest

import (
	"encoding/json"
	"strings"
	"testing"
)

// minimalManifest returns a valid Manifest with only the 4 required fields.
func minimalManifest() Manifest {
	return Manifest{
		Name:        "my-skill",
		Version:     "1.0.0",
		Description: "A minimal skill for testing purposes",
		Type:        "claw",
	}
}

// ptr is a generic helper to get a pointer to a value.
func ptr[T any](v T) *T { return &v }

// --- Parse tests ---

func TestParse_ValidMinimalManifest(t *testing.T) {
	data := []byte(`{
		"name": "my-skill",
		"version": "1.0.0",
		"description": "A minimal skill for testing purposes",
		"type": "claw"
	}`)

	m, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if m.Name != "my-skill" {
		t.Errorf("Name = %q, want %q", m.Name, "my-skill")
	}
	if m.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", m.Version, "1.0.0")
	}
	if m.Description != "A minimal skill for testing purposes" {
		t.Errorf("Description = %q, want %q", m.Description, "A minimal skill for testing purposes")
	}
	if m.Type != "claw" {
		t.Errorf("Type = %q, want %q", m.Type, "claw")
	}
}

func TestParse_JSONSyntaxErrorWithLineNumber(t *testing.T) {
	// Trailing comma on line 3
	data := []byte(`{
  "name": "test",
  "version": "1.0.0",
}`)

	_, err := Parse(data)
	if err == nil {
		t.Fatal("Parse() expected error for malformed JSON, got nil")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "line") || !strings.Contains(errMsg, "column") {
		t.Errorf("error should contain line/column info, got: %s", errMsg)
	}
}

func TestParse_SchemaVersionMajorMismatch(t *testing.T) {
	data := []byte(`{
		"schema_version": "2.0.0",
		"name": "my-skill",
		"version": "1.0.0",
		"description": "A test skill description",
		"type": "claw"
	}`)

	_, err := Parse(data)
	if err == nil {
		t.Fatal("Parse() expected error for schema_version 2.0.0, got nil")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "2.0.0") {
		t.Errorf("error should mention version 2.0.0, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "MAJOR version 1") && !strings.Contains(errMsg, "unsupported") {
		t.Errorf("error should mention version incompatibility, got: %s", errMsg)
	}
}

func TestParse_UnknownFieldsPreserved(t *testing.T) {
	data := []byte(`{
		"name": "my-skill",
		"version": "1.0.0",
		"description": "A test skill description",
		"type": "claw",
		"custom_field": "hello",
		"another_unknown": 42
	}`)

	m, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if m.ExtraMetadata == nil {
		t.Fatal("ExtraMetadata should not be nil")
	}
	if v, ok := m.ExtraMetadata["custom_field"]; !ok || v != "hello" {
		t.Errorf("ExtraMetadata[custom_field] = %v, want %q", v, "hello")
	}
	if _, ok := m.ExtraMetadata["another_unknown"]; !ok {
		t.Error("ExtraMetadata should contain another_unknown")
	}
}

// --- Validate tests ---

// hasValidationError checks if errs contains at least one error at the given level
// whose Field or Message contains fieldSubstr.
func hasValidationError(errs []ValidationError, level, fieldSubstr string) bool {
	for _, e := range errs {
		if e.Level == level && (strings.Contains(e.Field, fieldSubstr) || strings.Contains(e.Message, fieldSubstr)) {
			return true
		}
	}
	return false
}

func TestValidate_NameRegex(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid two chars", "ab", false},
		{"valid kebab-case", "a-b-c", false},
		{"valid with digits", "my-skill-01", false},
		{"invalid uppercase", "A", true},
		{"invalid leading dash", "-abc", true},
		{"invalid single char", "a", true},
		{"invalid too long", "a" + strings.Repeat("b", 64), true}, // 65 chars total, pattern allows max 64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := minimalManifest()
			m.Name = tt.value
			errs := Validate(&m)
			gotErr := hasValidationError(errs, "error", "name")
			if gotErr != tt.wantErr {
				t.Errorf("Validate() name=%q: gotError=%v, wantError=%v, errs=%+v", tt.value, gotErr, tt.wantErr, errs)
			}
		})
	}
}

func TestValidate_SemverFormat(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid basic", "1.0.0", false},
		{"valid prerelease", "0.1.0-alpha", false},
		{"valid with build", "1.0.0+build.123", false},
		{"invalid two parts", "1.0", true},
		{"invalid v prefix", "v1.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := minimalManifest()
			m.Version = tt.value
			errs := Validate(&m)
			gotErr := hasValidationError(errs, "error", "version")
			if gotErr != tt.wantErr {
				t.Errorf("Validate() version=%q: gotError=%v, wantError=%v, errs=%+v", tt.value, gotErr, tt.wantErr, errs)
			}
		})
	}
}

func TestValidate_DescriptionMinLength(t *testing.T) {
	m := minimalManifest()
	m.Description = "too short" // 9 chars
	errs := Validate(&m)
	if !hasValidationError(errs, "error", "description") {
		t.Errorf("Validate() should error for description < 10 chars, errs=%+v", errs)
	}
}

func TestValidate_TypeServiceConditional(t *testing.T) {
	t.Run("missing listen_port and protocol", func(t *testing.T) {
		m := minimalManifest()
		m.Type = "service"
		errs := Validate(&m)
		if !hasValidationError(errs, "error", "listen_port") {
			t.Errorf("expected error for missing listen_port, errs=%+v", errs)
		}
		if !hasValidationError(errs, "error", "protocol") {
			t.Errorf("expected error for missing protocol, errs=%+v", errs)
		}
	})

	t.Run("valid service with required fields", func(t *testing.T) {
		m := minimalManifest()
		m.Type = "service"
		m.ListenPort = ptr(8080)
		m.Protocol = ptr("grpc")
		errs := Validate(&m)
		if hasValidationError(errs, "error", "listen_port") || hasValidationError(errs, "error", "protocol") {
			t.Errorf("unexpected errors for valid service, errs=%+v", errs)
		}
	})
}

func TestValidate_TypeMCPConditional(t *testing.T) {
	t.Run("missing mcp field", func(t *testing.T) {
		m := minimalManifest()
		m.Type = "mcp"
		errs := Validate(&m)
		if !hasValidationError(errs, "error", "mcp") {
			t.Errorf("expected error for missing mcp field, errs=%+v", errs)
		}
	})

	t.Run("valid mcp with required fields", func(t *testing.T) {
		m := minimalManifest()
		m.Type = "mcp"
		m.MCP = &MCPConfig{
			Tools: []MCPTool{
				{
					Name:        "test-tool",
					Description: "A test tool",
					InputSchema: map[string]any{"type": "object"},
				},
			},
		}
		errs := Validate(&m)
		if hasValidationError(errs, "error", "mcp") {
			t.Errorf("unexpected mcp error for valid mcp manifest, errs=%+v", errs)
		}
	})
}

func TestValidate_MCPModeProxyConditional(t *testing.T) {
	t.Run("proxy mode without mcp_remote", func(t *testing.T) {
		m := minimalManifest()
		m.Type = "mcp"
		m.MCP = &MCPConfig{
			Tools: []MCPTool{
				{
					Name:        "test-tool",
					Description: "A test tool",
					InputSchema: map[string]any{"type": "object"},
				},
			},
		}
		m.MCPMode = ptr("proxy")
		errs := Validate(&m)
		if !hasValidationError(errs, "error", "mcp_remote") {
			t.Errorf("expected error for missing mcp_remote, errs=%+v", errs)
		}
	})

	t.Run("proxy mode with mcp_remote", func(t *testing.T) {
		m := minimalManifest()
		m.Type = "mcp"
		m.MCP = &MCPConfig{
			Tools: []MCPTool{
				{
					Name:        "test-tool",
					Description: "A test tool",
					InputSchema: map[string]any{"type": "object"},
				},
			},
		}
		m.MCPMode = ptr("proxy")
		m.MCPRemote = &MCPRemote{
			Endpoint: "https://example.com/mcp",
		}
		errs := Validate(&m)
		if hasValidationError(errs, "error", "mcp_remote") {
			t.Errorf("unexpected mcp_remote error, errs=%+v", errs)
		}
	})
}

func TestValidate_NetworkPermissionConditional(t *testing.T) {
	t.Run("network permission without network_endpoints", func(t *testing.T) {
		m := minimalManifest()
		m.Permissions = []string{"network"}
		errs := Validate(&m)
		if !hasValidationError(errs, "error", "network_endpoints") {
			t.Errorf("expected error for missing network_endpoints, errs=%+v", errs)
		}
	})

	t.Run("network permission with network_endpoints", func(t *testing.T) {
		m := minimalManifest()
		m.Permissions = []string{"network"}
		m.NetworkEndpoints = []string{"api.example.com"}
		errs := Validate(&m)
		if hasValidationError(errs, "error", "network_endpoints") {
			t.Errorf("unexpected network_endpoints error, errs=%+v", errs)
		}
	})
}

func TestValidate_PermissionsUniqueItems(t *testing.T) {
	m := minimalManifest()
	m.Permissions = []string{"network", "network"}
	m.NetworkEndpoints = []string{"api.example.com"} // satisfy conditional dep
	errs := Validate(&m)
	if !hasValidationError(errs, "error", "permissions") {
		t.Errorf("expected error for duplicate permissions, errs=%+v", errs)
	}
}

func TestValidate_CronFormat(t *testing.T) {
	tests := []struct {
		name    string
		cron    string
		wantErr bool
	}{
		{"valid every minute", "* * * * *", false},
		{"valid specific time", "30 2 * * 1", false},
		{"valid range", "0-30 * * * *", false},
		{"valid step", "*/5 * * * *", false},
		{"invalid too few fields", "* * *", true},
		{"invalid too many fields", "* * * * * *", true},
		{"invalid value out of range", "60 * * * *", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := minimalManifest()
			m.Schedule = &Schedule{Cron: tt.cron}
			errs := Validate(&m)
			gotErr := hasValidationError(errs, "error", "schedule.cron")
			if gotErr != tt.wantErr {
				t.Errorf("Validate() cron=%q: gotError=%v, wantError=%v, errs=%+v", tt.cron, gotErr, tt.wantErr, errs)
			}
		})
	}
}

func TestValidate_SemverRangeSyntax(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"valid exact", "1.0.0", false},
		{"valid caret", "^1.2.3", false},
		{"valid tilde", "~1.2.3", false},
		{"valid gte", ">=1.0.0", false},
		{"valid range", ">=1.0.0 <2.0.0", false},
		{"invalid incomplete", ">=1.2", true},
		{"invalid bare text", "latest", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := minimalManifest()
			m.DependsOn = []Dependency{
				{Name: "some-dep", Version: ptr(tt.version)},
			}
			errs := Validate(&m)
			gotErr := hasValidationError(errs, "error", "depends_on")
			if gotErr != tt.wantErr {
				t.Errorf("Validate() version=%q: gotError=%v, wantError=%v, errs=%+v", tt.version, gotErr, tt.wantErr, errs)
			}
		})
	}
}

func TestValidate_BCP47Format(t *testing.T) {
	tests := []struct {
		name    string
		locale  string
		wantErr bool
	}{
		{"valid zh-CN", "zh-CN", false},
		{"valid en-US", "en-US", false},
		{"valid en", "en", false},
		{"invalid numeric", "123", true},
		{"invalid too long subtag", "toolongsubtag-toolongsubtag", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := minimalManifest()
			m.Locale = ptr(tt.locale)
			errs := Validate(&m)
			gotErr := hasValidationError(errs, "error", "locale")
			if gotErr != tt.wantErr {
				t.Errorf("Validate() locale=%q: gotError=%v, wantError=%v, errs=%+v", tt.locale, gotErr, tt.wantErr, errs)
			}
		})
	}
}

func TestValidate_PricingConditional(t *testing.T) {
	t.Run("non-free model without unit_price", func(t *testing.T) {
		m := minimalManifest()
		m.Pricing = &Pricing{Model: ptr("per_call")}
		errs := Validate(&m)
		if !hasValidationError(errs, "error", "pricing.unit_price") {
			t.Errorf("expected error for missing unit_price, errs=%+v", errs)
		}
	})

	t.Run("free model without unit_price is ok", func(t *testing.T) {
		m := minimalManifest()
		m.Pricing = &Pricing{Model: ptr("free")}
		errs := Validate(&m)
		if hasValidationError(errs, "error", "pricing.unit_price") {
			t.Errorf("unexpected unit_price error for free model, errs=%+v", errs)
		}
	})

	t.Run("non-free model with unit_price is ok", func(t *testing.T) {
		m := minimalManifest()
		m.Pricing = &Pricing{Model: ptr("per_call"), UnitPrice: ptr(0.01)}
		errs := Validate(&m)
		if hasValidationError(errs, "error", "pricing.unit_price") {
			t.Errorf("unexpected unit_price error, errs=%+v", errs)
		}
	})
}

func TestValidate_MCPToolInputSchema(t *testing.T) {
	t.Run("inputSchema missing type field", func(t *testing.T) {
		m := minimalManifest()
		m.Type = "mcp"
		m.MCP = &MCPConfig{
			Tools: []MCPTool{
				{
					Name:        "bad-tool",
					Description: "A tool with bad schema",
					InputSchema: map[string]any{"properties": map[string]any{}},
				},
			},
		}
		errs := Validate(&m)
		if !hasValidationError(errs, "error", "mcp.tools[0].inputSchema") {
			t.Errorf("expected error for inputSchema missing type, errs=%+v", errs)
		}
	})

	t.Run("valid inputSchema", func(t *testing.T) {
		m := minimalManifest()
		m.Type = "mcp"
		m.MCP = &MCPConfig{
			Tools: []MCPTool{
				{
					Name:        "good-tool",
					Description: "A tool with valid schema",
					InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
				},
			},
		}
		errs := Validate(&m)
		if hasValidationError(errs, "error", "mcp.tools[0].inputSchema") {
			t.Errorf("unexpected inputSchema error, errs=%+v", errs)
		}
	})
}

func TestValidate_RiskLevelPermissionsWarning(t *testing.T) {
	t.Run("high-privilege permissions with low risk_level", func(t *testing.T) {
		m := minimalManifest()
		m.Permissions = []string{"payment", "credential"}
		m.NetworkEndpoints = []string{"api.example.com"} // not needed here but safe
		m.RiskLevel = ptr("low")
		errs := Validate(&m)
		if !hasValidationError(errs, "warn", "risk_level") {
			t.Errorf("expected warn for risk_level mismatch, errs=%+v", errs)
		}
	})

	t.Run("high-privilege permissions with high risk_level", func(t *testing.T) {
		m := minimalManifest()
		m.Permissions = []string{"payment"}
		m.RiskLevel = ptr("high")
		errs := Validate(&m)
		if hasValidationError(errs, "warn", "risk_level") {
			t.Errorf("unexpected warn for matching risk_level, errs=%+v", errs)
		}
	})
}

func TestValidate_MemoryAccessWarning(t *testing.T) {
	t.Run("write_memory true read_memory false", func(t *testing.T) {
		m := minimalManifest()
		m.MemoryAccess = &MemoryAccess{
			WriteMemory: ptr(true),
			ReadMemory:  ptr(false),
		}
		errs := Validate(&m)
		if !hasValidationError(errs, "warn", "memory_access") {
			t.Errorf("expected warn for write without read, errs=%+v", errs)
		}
	})

	t.Run("write_memory true read_memory true", func(t *testing.T) {
		m := minimalManifest()
		m.MemoryAccess = &MemoryAccess{
			WriteMemory: ptr(true),
			ReadMemory:  ptr(true),
		}
		errs := Validate(&m)
		if hasValidationError(errs, "warn", "memory_access") {
			t.Errorf("unexpected warn when both read and write are true, errs=%+v", errs)
		}
	})
}

// --- Marshal tests ---

func TestMarshal_RoundTrip(t *testing.T) {
	m := minimalManifest()
	m.Tags = []string{"test", "example"}
	m.Locale = ptr("en-US")

	data, err := Marshal(&m)
	if err != nil {
		t.Fatalf("Marshal() unexpected error: %v", err)
	}

	m2, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if m2.Name != m.Name {
		t.Errorf("Name = %q, want %q", m2.Name, m.Name)
	}
	if m2.Version != m.Version {
		t.Errorf("Version = %q, want %q", m2.Version, m.Version)
	}
	if m2.Description != m.Description {
		t.Errorf("Description = %q, want %q", m2.Description, m.Description)
	}
	if m2.Type != m.Type {
		t.Errorf("Type = %q, want %q", m2.Type, m.Type)
	}
	if len(m2.Tags) != len(m.Tags) {
		t.Errorf("Tags length = %d, want %d", len(m2.Tags), len(m.Tags))
	}
	if m2.Locale == nil || *m2.Locale != "en-US" {
		t.Errorf("Locale = %v, want %q", m2.Locale, "en-US")
	}
}

func TestMarshalIndent_Formatting(t *testing.T) {
	m := minimalManifest()

	data, err := MarshalIndent(&m, "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() unexpected error: %v", err)
	}

	// Verify it's valid JSON
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("MarshalIndent output is not valid JSON: %v", err)
	}

	// Verify indentation is present
	output := string(data)
	if !strings.Contains(output, "\n") {
		t.Error("MarshalIndent output should contain newlines")
	}
	if !strings.Contains(output, "  ") {
		t.Error("MarshalIndent output should contain indentation")
	}

	// Verify round-trip
	m2, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(MarshalIndent) unexpected error: %v", err)
	}
	if m2.Name != m.Name {
		t.Errorf("round-trip Name = %q, want %q", m2.Name, m.Name)
	}
}
