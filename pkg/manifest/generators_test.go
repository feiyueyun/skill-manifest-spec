// Copyright (c) 2025 Feiyueyun Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package manifest

import (
	"fmt"

	"pgregory.net/rapid"
)

// --- Helper generators ---

// genKebabName generates a valid kebab-case name matching ^[a-z][a-z0-9-]{1,20}$.
// Kept short for readability in test output.
func genKebabName(t *rapid.T) string {
	return rapid.StringMatching(`[a-z][a-z0-9-]{1,20}`).Draw(t, "kebab-name")
}

// genSemver generates a valid semver string (MAJOR.MINOR.PATCH).
func genSemver(t *rapid.T) string {
	major := rapid.IntRange(0, 20).Draw(t, "major")
	minor := rapid.IntRange(0, 50).Draw(t, "minor")
	patch := rapid.IntRange(0, 100).Draw(t, "patch")
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

// genDescription generates a description string of at least 10 characters.
func genDescription(t *rapid.T) string {
	// Generate a base string and ensure it's at least 10 chars.
	base := rapid.StringMatching(`[A-Za-z ]{10,60}`).Draw(t, "description")
	return base
}

// genType generates one of the valid type enum values.
func genType(t *rapid.T) string {
	return rapid.SampledFrom([]string{"claw", "service", "mcp"}).Draw(t, "type")
}

// genProtocol generates one of the valid protocol enum values.
func genProtocol(t *rapid.T) string {
	return rapid.SampledFrom([]string{"grpc", "json-rpc", "mcp"}).Draw(t, "protocol")
}

// genListenPort generates a valid listen port in range [1024, 65535].
func genListenPort(t *rapid.T) int {
	return rapid.IntRange(1024, 65535).Draw(t, "listen-port")
}

// genRiskLevel generates one of the valid risk_level enum values.
func genRiskLevel(t *rapid.T) string {
	return rapid.SampledFrom([]string{"low", "medium", "high", "critical"}).Draw(t, "risk-level")
}

// genCategory generates one of the valid category enum values.
func genCategory(t *rapid.T) string {
	return rapid.SampledFrom([]string{
		"ai-ml", "utility", "development", "productivity", "web",
		"science", "media", "social", "finance", "smart-home",
		"communication", "security", "data", "other",
	}).Draw(t, "category")
}

// genVisibility generates one of the valid visibility enum values.
func genVisibility(t *rapid.T) string {
	return rapid.SampledFrom([]string{"private", "network", "public"}).Draw(t, "visibility")
}

// allPermissions is the complete set of valid permission values.
var allPermissions = []string{
	"network", "filesystem", "process", "env", "browser",
	"camera", "microphone", "payment", "credential",
	"system_exec", "irreversible",
}

// genPermissions generates a unique subset of valid permissions.
func genPermissions(t *rapid.T) []string {
	// Generate a bitmask to select a unique subset.
	n := len(allPermissions)
	mask := rapid.IntRange(0, (1<<n)-1).Draw(t, "perm-mask")
	var perms []string
	for i := 0; i < n; i++ {
		if mask&(1<<i) != 0 {
			perms = append(perms, allPermissions[i])
		}
	}
	return perms
}

// genNetworkEndpoints generates a slice of network endpoint strings.
func genNetworkEndpoints(t *rapid.T) []string {
	count := rapid.IntRange(1, 3).Draw(t, "endpoint-count")
	endpoints := make([]string, count)
	for i := range endpoints {
		domain := rapid.StringMatching(`[a-z]{3,10}\.[a-z]{2,4}`).Draw(t, fmt.Sprintf("endpoint-%d", i))
		endpoints[i] = domain
	}
	return endpoints
}

// genMCPConfig generates a valid MCPConfig with at least one tool.
func genMCPConfig(t *rapid.T) *MCPConfig {
	toolCount := rapid.IntRange(1, 3).Draw(t, "tool-count")
	tools := make([]MCPTool, toolCount)
	for i := range tools {
		tools[i] = MCPTool{
			Name:        rapid.StringMatching(`[a-z][a-z0-9-]{2,15}`).Draw(t, fmt.Sprintf("tool-name-%d", i)),
			Description: genDescription(t),
			InputSchema: map[string]any{"type": "object"},
		}
	}
	cfg := &MCPConfig{Tools: tools}

	// Optionally add resources.
	if rapid.Bool().Draw(t, "has-resources") {
		resCount := rapid.IntRange(1, 2).Draw(t, "res-count")
		resources := make([]MCPResource, resCount)
		for i := range resources {
			resources[i] = MCPResource{
				URI:      fmt.Sprintf("file:///data/%s", rapid.StringMatching(`[a-z]{3,10}`).Draw(t, fmt.Sprintf("res-name-%d", i))),
				Name:     rapid.StringMatching(`[a-z]{3,10}`).Draw(t, fmt.Sprintf("res-display-%d", i)),
				MimeType: rapid.SampledFrom([]string{"application/json", "text/plain", "text/csv"}).Draw(t, fmt.Sprintf("mime-%d", i)),
			}
		}
		cfg.Resources = resources
	}
	return cfg
}

// genMCPRemote generates a valid MCPRemote with an https endpoint.
func genMCPRemote(t *rapid.T) *MCPRemote {
	domain := rapid.StringMatching(`[a-z]{3,10}\.[a-z]{2,4}`).Draw(t, "remote-domain")
	remote := &MCPRemote{
		Endpoint: fmt.Sprintf("https://%s/mcp", domain),
	}
	if rapid.Bool().Draw(t, "has-transport") {
		transport := rapid.SampledFrom([]string{"streamable-http", "sse"}).Draw(t, "transport")
		remote.Transport = &transport
	}
	if rapid.Bool().Draw(t, "has-auth-type") {
		authType := rapid.SampledFrom([]string{"none", "api_key", "oauth2", "bearer"}).Draw(t, "auth-type")
		remote.AuthType = &authType
	}
	if rapid.Bool().Draw(t, "has-timeout-ms") {
		timeoutMs := rapid.IntRange(1000, 300000).Draw(t, "timeout-ms")
		remote.TimeoutMs = &timeoutMs
	}
	if rapid.Bool().Draw(t, "has-cache-ttl") {
		cacheTTL := rapid.IntRange(0, 86400).Draw(t, "cache-ttl")
		remote.CacheTTLs = &cacheTTL
	}
	return remote
}

// genPricing generates a valid Pricing object.
// When model is not "free", unit_price is always included.
func genPricing(t *rapid.T) *Pricing {
	model := rapid.SampledFrom([]string{"free", "per_call", "per_minute", "per_token", "subscription"}).Draw(t, "pricing-model")
	p := &Pricing{Model: &model}

	if model != "free" {
		// Non-free models require unit_price.
		price := rapid.Float64Range(0.001, 100.0).Draw(t, "unit-price")
		p.UnitPrice = &price
	} else if rapid.Bool().Draw(t, "free-has-price") {
		// Free model: unit_price is optional.
		price := rapid.Float64Range(0.0, 0.0).Draw(t, "free-unit-price")
		p.UnitPrice = &price
	}

	if rapid.Bool().Draw(t, "has-currency") {
		currency := rapid.SampledFrom([]string{"credit", "USD", "CNY"}).Draw(t, "currency")
		p.Currency = &currency
	}
	if rapid.Bool().Draw(t, "has-free-tier") {
		freeTier := rapid.IntRange(0, 1000).Draw(t, "free-tier")
		p.FreeTier = &freeTier
	}
	return p
}

// genSchedule generates a valid Schedule with a 5-field cron expression.
func genSchedule(t *rapid.T) *Schedule {
	minute := rapid.IntRange(0, 59).Draw(t, "cron-min")
	hour := rapid.IntRange(0, 23).Draw(t, "cron-hour")
	dom := rapid.IntRange(1, 28).Draw(t, "cron-dom") // 1-28 to avoid invalid day-of-month
	month := rapid.IntRange(1, 12).Draw(t, "cron-month")
	dow := rapid.IntRange(0, 6).Draw(t, "cron-dow")

	// Mix of specific values and wildcards.
	cronFields := [5]string{
		fmt.Sprintf("%d", minute),
		fmt.Sprintf("%d", hour),
		fmt.Sprintf("%d", dom),
		fmt.Sprintf("%d", month),
		fmt.Sprintf("%d", dow),
	}
	// Randomly replace some fields with "*".
	for i := range cronFields {
		if rapid.Bool().Draw(t, fmt.Sprintf("cron-wildcard-%d", i)) {
			cronFields[i] = "*"
		}
	}
	cron := fmt.Sprintf("%s %s %s %s %s", cronFields[0], cronFields[1], cronFields[2], cronFields[3], cronFields[4])

	s := &Schedule{Cron: cron}
	if rapid.Bool().Draw(t, "has-timezone") {
		tz := rapid.SampledFrom([]string{"UTC", "Asia/Shanghai", "America/New_York", "Europe/London"}).Draw(t, "timezone")
		s.Timezone = &tz
	}
	if rapid.Bool().Draw(t, "has-enabled") {
		enabled := rapid.Bool().Draw(t, "schedule-enabled")
		s.Enabled = &enabled
	}
	return s
}

// genSLA generates a valid SLA object.
func genSLA(t *rapid.T) *SLA {
	sla := &SLA{}
	if rapid.Bool().Draw(t, "has-availability") {
		avail := rapid.Float64Range(90.0, 100.0).Draw(t, "availability")
		sla.Availability = &avail
	}
	if rapid.Bool().Draw(t, "has-max-response") {
		ms := rapid.IntRange(1, 60000).Draw(t, "max-response-ms")
		sla.MaxResponseMs = &ms
	}
	if rapid.Bool().Draw(t, "has-max-concurrency") {
		conc := rapid.IntRange(1, 1000).Draw(t, "max-concurrency")
		sla.MaxConcurrency = &conc
	}
	return sla
}

// genMemoryAccess generates a valid MemoryAccess object.
// Avoids write_memory=true + read_memory=false to prevent warnings
// (which are OK but we want clean valid manifests for Property 4).
func genMemoryAccess(t *rapid.T) *MemoryAccess {
	readSoul := rapid.Bool().Draw(t, "read-soul")
	readMemory := rapid.Bool().Draw(t, "read-memory")
	writeMemory := rapid.Bool().Draw(t, "write-memory")

	// Avoid the warn condition: write_memory=true + read_memory=false.
	if writeMemory && !readMemory {
		readMemory = true
	}

	return &MemoryAccess{
		ReadSoul:    &readSoul,
		ReadMemory:  &readMemory,
		WriteMemory: &writeMemory,
	}
}

// genLocale generates a valid BCP-47 locale string.
func genLocale(t *rapid.T) string {
	return rapid.SampledFrom([]string{
		"zh-CN", "en-US", "en-GB", "ja-JP", "ko-KR", "fr-FR", "de-DE", "es-ES",
	}).Draw(t, "locale")
}

// --- Main generator ---

// genValidManifest generates a valid Manifest satisfying all Schema constraints
// and conditional dependencies. The generated manifest should pass Validate()
// with no error-level errors.
func genValidManifest(t *rapid.T) *Manifest {
	skillType := genType(t)

	m := &Manifest{
		Name:        genKebabName(t),
		Version:     genSemver(t),
		Description: genDescription(t),
		Type:        skillType,
	}

	// --- Conditional dependencies based on type ---

	switch skillType {
	case "service":
		port := genListenPort(t)
		m.ListenPort = &port
		protocol := genProtocol(t)
		m.Protocol = &protocol

	case "mcp":
		m.MCP = genMCPConfig(t)

		// Optionally set mcp_mode.
		if rapid.Bool().Draw(t, "has-mcp-mode") {
			mode := rapid.SampledFrom([]string{"local", "proxy"}).Draw(t, "mcp-mode")
			m.MCPMode = &mode

			// If proxy, mcp_remote is required.
			if mode == "proxy" {
				m.MCPRemote = genMCPRemote(t)
			}
		}
	}

	// --- Optional fields (randomly present or absent) ---

	// schema_version
	if rapid.Bool().Draw(t, "has-schema-version") {
		sv := "1.0.0"
		m.SchemaVersion = &sv
	}

	// entry_point
	if rapid.Bool().Draw(t, "has-entry-point") {
		ep := rapid.StringMatching(`[a-z_/]{3,20}\.(py|js|go|ts)`).Draw(t, "entry-point")
		m.EntryPoint = &ep
	}

	// auto_start
	if rapid.Bool().Draw(t, "has-auto-start") {
		as := rapid.Bool().Draw(t, "auto-start")
		m.AutoStart = &as
	}

	// openclaw_compatible
	if rapid.Bool().Draw(t, "has-openclaw") {
		oc := rapid.Bool().Draw(t, "openclaw-compatible")
		m.OpenClawCompatible = &oc
	}

	// permissions + network_endpoints conditional dependency
	if rapid.Bool().Draw(t, "has-permissions") {
		perms := genPermissions(t)
		if len(perms) > 0 {
			m.Permissions = perms

			// If "network" is in permissions, network_endpoints is required.
			hasNetwork := false
			for _, p := range perms {
				if p == "network" {
					hasNetwork = true
					break
				}
			}
			if hasNetwork {
				m.NetworkEndpoints = genNetworkEndpoints(t)
			} else if rapid.Bool().Draw(t, "has-network-endpoints-anyway") {
				m.NetworkEndpoints = genNetworkEndpoints(t)
			}

			// Set risk_level appropriately to avoid warnings.
			// Check if any high-privilege permissions are present.
			hasHighPriv := false
			for _, p := range perms {
				if highPrivilegePermissions[p] {
					hasHighPriv = true
					break
				}
			}
			if hasHighPriv {
				// Set risk_level to "high" or "critical" to avoid warnings.
				rl := rapid.SampledFrom([]string{"high", "critical"}).Draw(t, "risk-level-high")
				m.RiskLevel = &rl
			} else if rapid.Bool().Draw(t, "has-risk-level") {
				rl := genRiskLevel(t)
				m.RiskLevel = &rl
			}
		}
	} else if rapid.Bool().Draw(t, "has-risk-level-no-perms") {
		rl := genRiskLevel(t)
		m.RiskLevel = &rl
	}

	// sandbox_override
	if rapid.Bool().Draw(t, "has-sandbox-override") {
		so := rapid.SampledFrom([]string{"process", "seccomp", "container"}).Draw(t, "sandbox-override")
		m.SandboxOverride = &so
	}

	// visibility
	if rapid.Bool().Draw(t, "has-visibility") {
		vis := genVisibility(t)
		m.Visibility = &vis
	}

	// category
	if rapid.Bool().Draw(t, "has-category") {
		cat := genCategory(t)
		m.Category = &cat
	}

	// when_to_use
	if rapid.Bool().Draw(t, "has-when-to-use") {
		wtu := rapid.StringMatching(`[A-Za-z ]{10,40}`).Draw(t, "when-to-use")
		m.WhenToUse = &wtu
	}

	// tags
	if rapid.Bool().Draw(t, "has-tags") {
		tagCount := rapid.IntRange(1, 5).Draw(t, "tag-count")
		tags := make([]string, tagCount)
		for i := range tags {
			tags[i] = rapid.StringMatching(`[a-z:]{2,15}`).Draw(t, fmt.Sprintf("tag-%d", i))
		}
		m.Tags = tags
	}

	// pricing
	if rapid.Bool().Draw(t, "has-pricing") {
		m.Pricing = genPricing(t)
	}

	// sla
	if rapid.Bool().Draw(t, "has-sla") {
		m.SLA = genSLA(t)
	}

	// resellable
	if rapid.Bool().Draw(t, "has-resellable") {
		r := rapid.Bool().Draw(t, "resellable")
		m.Resellable = &r
	}

	// resale_commission
	if rapid.Bool().Draw(t, "has-resale-commission") {
		rc := rapid.Float64Range(0.0, 1.0).Draw(t, "resale-commission")
		m.ResaleCommission = &rc
	}

	// memory_access
	if rapid.Bool().Draw(t, "has-memory-access") {
		m.MemoryAccess = genMemoryAccess(t)
	}

	// schedule
	if rapid.Bool().Draw(t, "has-schedule") {
		m.Schedule = genSchedule(t)
	}

	// remote_config
	if rapid.Bool().Draw(t, "has-remote-config") {
		rc := rapid.Bool().Draw(t, "remote-config")
		m.RemoteConfig = &rc
	}

	// locale
	if rapid.Bool().Draw(t, "has-locale") {
		loc := genLocale(t)
		m.Locale = &loc
	}

	// interop
	if rapid.Bool().Draw(t, "has-interop") {
		as := rapid.Bool().Draw(t, "agent-skills")
		ch := rapid.Bool().Draw(t, "clawhub")
		a2a := rapid.Bool().Draw(t, "a2a-card")
		m.Interop = &Interop{
			AgentSkills: &as,
			ClawHub:     &ch,
			A2ACard:     &a2a,
		}
	}

	return m
}
