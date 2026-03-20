package filter

import (
	"github.com/zricethezav/gitleaks/v8/config"
	"github.com/zricethezav/gitleaks/v8/regexp"
)

// builtinRules are nblibram's built-in privacy rules, added on top of gitleaks defaults.
// These cover patterns common in notebook workflows that gitleaks doesn't detect by default.
var builtinRules = []config.Rule{
	{
		RuleID:      "ipv4-address",
		Description: "Detects IPv4 addresses",
		Regex:       regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`),
	},
	{
		RuleID:      "domain-name",
		Description: "Detects domain names",
		Regex:       regexp.MustCompile(`[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.(com|org|net|jp|io|dev|local|internal)`),
	},
}

func addBuiltinRules(cfg *config.Config) {
	for _, rule := range builtinRules {
		if _, exists := cfg.Rules[rule.RuleID]; exists {
			continue
		}
		cfg.Rules[rule.RuleID] = rule
		cfg.OrderedRules = append(cfg.OrderedRules, rule.RuleID)
	}
}
