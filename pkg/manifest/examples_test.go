// Copyright (c) 2025 Feiyueyun Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExamples_AllPassValidation loads all *.skill.json files from the
// examples/ directory, runs Parse + Validate on each, and ensures no
// error-level validation errors are produced.
func TestExamples_AllPassValidation(t *testing.T) {
	examplesDir := filepath.Join("..", "..", "examples")

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("failed to read examples directory %s: %v", examplesDir, err)
	}

	var skillFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".skill.json") {
			skillFiles = append(skillFiles, filepath.Join(examplesDir, entry.Name()))
		}
	}

	if len(skillFiles) == 0 {
		t.Fatal("no *.skill.json files found in examples/ directory")
	}

	for _, path := range skillFiles {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", path, err)
			}

			m, err := Parse(data)
			if err != nil {
				t.Fatalf("Parse(%s) failed: %v", name, err)
			}

			errs := Validate(m)
			var errorLevel []ValidationError
			for _, ve := range errs {
				if ve.Level == "error" {
					errorLevel = append(errorLevel, ve)
				}
			}

			if len(errorLevel) > 0 {
				t.Errorf("Validate(%s) produced %d error-level issues:", name, len(errorLevel))
				for _, ve := range errorLevel {
					t.Errorf("  [%s] %s: %s", ve.Level, ve.Field, ve.Message)
				}
			}
		})
	}
}

// TestSchema_Draft07Valid verifies that the embedded schemaJSON string
// in schema.go is itself valid JSON and a valid JSON Schema draft-07
// that can be compiled by the jsonschema library.
func TestSchema_Draft07Valid(t *testing.T) {
	// Verify schemaJSON is valid JSON.
	var raw any
	if err := json.Unmarshal([]byte(schemaJSON), &raw); err != nil {
		t.Fatalf("schemaJSON is not valid JSON: %v", err)
	}

	// Verify it can be compiled as a valid JSON Schema draft-07.
	// loadSchema() uses the jsonschema library to parse, register, and
	// compile the schema — if it returns no error, the schema is valid.
	schema, err := loadSchema()
	if err != nil {
		t.Fatalf("loadSchema() failed — schemaJSON is not a valid draft-07 schema: %v", err)
	}
	if schema == nil {
		t.Fatal("loadSchema() returned nil schema without error")
	}
}
