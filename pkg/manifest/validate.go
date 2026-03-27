// Copyright (c) 2025 Feiyueyun Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package manifest

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// englishPrinter is used to format jsonschema error messages.
var englishPrinter = message.NewPrinter(language.English)

// Validate performs JSON Schema constraint validation and business rule validation on a Manifest.
// Returns all validation errors (does not stop at the first error).
// Returns an empty slice for a valid Manifest.
func Validate(m *Manifest) []ValidationError {
	var errs []ValidationError

	// Layer 1: JSON Schema structural validation
	errs = append(errs, validateSchema(m)...)

	// Layer 2: Business rule validation
	errs = append(errs, validateBusinessRules(m)...)

	// Layer 3: Custom format validators
	errs = append(errs, validateCustomFormats(m)...)

	if errs == nil {
		errs = []ValidationError{}
	}
	return errs
}

// --- Layer 1: JSON Schema structural validation ---

func validateSchema(m *Manifest) []ValidationError {
	schema, err := loadSchema()
	if err != nil {
		return []ValidationError{{
			Field:   "",
			Message: fmt.Sprintf("failed to load schema: %v", err),
			Level:   "error",
		}}
	}

	// Marshal the Manifest to JSON, then unmarshal to an any value
	// suitable for the jsonschema library.
	data, err := json.Marshal(m)
	if err != nil {
		return []ValidationError{{
			Field:   "",
			Message: fmt.Sprintf("failed to marshal manifest for validation: %v", err),
			Level:   "error",
		}}
	}

	var doc any
	if err := json.Unmarshal(data, &doc); err != nil {
		return []ValidationError{{
			Field:   "",
			Message: fmt.Sprintf("failed to unmarshal manifest for validation: %v", err),
			Level:   "error",
		}}
	}

	// Use jsonschema.UnmarshalJSON for correct type handling (numbers as json.Number etc.)
	doc, err = jsonschema.UnmarshalJSON(strings.NewReader(string(data)))
	if err != nil {
		return []ValidationError{{
			Field:   "",
			Message: fmt.Sprintf("failed to unmarshal manifest for schema validation: %v", err),
			Level:   "error",
		}}
	}

	validateErr := schema.Validate(doc)
	if validateErr == nil {
		return nil
	}

	valErr, ok := validateErr.(*jsonschema.ValidationError)
	if !ok {
		return []ValidationError{{
			Field:   "",
			Message: fmt.Sprintf("schema validation error: %v", validateErr),
			Level:   "error",
		}}
	}

	return collectSchemaErrors(valErr)
}

// collectSchemaErrors recursively walks the jsonschema.ValidationError tree
// and extracts leaf errors with their field paths.
func collectSchemaErrors(ve *jsonschema.ValidationError) []ValidationError {
	var errs []ValidationError

	// If this node has causes, recurse into them.
	if len(ve.Causes) > 0 {
		for _, cause := range ve.Causes {
			errs = append(errs, collectSchemaErrors(cause)...)
		}
		return errs
	}

	// Leaf error — extract field path and message.
	field := instanceLocationToField(ve.InstanceLocation)
	msg := ve.ErrorKind.LocalizedString(englishPrinter)

	errs = append(errs, ValidationError{
		Field:   field,
		Message: msg,
		Level:   "error",
	})
	return errs
}

// instanceLocationToField converts a jsonschema InstanceLocation ([]string)
// to a dot-separated field path with array indices in brackets.
// e.g. ["mcp", "tools", "0", "inputSchema"] → "mcp.tools[0].inputSchema"
func instanceLocationToField(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	var result []string
	for i, p := range parts {
		// If the part is a numeric index, attach it to the previous part
		if _, err := strconv.Atoi(p); err == nil && i > 0 {
			result[len(result)-1] = result[len(result)-1] + "[" + p + "]"
			continue
		}
		result = append(result, p)
	}
	return strings.Join(result, ".")
}

// --- Layer 2: Business rule validation ---

func validateBusinessRules(m *Manifest) []ValidationError {
	var errs []ValidationError

	// Rule: pricing.model is not "free" → unit_price is required
	errs = append(errs, validatePricingRules(m)...)

	// Rule: risk_level vs permissions consistency
	errs = append(errs, validateRiskPermissions(m)...)

	// Rule: memory_access write_memory=true + read_memory=false → warn
	errs = append(errs, validateMemoryAccess(m)...)

	return errs
}

// highPrivilegePermissions are permissions that should require risk_level >= "high".
var highPrivilegePermissions = map[string]bool{
	"browser":      true,
	"camera":       true,
	"microphone":   true,
	"payment":      true,
	"credential":   true,
	"system_exec":  true,
	"irreversible": true,
}

