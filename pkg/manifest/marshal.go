// Copyright (c) 2025 Feiyueyun Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package manifest

import (
	"encoding/json"
)

// Marshal serializes a Manifest into compact JSON bytes.
// Extension fields from ExtraMetadata are merged as top-level fields in the output.
func Marshal(m *Manifest) ([]byte, error) {
	merged, err := toMergedMap(m)
	if err != nil {
		return nil, err
	}
	return json.Marshal(merged)
}

// MarshalIndent serializes a Manifest into indented, human-readable JSON bytes.
// Extension fields from ExtraMetadata are merged as top-level fields in the output.
func MarshalIndent(m *Manifest, indent string) ([]byte, error) {
	merged, err := toMergedMap(m)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(merged, "", indent)
}

// toMergedMap serializes a Manifest to map[string]any and promotes ExtraMetadata
// extension fields to top-level keys (symmetric inverse of Parse).
func toMergedMap(m *Manifest) (map[string]any, error) {
	// 1. Marshal the Manifest to JSON bytes
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	// 2. Unmarshal into map[string]any
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	// 3. If ExtraMetadata exists, promote its entries to top-level fields
	if extra, ok := raw["extra_metadata"]; ok {
		// Remove the nested extra_metadata key
		delete(raw, "extra_metadata")

		// Merge each ExtraMetadata entry as a top-level field
		if extraMap, ok := extra.(map[string]any); ok {
			for k, v := range extraMap {
				raw[k] = v
			}
		}
	}

	return raw, nil
}
