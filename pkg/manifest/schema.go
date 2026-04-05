// Copyright (c) 2025 Feiyueyun Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package manifest

import (
	"fmt"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// schemaJSON contains the raw JSON Schema for Skill Manifest v1.
// Embedded as a string constant because Go's //go:embed directive
// does not support ".." paths, and the schema file lives at
// spec/skill-manifest-v1.schema.json (outside this package directory).
//
//nolint:lll
var schemaJSON = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://github.com/feiyueyun/skill-manifest-spec/spec/skill-manifest-v1.schema.json",
  "title": "Skill Manifest v1",
  "description": "FYY Skill Manifest v1 — Static skill descriptor file format",
  "type": "object",
  "required": [
    "name",
    "version",
    "description",
    "type"
  ],
  "additionalProperties": true,
  "properties": {
    "$schema": {
      "type": "string",
      "description": "JSON Schema reference URI"
    },
    "schema_version": {
      "type": "string",
      "const": "1.0.0",
      "description": "Skill Manifest standard version"
    },
    "name": {
      "type": "string",
      "pattern": "^[a-z][a-z0-9-]{1,63}$",
      "description": "Unique skill identifier, kebab-case, 2-64 characters"
    },
    "version": {
      "type": "string",
      "pattern": "^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$",
      "description": "Skill version, Semver format"
    },
    "description": {
      "type": "string",
      "minLength": 10,
      "description": "Skill description, minimum 10 characters"
    },
    "type": {
      "type": "string",
      "enum": [
        "claw",
        "service",
        "mcp"
      ],
      "description": "Skill type"
    },
    "entry_point": {
      "type": "string",
      "description": "Skill entry point file path"
    },
    "auto_start": {
      "type": "boolean",
      "default": false,
      "description": "Whether to auto-start when CLI daemon starts"
    },
    "listen_port": {
      "type": "integer",
      "minimum": 1024,
      "maximum": 65535,
      "description": "Preferred listen port (required when type=service)"
    },
    "protocol": {
      "type": "string",
      "enum": [
        "grpc",
        "json-rpc",
        "mcp"
      ],
      "description": "Communication protocol (required when type=service)"
    },
    "mcp": {
      "type": "object",
      "description": "MCP primitives declaration (required when type=mcp)",
      "properties": {
        "tools": {
          "type": "array",
          "items": {
            "type": "object",
            "required": [
              "name",
              "description",
              "inputSchema"
            ],
            "properties": {
              "name": {
                "type": "string"
              },
              "description": {
                "type": "string"
              },
              "inputSchema": {
                "type": "object",
                "description": "Input parameter definition in JSON Schema format"
              }
            },
            "additionalProperties": false
          }
        },
        "resources": {
          "type": "array",
          "items": {
            "type": "object",
            "required": [
              "uri",
              "name",
              "mimeType"
            ],
            "properties": {
              "uri": {
                "type": "string"
              },
              "name": {
                "type": "string"
              },
              "mimeType": {
                "type": "string"
              }
            },
            "additionalProperties": false
          }
        },
        "prompts": {
          "type": "array",
          "items": {
            "type": "object"
          }
        }
      },
      "additionalProperties": false
    },
    "mcp_mode": {
      "type": "string",
      "enum": [
        "local",
        "proxy"
      ],
      "default": "local",
      "description": "MCP runtime mode (only effective when type=mcp)"
    },
    "mcp_remote": {
      "type": "object",
      "description": "MCP proxy bridge configuration (required when mcp_mode=proxy)",
      "required": [
        "endpoint"
      ],
      "properties": {
        "endpoint": {
          "type": "string",
          "format": "uri",
          "pattern": "^https://"
        },
        "transport": {
          "type": "string",
          "enum": [
            "streamable-http",
            "sse"
          ],
          "default": "streamable-http"
        },
        "auth_type": {
          "type": "string",
          "enum": [
            "none",
            "api_key",
            "oauth2",
            "bearer"
          ]
        },
        "auth_ref": {
          "type": "string",
          "pattern": "^secret:"
        },
        "headers": {
          "type": "object",
          "additionalProperties": {
            "type": "string"
          }
        },
        "timeout_ms": {
          "type": "integer",
          "minimum": 1000,
          "maximum": 300000,
          "default": 30000
        },
        "cache_ttl_s": {
          "type": "integer",
          "minimum": 0,
          "maximum": 86400,
          "default": 0
        }
      },
      "additionalProperties": false
    },
    "openclaw_compatible": {
      "type": "boolean",
      "description": "Whether compatible with OpenClaw SKILL.md format"
    },
    "permissions": {
      "type": "array",
      "items": {
        "type": "string",
        "enum": [
          "network",
          "filesystem",
          "process",
          "env",
          "browser",
          "camera",
          "microphone",
          "payment",
          "credential",
          "system_exec",
          "irreversible"
        ]
      },
      "uniqueItems": true,
      "description": "Permission category list"
    },
    "network_endpoints": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Outbound network allowlist (domain or IP:port)"
    },
    "risk_level": {
      "type": "string",
      "enum": [
        "low",
        "medium",
        "high",
        "critical"
      ],
      "default": "low",
      "description": "Risk level"
    },
    "sandbox_override": {
      "type": "string",
      "enum": [
        "process",
        "seccomp",
        "container"
      ],
      "description": "Override automatic sandbox mapping"
    },
    "seccomp_profile": {
      "type": "string",
      "description": "Custom seccomp profile file path"
    },
    "allowed_networks": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Mesh networks the skill exposes services to"
    },
    "visibility": {
      "type": "string",
      "enum": [
        "private",
        "network",
        "public"
      ],
      "default": "private",
      "description": "Skill visibility"
    },
    "capabilities": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Capability list for Grant permission matching"
    },
    "when_to_use": {
      "type": "string",
      "description": "When to invoke this skill (core basis for Agent decision-making)"
    },
    "category": {
      "type": "string",
      "enum": [
        "ai-ml",
        "utility",
        "development",
        "productivity",
        "web",
        "science",
        "media",
        "social",
        "finance",
        "smart-home",
        "communication",
        "security",
        "data",
        "ecommerce",
        "logistics",
        "customer-service",
        "compliance",
        "legal",
        "healthcare",
        "other"
      ],
      "default": "other",
      "description": "Skill category"
    },
    "models": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Compatible AI model list (supports wildcards)"
    },
    "input_examples": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Typical input examples"
    },
    "output_examples": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Typical output examples"
    },
    "related_skills": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Functionally related or complementary skill names"
    },
    "role_affinity": {
      "type": "array",
      "items": {
        "type": "string",
        "pattern": "^[a-z][a-z0-9-]*$"
      },
      "description": "Suitable Agent team roles (kebab-case)"
    },
    "tags": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Free-form tags and prefixed tags"
    },
    "pricing": {
      "type": "object",
      "properties": {
        "model": {
          "type": "string",
          "enum": [
            "free",
            "per_call",
            "per_output",
            "per_minute",
            "per_token",
            "subscription"
          ]
        },
        "unit_price": {
          "type": "number",
          "minimum": 0
        },
        "currency": {
          "type": "string",
          "default": "credit"
        },
        "free_tier": {
          "type": "integer",
          "minimum": 0,
          "default": 0
        }
      },
      "additionalProperties": false
    },
    "sla": {
      "type": "object",
      "properties": {
        "availability": {
          "type": "number",
          "minimum": 0,
          "maximum": 100
        },
        "max_response_ms": {
          "type": "integer",
          "exclusiveMinimum": 0
        },
        "max_concurrency": {
          "type": "integer",
          "exclusiveMinimum": 0
        }
      },
      "additionalProperties": false
    },
    "resellable": {
      "type": "boolean",
      "default": true,
      "description": "Whether resale is allowed"
    },
    "resale_commission": {
      "type": "number",
      "minimum": 0,
      "maximum": 1,
      "default": 0.3,
      "description": "Resale commission ratio"
    },
    "memory_access": {
      "type": "object",
      "properties": {
        "read_soul": {
          "type": "boolean",
          "default": false
        },
        "read_memory": {
          "type": "boolean",
          "default": false
        },
        "write_memory": {
          "type": "boolean",
          "default": false
        }
      },
      "additionalProperties": false
    },
    "schedule": {
      "type": "object",
      "required": [
        "cron"
      ],
      "properties": {
        "cron": {
          "type": "string",
          "description": "Standard 5-field cron expression"
        },
        "timezone": {
          "type": "string",
          "default": "UTC",
          "description": "IANA timezone identifier"
        },
        "enabled": {
          "type": "boolean",
          "default": true
        },
        "action": {
          "type": "string",
          "default": "invoke"
        },
        "params": {
          "type": "object"
        }
      },
      "additionalProperties": false
    },
    "remote_config": {
      "type": "boolean",
      "default": false,
      "description": "Whether remote config hot-reload is supported"
    },
    "depends_on": {
      "type": "array",
      "items": {
        "oneOf": [
          {
            "type": "string",
            "description": "Simple format: skill name only"
          },
          {
            "type": "object",
            "required": [
              "name"
            ],
            "properties": {
              "name": {
                "type": "string"
              },
              "version": {
                "type": "string",
                "description": "Semver range syntax"
              }
            },
            "additionalProperties": false
          }
        ]
      },
      "description": "Dependent skill list"
    },
    "conflicts_with": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Incompatible skill name list"
    },
    "min_cli_version": {
      "type": "string",
      "description": "Minimum fyy CLI version (Semver)"
    },
    "min_platform_version": {
      "type": "string",
      "description": "Minimum control plane version (Semver)"
    },
    "locale": {
      "type": "string",
      "default": "zh-CN",
      "description": "Primary language (BCP-47 format)"
    },
    "locales": {
      "type": "object",
      "additionalProperties": {
        "type": "object",
        "properties": {
          "description": {
            "type": "string"
          },
          "when_to_use": {
            "type": "string"
          },
          "input_examples": {
            "type": "array",
            "items": {
              "type": "string"
            }
          }
        },
        "additionalProperties": true
      },
      "description": "Multilingual descriptions"
    },
    "requirements": {
      "type": "object",
      "properties": {
        "runtime": {
          "type": "string"
        },
        "min_version": {
          "type": "string"
        },
        "dependencies": {
          "type": "array",
          "items": {
            "type": "string"
          }
        }
      },
      "additionalProperties": false
    },
    "resource_limits": {
      "type": "object",
      "properties": {
        "cpu": {
          "type": "string"
        },
        "memory": {
          "type": "string"
        }
      },
      "additionalProperties": false
    },
    "data_dir": {
      "type": "string",
      "description": "Data directory path"
    },
    "timeout": {
      "type": "integer",
      "description": "Default timeout in seconds"
    },
    "health_check": {
      "type": "object",
      "properties": {
        "endpoint": {
          "type": "string"
        },
        "interval": {
          "type": "integer",
          "default": 30
        },
        "timeout": {
          "type": "integer",
          "default": 5
        }
      },
      "additionalProperties": false
    },
    "interop": {
      "type": "object",
      "properties": {
        "agent_skills": {
          "type": "boolean",
          "default": false
        },
        "clawhub": {
          "type": "boolean",
          "default": false
        },
        "a2a_card": {
          "type": "boolean",
          "default": false
        }
      },
      "additionalProperties": false
    },
    "disclosure": {
      "type": "object",
      "properties": {
        "level_0": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "level_1": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "level_2": {
          "type": "array",
          "items": {
            "type": "string"
          }
        }
      },
      "additionalProperties": false
    },
    "ontology": {
      "type": "object",
      "properties": {
        "domain": {
          "type": "string"
        },
        "category": {
          "type": "string"
        },
        "skill_node": {
          "type": "string"
        }
      },
      "additionalProperties": false
    },
    "instructions_path": {
      "type": "string",
      "description": "Relative path to instructions.md"
    },
    "output_format": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Supported output format list (e.g. json/markdown/pdf/excel)"
    },
    "output_schema": {
      "type": "string",
      "format": "uri",
      "description": "JSON Schema URL for output data structure"
    },
    "quality_indicators": {
      "type": "object",
      "properties": {
        "verified": {
          "type": "boolean",
          "default": false
        },
        "security_audit": {
          "type": "string"
        },
        "quality_score": {
          "type": "number",
          "minimum": 0,
          "maximum": 5
        },
        "total_calls": {
          "type": "integer",
          "minimum": 0
        },
        "source_type": {
          "type": "string",
          "enum": [
            "manual",
            "github",
            "agent-skills",
            "clawhub"
          ]
        },
        "source_url": {
          "type": "string",
          "format": "uri"
        }
      },
      "additionalProperties": false,
      "description": "Skill quality and security indicators (some fields populated by platform at runtime)"
    },
    "extra_metadata": {
      "type": "object",
      "additionalProperties": true,
      "description": "Extension field storage"
    }
  },
  "allOf": [
    {
      "if": {
        "properties": {
          "type": {
            "const": "service"
          }
        },
        "required": [
          "type"
        ]
      },
      "then": {
        "required": [
          "listen_port",
          "protocol"
        ]
      }
    },
    {
      "if": {
        "properties": {
          "type": {
            "const": "mcp"
          }
        },
        "required": [
          "type"
        ]
      },
      "then": {
        "required": [
          "mcp"
        ]
      }
    },
    {
      "if": {
        "properties": {
          "mcp_mode": {
            "const": "proxy"
          }
        },
        "required": [
          "mcp_mode"
        ]
      },
      "then": {
        "required": [
          "mcp_remote"
        ]
      }
    },
    {
      "if": {
        "properties": {
          "permissions": {
            "contains": {
              "const": "network"
            }
          }
        },
        "required": [
          "permissions"
        ]
      },
      "then": {
        "required": [
          "network_endpoints"
        ]
      }
    }
  ]
}
`

// schemaResourceID is the $id used to register and compile the schema.
const schemaResourceID = "https://github.com/feiyueyun/skill-manifest-spec/spec/skill-manifest-v1.schema.json"

var (
	compiledSchema *jsonschema.Schema
	schemaOnce     sync.Once
	schemaErr      error
)

// loadSchema compiles and caches the JSON Schema for Skill Manifest v1.
// It uses sync.Once to ensure the schema is compiled exactly once,
// making it safe for concurrent use.
// Returns the compiled schema or an error if compilation fails.
func loadSchema() (*jsonschema.Schema, error) {
	schemaOnce.Do(func() {
		// Parse the raw JSON schema string into an any value.
		schemaDoc, err := jsonschema.UnmarshalJSON(strings.NewReader(schemaJSON))
		if err != nil {
			schemaErr = fmt.Errorf("failed to unmarshal embedded schema JSON: %w", err)
			return
		}

		// Create a compiler and register the schema as a resource.
		c := jsonschema.NewCompiler()
		if err := c.AddResource(schemaResourceID, schemaDoc); err != nil {
			schemaErr = fmt.Errorf("failed to add schema resource: %w", err)
			return
		}

		// Compile the schema.
		compiledSchema, schemaErr = c.Compile(schemaResourceID)
		if schemaErr != nil {
			schemaErr = fmt.Errorf("failed to compile schema: %w", schemaErr)
		}
	})
	return compiledSchema, schemaErr
}