// riskLevelOrder maps risk levels to numeric order for comparison.
var riskLevelOrder = map[string]int{
	"low":      0,
	"medium":   1,
	"high":     2,
	"critical": 3,
}

func validatePricingRules(m *Manifest) []ValidationError {
	if m.Pricing == nil || m.Pricing.Model == nil {
		return nil
	}
	model := *m.Pricing.Model
	if model != "free" && m.Pricing.UnitPrice == nil {
		return []ValidationError{{
			Field:   "pricing.unit_price",
			Message: fmt.Sprintf("unit_price is required when pricing.model is %q", model),
			Level:   "error",
		}}
	}
	return nil
}

func validateRiskPermissions(m *Manifest) []ValidationError {
	if len(m.Permissions) == 0 {
		return nil
	}

	// Collect high-privilege permissions present
	var highPerms []string
	for _, p := range m.Permissions {
		if highPrivilegePermissions[p] {
			highPerms = append(highPerms, p)
		}
	}
	if len(highPerms) == 0 {
		return nil
	}

	// Determine effective risk level (default is "low")
	riskLevel := "low"
	if m.RiskLevel != nil {
		riskLevel = *m.RiskLevel
	}

	order, ok := riskLevelOrder[riskLevel]
	if !ok {
		// Unknown risk level — schema validation will catch this
		return nil
	}

	if order < riskLevelOrder["high"] {
		return []ValidationError{{
			Field: "risk_level",
			Message: fmt.Sprintf(
				"risk_level is %q but permissions contain high-privilege items %v; consider setting risk_level to \"high\" or above",
				riskLevel, highPerms,
			),
			Level: "warn",
		}}
	}
	return nil
}

func validateMemoryAccess(m *Manifest) []ValidationError {
	if m.MemoryAccess == nil {
		return nil
	}
	writeMemory := m.MemoryAccess.WriteMemory != nil && *m.MemoryAccess.WriteMemory
	readMemory := m.MemoryAccess.ReadMemory != nil && *m.MemoryAccess.ReadMemory

	if writeMemory && !readMemory {
		return []ValidationError{{
			Field:   "memory_access",
			Message: "write_memory is true but read_memory is false; writing without reading is usually unintended",
			Level:   "warn",
		}}
	}
	return nil
}

// --- Layer 3: Custom format validators ---

func validateCustomFormats(m *Manifest) []ValidationError {
	var errs []ValidationError

	// Cron expression validation
	errs = append(errs, validateCron(m)...)

	// Semver range syntax validation
	errs = append(errs, validateSemverRanges(m)...)

	// BCP-47 language tag validation
	errs = append(errs, validateBCP47(m)...)

	// MCP tools inputSchema validity
	errs = append(errs, validateMCPToolInputSchemas(m)...)

	return errs
}

// cronFieldPatterns defines valid patterns for each of the 5 cron fields.
// Fields: minute, hour, day-of-month, month, day-of-week
var cronFieldPatterns = [5]struct {
	name string
	min  int
	max  int
}{
	{"minute", 0, 59},
	{"hour", 0, 23},
	{"day-of-month", 1, 31},
	{"month", 1, 12},
	{"day-of-week", 0, 7}, // 0 and 7 both represent Sunday
}

// cronFieldRegex matches a single cron field element (value, range, step, or wildcard).
var cronFieldRegex = regexp.MustCompile(`^(\*|[0-9]+(-[0-9]+)?)(/[0-9]+)?$`)

func validateCron(m *Manifest) []ValidationError {
	if m.Schedule == nil || m.Schedule.Cron == "" {
		return nil
	}

	cron := strings.TrimSpace(m.Schedule.Cron)
	fields := strings.Fields(cron)
	if len(fields) != 5 {
		return []ValidationError{{
			Field:   "schedule.cron",
			Message: fmt.Sprintf("invalid cron expression: expected 5 fields, got %d", len(fields)),
			Level:   "error",
		}}
	}

	for i, field := range fields {
		if err := validateCronField(field, cronFieldPatterns[i].name, cronFieldPatterns[i].min, cronFieldPatterns[i].max); err != nil {
			return []ValidationError{{
				Field:   "schedule.cron",
				Message: fmt.Sprintf("invalid cron expression: %s field %q: %v", cronFieldPatterns[i].name, field, err),
				Level:   "error",
			}}
		}
	}
	return nil
}

func validateCronField(field, _ string, min, max int) error {
	// Handle comma-separated lists
	parts := strings.Split(field, ",")
	for _, part := range parts {
		if err := validateCronPart(part, min, max); err != nil {
			return err
		}
	}
	return nil
}

