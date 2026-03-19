package filter

import (
	"fmt"
	"regexp"

	"github.com/BurntSushi/toml"
)

type gitleaksConfig struct {
	Rules []gitleaksRule `toml:"rules"`
}

type gitleaksRule struct {
	ID        string        `toml:"id"`
	Regex     string        `toml:"regex"`
	Allowlist gitleaksAllow `toml:"allowlist"`
}

type gitleaksAllow struct {
	Regexes []string `toml:"regexes"`
}

// LoadGitleaksRules parses a gitleaks.toml and converts rules to FilterConfigs.
// Each rule's id becomes the label (e.g. "email-address" → "[email-address_#]").
func LoadGitleaksRules(path string) ([]FilterConfig, error) {
	var cfg gitleaksConfig
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}

	var configs []FilterConfig
	for _, rule := range cfg.Rules {
		if rule.Regex == "" {
			continue
		}
		if _, err := regexp.Compile(rule.Regex); err != nil {
			return nil, fmt.Errorf("gitleaks rule %q: invalid regex: %w", rule.ID, err)
		}

		var allowRegexes []string
		for _, a := range rule.Allowlist.Regexes {
			if _, err := regexp.Compile(a); err != nil {
				return nil, fmt.Errorf("gitleaks rule %q allowlist: invalid regex: %w", rule.ID, err)
			}
			allowRegexes = append(allowRegexes, a)
		}

		label := fmt.Sprintf("[%s_#]", rule.ID)
		configs = append(configs, FilterConfig{
			Pattern:   rule.Regex,
			Label:     label,
			Allowlist: allowRegexes,
		})
	}

	return configs, nil
}
