package filter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type FilterConfig struct {
	Pattern   string   `toml:"pattern"`
	Label     string   `toml:"label"`
	Allowlist []string `toml:"-"` // compiled from gitleaks rules
}

type Filter struct {
	config    FilterConfig
	regex     *regexp.Regexp
	allowlist []*regexp.Regexp
	valueMap  map[string]int
	counter   int
}

func NewFilter(config FilterConfig) (*Filter, error) {
	regex, err := regexp.Compile(config.Pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern %q: %w", config.Pattern, err)
	}
	var allowlist []*regexp.Regexp
	for _, a := range config.Allowlist {
		re, err := regexp.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("invalid allowlist pattern %q: %w", a, err)
		}
		allowlist = append(allowlist, re)
	}
	return &Filter{
		config:    config,
		regex:     regex,
		allowlist: allowlist,
		valueMap:  make(map[string]int),
	}, nil
}

func (f *Filter) Replace(text string) string {
	return f.regex.ReplaceAllStringFunc(text, func(match string) string {
		for _, re := range f.allowlist {
			if re.MatchString(match) {
				return match
			}
		}
		if num, exists := f.valueMap[match]; exists {
			return f.formatLabel(num)
		}
		f.counter++
		f.valueMap[match] = f.counter
		return f.formatLabel(f.counter)
	})
}

func (f *Filter) formatLabel(num int) string {
	if strings.Contains(f.config.Label, "#") {
		return strings.Replace(f.config.Label, "#", fmt.Sprintf("%d", num), 1)
	}
	return f.config.Label
}

type Sanitizer struct {
	filters []*Filter
}

func NewSanitizer(configs []FilterConfig) (*Sanitizer, error) {
	filters := make([]*Filter, 0, len(configs))
	for _, fc := range configs {
		f, err := NewFilter(fc)
		if err != nil {
			return nil, err
		}
		filters = append(filters, f)
	}
	return &Sanitizer{filters: filters}, nil
}

func (s *Sanitizer) Sanitize(text string) string {
	for _, f := range s.filters {
		text = f.Replace(text)
	}
	return text
}

func (s *Sanitizer) SanitizeLines(lines []string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = s.Sanitize(line)
	}
	return result
}

func (s *Sanitizer) SanitizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		return s.Sanitize(val)
	case map[string]interface{}:
		for k, v := range val {
			val[k] = s.SanitizeValue(v)
		}
		return val
	case []interface{}:
		for i, v := range val {
			val[i] = s.SanitizeValue(v)
		}
		return val
	default:
		return v
	}
}

// SanitizeRawNotebook works on raw JSON to preserve source format (string vs []string).
func (s *Sanitizer) SanitizeRawNotebook(raw map[string]interface{}) {
	if md, ok := raw["metadata"]; ok {
		raw["metadata"] = s.SanitizeValue(md)
	}

	cellsRaw, ok := raw["cells"].([]interface{})
	if !ok {
		return
	}

	for _, cellRaw := range cellsRaw {
		cell, ok := cellRaw.(map[string]interface{})
		if !ok {
			continue
		}

		if src, ok := cell["source"]; ok {
			cell["source"] = s.sanitizeSource(src)
		}

		if md, ok := cell["metadata"]; ok {
			cell["metadata"] = s.SanitizeValue(md)
		}

		outputsRaw, ok := cell["outputs"].([]interface{})
		if !ok {
			continue
		}
		for _, outRaw := range outputsRaw {
			out, ok := outRaw.(map[string]interface{})
			if !ok {
				continue
			}
			if txt, ok := out["text"]; ok {
				out["text"] = s.sanitizeSource(txt)
			}
			if data, ok := out["data"]; ok {
				out["data"] = s.SanitizeValue(data)
			}
			if tb, ok := out["traceback"].([]interface{}); ok {
				for i, line := range tb {
					if str, ok := line.(string); ok {
						tb[i] = s.Sanitize(str)
					}
				}
			}
			if ev, ok := out["evalue"].(string); ok {
				out["evalue"] = s.Sanitize(ev)
			}
		}
	}
}

func (s *Sanitizer) sanitizeSource(src interface{}) interface{} {
	switch v := src.(type) {
	case string:
		return s.Sanitize(v)
	case []interface{}:
		for i, item := range v {
			if str, ok := item.(string); ok {
				v[i] = s.Sanitize(str)
			}
		}
		return v
	default:
		return src
	}
}

// RunFilter reads a notebook, applies filters, and writes the result.
func RunFilter(input io.Reader, output io.Writer, sanitizer *Sanitizer) error {
	var raw map[string]interface{}
	if err := json.NewDecoder(input).Decode(&raw); err != nil {
		return err
	}

	sanitizer.SanitizeRawNotebook(raw)

	enc := json.NewEncoder(output)
	enc.SetIndent("", " ")
	enc.SetEscapeHTML(false)
	return enc.Encode(raw)
}

// RunFilterFile is a convenience wrapper for file-based filtering.
func RunFilterFile(path string, sanitizer *Sanitizer, inPlace bool) error {
	var input io.Reader = os.Stdin
	if path != "" {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		input = f
	}

	if inPlace && path != "" {
		data, err := io.ReadAll(input)
		if err != nil {
			return err
		}

		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}

		sanitizer.SanitizeRawNotebook(raw)

		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		enc := json.NewEncoder(f)
		enc.SetIndent("", " ")
		enc.SetEscapeHTML(false)
		return enc.Encode(raw)
	}

	return RunFilter(input, os.Stdout, sanitizer)
}