func validateCronPart(part string, min, max int) error {
	if !cronFieldRegex.MatchString(part) {
		return fmt.Errorf("invalid syntax %q", part)
	}

	// Split on "/" for step values
	stepParts := strings.SplitN(part, "/", 2)
	rangePart := stepParts[0]

	if len(stepParts) == 2 {
		step, err := strconv.Atoi(stepParts[1])
		if err != nil || step < 1 {
			return fmt.Errorf("invalid step value %q", stepParts[1])
		}
	}

	if rangePart == "*" {
		return nil
	}

	// Check for range (e.g., "1-5")
	if strings.Contains(rangePart, "-") {
		rangeBounds := strings.SplitN(rangePart, "-", 2)
		lo, err := strconv.Atoi(rangeBounds[0])
		if err != nil {
			return fmt.Errorf("invalid range start %q", rangeBounds[0])
		}
		hi, err := strconv.Atoi(rangeBounds[1])
		if err != nil {
			return fmt.Errorf("invalid range end %q", rangeBounds[1])
		}
		if lo < min || lo > max || hi < min || hi > max {
			return fmt.Errorf("value out of range [%d-%d]", min, max)
		}
		if lo > hi {
			return fmt.Errorf("range start %d is greater than end %d", lo, hi)
		}
		return nil
	}

	// Single value
	val, err := strconv.Atoi(rangePart)
	if err != nil {
		return fmt.Errorf("invalid value %q", rangePart)
	}
	if val < min || val > max {
		return fmt.Errorf("value %d out of range [%d-%d]", val, min, max)
	}
	return nil
}

// semverRangeRegex matches common semver range operators.
// Supports: >=1.0.0, <2.0.0, ^1.2.3, ~1.2.3, =1.0.0, 1.0.0, >=1.0.0 <2.0.0
var semverPartRegex = regexp.MustCompile(
	`^([~^]|>=?|<=?|=)?` +
		`(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)` +
		`(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
		`(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`,
)

func validateSemverRanges(m *Manifest) []ValidationError {
	var errs []ValidationError
	for i, dep := range m.DependsOn {
		if dep.Version == nil {
			continue
		}
		ver := *dep.Version
		if !isValidSemverRange(ver) {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("depends_on[%d].version", i),
				Message: fmt.Sprintf("invalid semver range: %q", ver),
				Level:   "error",
			})
		}
	}
	return errs
}

func isValidSemverRange(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	// A semver range can be space-separated comparators (AND) or || separated (OR)
	orParts := strings.Split(s, "||")
	for _, orPart := range orParts {
		orPart = strings.TrimSpace(orPart)
		if orPart == "" {
			return false
		}
		// Each OR part is space-separated comparators
		comparators := strings.Fields(orPart)
		for _, comp := range comparators {
			if !semverPartRegex.MatchString(comp) {
				return false
			}
		}
	}
	return true
}

func validateBCP47(m *Manifest) []ValidationError {
	var errs []ValidationError

	// Validate locale field
	if m.Locale != nil && *m.Locale != "" {
		if _, err := language.Parse(*m.Locale); err != nil {
			errs = append(errs, ValidationError{
				Field:   "locale",
				Message: fmt.Sprintf("invalid BCP-47 language tag: %q", *m.Locale),
				Level:   "error",
			})
		}
	}

	// Validate locales keys
	for key := range m.Locales {
		if _, err := language.Parse(key); err != nil {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("locales.%s", key),
				Message: fmt.Sprintf("invalid BCP-47 language tag: %q", key),
				Level:   "error",
			})
		}
	}

	return errs
}

func validateMCPToolInputSchemas(m *Manifest) []ValidationError {
	if m.MCP == nil {
		return nil
	}

	var errs []ValidationError
	for i, tool := range m.MCP.Tools {
		if tool.InputSchema == nil {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("mcp.tools[%d].inputSchema", i),
				Message: "inputSchema is required",
				Level:   "error",
			})
			continue
		}

		// The inputSchema must be a JSON object (map) with a "type" field.
		schemaMap, ok := toMap(tool.InputSchema)
		if !ok {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("mcp.tools[%d].inputSchema", i),
				Message: "inputSchema must be a valid JSON Schema object",
				Level:   "error",
			})
			continue
		}

		if _, hasType := schemaMap["type"]; !hasType {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("mcp.tools[%d].inputSchema", i),
				Message: "inputSchema must contain a \"type\" field",
				Level:   "error",
			})
		}
	}
	return errs
}

// toMap attempts to convert an any value to map[string]any.
// Handles both map[string]any (from JSON unmarshal) and struct types
// by marshaling/unmarshaling through JSON.
func toMap(v any) (map[string]any, bool) {
	if m, ok := v.(map[string]any); ok {
		return m, true
	}
	// Try JSON round-trip for other types
	data, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, false
	}
	return m, true
}
