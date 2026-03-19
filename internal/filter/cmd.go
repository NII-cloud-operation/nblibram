package filter

import (
	"flag"
)

func Run(args []string) error {
	fs := flag.NewFlagSet("filter", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	configPath := fs.String("config", "", "filter config path (default: ~/.nbfilterrc.toml)")
	gitleaksPath := fs.String("gitleaks", "", "gitleaks.toml to load additional rules from")
	inPlace := fs.Bool("i", false, "modify file in place")
	if err := fs.Parse(args); err != nil {
		return err
	}

	config, err := LoadConfig(*configPath)
	if err != nil {
		return err
	}

	allFilters := config.Filters

	if *gitleaksPath != "" {
		glRules, err := LoadGitleaksRules(*gitleaksPath)
		if err != nil {
			return err
		}
		allFilters = append(allFilters, glRules...)
	}

	sanitizer, err := NewSanitizer(allFilters)
	if err != nil {
		return err
	}

	return RunFilterFile(*file, sanitizer, *inPlace)
}
