// Copyright (c) 2025 Feiyueyun Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package manifest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Parse decodes skill.json bytes into a Manifest struct.
// Unknown fields are preserved in ExtraMetadata without errors (forward compatible).
// On failure, returns an error with line and column information.
func Parse(data []byte) (*Manifest, error) {
	// 1. JSON syntax check: unmarshal into raw map
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		if synErr, ok := err.(*json.SyntaxError); ok {
			line, col := offsetToLineCol(data, synErr.Offset)
			return nil, fmt.Errorf("JSON syntax error at line %d, column %d: %s", line, col, synErr.Error())
		}
		return nil, err
	}

	// 2. schema_version MAJOR version check
	if sv, ok := raw["schema_version"].(string); ok {
		major, ok := parseSemverMajor(sv)
		if ok && major != 1 {
			return nil, fmt.Errorf("unsupported schema_version %q: this library supports MAJOR version 1", sv)
		}
	}

	// 3. Unmarshal into Manifest struct (known fields)
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	// 4. Collect unknown fields into ExtraMetadata
	knownFields := getKnownFieldNames()
	for k, v := range raw {
		if !knownFields[k] {
			if m.ExtraMetadata == nil {
				m.ExtraMetadata = make(map[string]any)
			}
			m.ExtraMetadata[k] = v
		}
	}

	return &m, nil
}

// offsetToLineCol converts a byte offset to line and column numbers (both 1-based).
func offsetToLineCol(data []byte, offset int64) (line, col int) {
	line = 1
	col = 1
	for i := int64(0); i < offset && i < int64(len(data)); i++ {
		if data[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

// parseSemverMajor extracts the MAJOR version number from a semver string.
// Returns the MAJOR version and whether parsing succeeded.
func parseSemverMajor(sv string) (int, bool) {
	// Take the part before the first '.' as MAJOR
	parts := strings.SplitN(sv, ".", 2)
	if len(parts) == 0 {
		return 0, false
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false
	}
	return major, true
}

// getKnownFieldNames uses reflection to get the set of all known JSON field names from the Manifest struct.
func getKnownFieldNames() map[string]bool {
	known := make(map[string]bool)
	t := reflect.TypeOf(Manifest{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		// Take the part before the comma as the JSON field name
		name := strings.SplitN(tag, ",", 2)[0]
		if name != "" {
			known[name] = true
		}
	}
	return known
}
