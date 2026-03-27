// Copyright (c) 2025 Feiyueyun Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package manifest provides parsing, validation, and serialization
// of Skill Manifest (skill.json) files.
//
// Core API:
//   - Parse(data []byte) (*Manifest, error) — parse skill.json bytes into a Manifest struct
//   - Validate(m *Manifest) []ValidationError — validate against JSON Schema + business rules
//   - Marshal(m *Manifest) ([]byte, error) — serialize to compact JSON
//   - MarshalIndent(m *Manifest, indent string) ([]byte, error) — serialize to indented JSON
package manifest
