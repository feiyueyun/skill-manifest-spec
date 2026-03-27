// Copyright (c) 2025 Feiyueyun Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"pgregory.net/rapid"
)

// --- Property 1: Round-Trip ---
// **Validates: Req 13.1, Req 11.1, Req 11.4**
//
// For any valid Manifest m, Parse(Marshal(m)) produces a semantically
// equivalent struct. We verify by comparing Marshal(m) == Marshal(Parse(Marshal(m))).
func TestPBT_MarshalParseRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		m := genValidManifest(t)

		data1, err := Marshal(m)
		if err != nil {
			t.Fatalf("Marshal(m) error: %v", err)
		}

		m2, err := Parse(data1)
		if err != nil {
			t.Fatalf("Parse(Marshal(m)) error: %v", err)
		}

		data2, err := Marshal(m2)
		if err != nil {
			t.Fatalf("Marshal(Parse(Marshal(m))) error: %v", err)
		}

		if !bytes.Equal(data1, data2) {
			t.Fatalf("round-trip mismatch:\n  Marshal(m):               %s\n  Marshal(Parse(Marshal(m))): %s", string(data1), string(data2))
		}
	})
}

// --- Property 2: Pretty-Printer Round-Trip ---
// **Validates: Req 13.2, Req 11.5**
//
// For any valid Manifest m and indent string, Parse(MarshalIndent(m, indent))
// produces a semantically equivalent struct.
func TestPBT_MarshalIndentParseRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		m := genValidManifest(t)
		indent := rapid.SampledFrom([]string{"  ", "    ", "\t", " "}).Draw(t, "indent")

		indented, err := MarshalIndent(m, indent)
		if err != nil {
			t.Fatalf("MarshalIndent error: %v", err)
		}

		m2, err := Parse(indented)
		if err != nil {
			t.Fatalf("Parse(MarshalIndent(m)) error: %v", err)
		}

		// Compare via compact marshal
		data1, err := Marshal(m)
		if err != nil {
			t.Fatalf("Marshal(m) error: %v", err)
		}
		data2, err := Marshal(m2)
		if err != nil {
			t.Fatalf("Marshal(m2) error: %v", err)
		}

		if !bytes.Equal(data1, data2) {
			t.Fatalf("pretty-printer round-trip mismatch:\n  Marshal(m):                        %s\n  Marshal(Parse(MarshalIndent(m))): %s", string(data1), string(data2))
		}
	})
}

// --- Property 3: Parse Never Panics ---
// **Validates: Req 13.3**
//
// For any arbitrary byte slice, Parse(data) never panics.
func TestPBT_ParseNeverPanics(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		data := rapid.SliceOf(rapid.Byte()).Draw(t, "data")

		// Use deferred recover to catch panics.
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Parse panicked on input %q: %v", data, r)
				}
			}()
			_, _ = Parse(data)
		}()
	})
}

// --- Property 4: Valid Manifest No Errors ---
// **Validates: Req 13.4, Req 11.2**
//
// For any valid Manifest m from genValidManifest, Validate(m) returns
// no error-level ValidationErrors.
func TestPBT_ValidManifestNoErrors(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		m := genValidManifest(t)
		errs := Validate(m)

		for _, e := range errs {
			if e.Level == "error" {
				t.Fatalf("Validate returned error for valid manifest: %+v\nManifest: %+v", e, m)
			}
		}
	})
}

// --- Property 5: Marshal Output Schema Compliant ---
// **Validates: Req 13.5**
//
// For any valid Manifest m, Marshal(m) output passes JSON Schema validation.
func TestPBT_MarshalOutputSchemaCompliant(t *testing.T) {
	schema, err := loadSchema()
	if err != nil {
		t.Fatalf("failed to load schema: %v", err)
	}

	rapid.Check(t, func(t *rapid.T) {
		m := genValidManifest(t)

		data, err := Marshal(m)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		// Use jsonschema.UnmarshalJSON for correct type handling.
		doc, err := jsonschema.UnmarshalJSON(strings.NewReader(string(data)))
		if err != nil {
			t.Fatalf("jsonschema.UnmarshalJSON error: %v", err)
		}

		if validateErr := schema.Validate(doc); validateErr != nil {
			t.Fatalf("Marshal output failed schema validation: %v\nJSON: %s", validateErr, string(data))
		}
	})
}

// --- Property 6: Marshal Idempotent ---
// **Validates: Req 13.6, Req 13.7**
//
// For any valid Manifest m, Marshal(Parse(Marshal(m))) is byte-identical
// to Marshal(m).
func TestPBT_MarshalIdempotent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		m := genValidManifest(t)

		data1, err := Marshal(m)
		if err != nil {
			t.Fatalf("Marshal(m) error: %v", err)
		}

		m2, err := Parse(data1)
		if err != nil {
			t.Fatalf("Parse(Marshal(m)) error: %v", err)
		}

		data2, err := Marshal(m2)
		if err != nil {
			t.Fatalf("Marshal(Parse(Marshal(m))) error: %v", err)
		}

		if !bytes.Equal(data1, data2) {
			t.Fatalf("idempotency violated:\n  first:  %s\n  second: %s", string(data1), string(data2))
		}
	})
}

// --- Property 7: Conditional Deps Produce Errors ---
// **Validates: Req 2.1, 2.2, 2.4, Req 3.3, Req 12.1-3, 12.7**
//
// Generate manifests that violate conditional dependencies, verify Validate
// returns at least one error.
func TestPBT_ConditionalDepsProduceErrors(t *testing.T) {
	t.Run("type=service missing listen_port/protocol", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			m := &Manifest{
				Name:        genKebabName(t),
				Version:     genSemver(t),
				Description: genDescription(t),
				Type:        "service",
				// Intentionally omit ListenPort and Protocol.
			}
			errs := Validate(m)
			if !hasAnyErrorLevel(errs, "error") {
				t.Fatalf("expected error for service without listen_port/protocol, got: %+v", errs)
			}
		})
	})

	t.Run("type=mcp missing mcp field", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			m := &Manifest{
				Name:        genKebabName(t),
				Version:     genSemver(t),
				Description: genDescription(t),
				Type:        "mcp",
				// Intentionally omit MCP.
			}
			errs := Validate(m)
			if !hasAnyErrorLevel(errs, "error") {
				t.Fatalf("expected error for mcp without mcp field, got: %+v", errs)
			}
		})
	})

	t.Run("mcp_mode=proxy missing mcp_remote", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			m := &Manifest{
				Name:        genKebabName(t),
				Version:     genSemver(t),
				Description: genDescription(t),
				Type:        "mcp",
				MCP:         genMCPConfig(t),
				MCPMode:     ptr("proxy"),
				// Intentionally omit MCPRemote.
			}
			errs := Validate(m)
			if !hasAnyErrorLevel(errs, "error") {
				t.Fatalf("expected error for proxy without mcp_remote, got: %+v", errs)
			}
		})
	})

	t.Run("permissions contains network but no network_endpoints", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate some extra permissions, always include "network".
			perms := []string{"network"}
			extra := rapid.SliceOfN(
				rapid.SampledFrom([]string{"filesystem", "process", "env"}),
				0, 2,
			).Draw(t, "extra-perms")
			perms = append(perms, extra...)

			m := &Manifest{
				Name:        genKebabName(t),
				Version:     genSemver(t),
				Description: genDescription(t),
				Type:        "claw",
				Permissions: perms,
				// Intentionally omit NetworkEndpoints.
			}
			errs := Validate(m)
			if !hasAnyErrorLevel(errs, "error") {
				t.Fatalf("expected error for network permission without endpoints, got: %+v", errs)
			}
		})
	})

	t.Run("pricing.model non-free but no unit_price", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			model := rapid.SampledFrom([]string{"per_call", "per_minute", "per_token", "subscription"}).Draw(t, "pricing-model")
			m := &Manifest{
				Name:        genKebabName(t),
				Version:     genSemver(t),
				Description: genDescription(t),
				Type:        "claw",
				Pricing:     &Pricing{Model: &model},
				// Intentionally omit UnitPrice.
			}
			errs := Validate(m)
			if !hasAnyErrorLevel(errs, "error") {
				t.Fatalf("expected error for non-free pricing without unit_price, got: %+v", errs)
			}
		})
	})
}

// --- Property 8: Business Rules Produce Warnings ---
// **Validates: Req 3.5, Req 6.2**
//
// Generate manifests with business rule inconsistencies, verify Validate
// returns at least one warn.
func TestPBT_BusinessRulesProduceWarnings(t *testing.T) {
	t.Run("high-privilege permissions with risk_level < high", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Pick at least one high-privilege permission.
			highPriv := rapid.SampledFrom([]string{
				"browser", "camera", "microphone", "payment",
				"credential", "system_exec", "irreversible",
			}).Draw(t, "high-priv")

			riskLevel := rapid.SampledFrom([]string{"low", "medium"}).Draw(t, "risk-level")

			m := &Manifest{
				Name:        genKebabName(t),
				Version:     genSemver(t),
				Description: genDescription(t),
				Type:        "claw",
				Permissions: []string{highPriv},
				RiskLevel:   &riskLevel,
			}
			errs := Validate(m)
			if !hasAnyErrorLevel(errs, "warn") {
				t.Fatalf("expected warn for high-priv permissions with low risk_level, got: %+v", errs)
			}
		})
	})

	t.Run("write_memory=true read_memory=false", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			m := &Manifest{
				Name:        genKebabName(t),
				Version:     genSemver(t),
				Description: genDescription(t),
				Type:        "claw",
				MemoryAccess: &MemoryAccess{
					WriteMemory: ptr(true),
					ReadMemory:  ptr(false),
				},
			}
			errs := Validate(m)
			if !hasAnyErrorLevel(errs, "warn") {
				t.Fatalf("expected warn for write_memory=true + read_memory=false, got: %+v", errs)
			}
		})
	})
}

// --- Property 9: Unknown Fields Preserved ---
// **Validates: Req 11.6, Req 20.3, 20.4**
//
// Generate valid JSON with extra unknown fields, Parse it, verify unknown
// fields are in ExtraMetadata, Marshal and Parse again, verify they're still there.
func TestPBT_UnknownFieldsPreserved(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random extra field name that won't collide with known fields.
		extraKey := "x_custom_" + rapid.StringMatching(`[a-z]{3,10}`).Draw(t, "extra-key")
		extraValue := rapid.StringMatching(`[a-zA-Z0-9 ]{1,20}`).Draw(t, "extra-value")

		// Build a valid JSON with the extra field.
		jsonStr := fmt.Sprintf(`{
			"name": "test-skill",
			"version": "1.0.0",
			"description": "A test skill for unknown fields",
			"type": "claw",
			%q: %q
		}`, extraKey, extraValue)

		m, err := Parse([]byte(jsonStr))
		if err != nil {
			t.Fatalf("Parse error: %v", err)
		}

		// Verify the extra field is in ExtraMetadata.
		if m.ExtraMetadata == nil {
			t.Fatalf("ExtraMetadata is nil, expected key %q", extraKey)
		}
		val, ok := m.ExtraMetadata[extraKey]
		if !ok {
			t.Fatalf("ExtraMetadata missing key %q", extraKey)
		}
		if val != extraValue {
			t.Fatalf("ExtraMetadata[%q] = %v, want %q", extraKey, val, extraValue)
		}

		// Marshal and Parse again — verify preservation.
		data, err := Marshal(m)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		m2, err := Parse(data)
		if err != nil {
			t.Fatalf("Parse(Marshal(m)) error: %v", err)
		}

		if m2.ExtraMetadata == nil {
			t.Fatalf("ExtraMetadata nil after round-trip, expected key %q", extraKey)
		}
		val2, ok := m2.ExtraMetadata[extraKey]
		if !ok {
			t.Fatalf("ExtraMetadata missing key %q after round-trip", extraKey)
		}
		if val2 != extraValue {
			t.Fatalf("ExtraMetadata[%q] = %v after round-trip, want %q", extraKey, val2, extraValue)
		}
	})
}

// --- Property 10: Major Version Mismatch Rejected ---
// **Validates: Req 20.2**
//
// Generate schema_version with MAJOR != 1, verify Parse returns an error.
func TestPBT_MajorVersionMismatchRejected(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a MAJOR version that is NOT 1.
		major := rapid.IntRange(0, 100).Filter(func(v int) bool {
			return v != 1
		}).Draw(t, "major")
		minor := rapid.IntRange(0, 50).Draw(t, "minor")
		patch := rapid.IntRange(0, 50).Draw(t, "patch")
		schemaVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)

		jsonStr := fmt.Sprintf(`{
			"schema_version": %q,
			"name": "test-skill",
			"version": "1.0.0",
			"description": "A test skill for version check",
			"type": "claw"
		}`, schemaVersion)

		_, err := Parse([]byte(jsonStr))
		if err == nil {
			t.Fatalf("Parse should reject schema_version %q (MAJOR != 1), but got nil error", schemaVersion)
		}
		if !strings.Contains(err.Error(), "unsupported") && !strings.Contains(err.Error(), "schema_version") {
			t.Fatalf("error should mention version incompatibility, got: %v", err)
		}
	})
}

// --- Helper ---

// hasAnyErrorLevel checks if errs contains at least one entry at the given level.
func hasAnyErrorLevel(errs []ValidationError, level string) bool {
	for _, e := range errs {
		if e.Level == level {
			return true
		}
	}
	return false
}

// Ensure json import is used (for Property 9 test construction).
var _ = json.Marshal
