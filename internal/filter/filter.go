package filter

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/zricethezav/gitleaks/v8/config"
	"github.com/zricethezav/gitleaks/v8/detect"

	nb "github.com/nii-cloud/nblibram/internal/notebook"
)

const GitleaksEnvVar = "NBLIBRAM_GITLEAKS_CONFIG"

type Sanitizer struct {
	detector *detect.Detector
	valueMap map[string]replacement
	counter  map[string]int
}

type replacement struct {
	ruleID string
	num    int
}

func NewSanitizer(detector *detect.Detector) *Sanitizer {
	return &Sanitizer{
		detector: detector,
		valueMap: make(map[string]replacement),
		counter:  make(map[string]int),
	}
}

func (s *Sanitizer) Sanitize(text string) string {
	findings := s.detector.Detect(detect.Fragment{
		Raw:      text,
		FilePath: "notebook.ipynb",
	})
	if len(findings) == 0 {
		return text
	}

	result := text
	for _, f := range findings {
		if f.Secret == "" {
			continue
		}
		label := s.labelFor(f.Secret, f.RuleID)
		result = strings.ReplaceAll(result, f.Secret, label)
	}
	return result
}

func (s *Sanitizer) labelFor(secret, ruleID string) string {
	if r, exists := s.valueMap[secret]; exists {
		return fmt.Sprintf("[%s_%d]", r.ruleID, r.num)
	}
	s.counter[ruleID]++
	num := s.counter[ruleID]
	s.valueMap[secret] = replacement{ruleID: ruleID, num: num}
	return fmt.Sprintf("[%s_%d]", ruleID, num)
}

func (s *Sanitizer) SanitizeCells(cells []nb.Cell) {
	for i := range cells {
		for j, line := range cells[i].Source {
			cells[i].Source[j] = s.Sanitize(line)
		}
		for k := range cells[i].Outputs {
			out := &cells[i].Outputs[k]
			for j, line := range out.Text {
				out.Text[j] = s.Sanitize(line)
			}
			if out.Evalue != "" {
				out.Evalue = s.Sanitize(out.Evalue)
			}
			for j, line := range out.Traceback {
				out.Traceback[j] = s.Sanitize(line)
			}
		}
	}
}

func (s *Sanitizer) SanitizeHeadings(headings []nb.Heading) {
	for i := range headings {
		headings[i].Title = s.Sanitize(headings[i].Title)
		headings[i].Preview = s.Sanitize(headings[i].Preview)
	}
}

func LoadDefault(noFilter bool) *Sanitizer {
	if noFilter {
		return nil
	}

	path := os.Getenv(GitleaksEnvVar)
	if path == "" {
		fmt.Fprintf(os.Stderr, "nblibram: %s not set, using gitleaks default rules only\n", GitleaksEnvVar)
	}

	detector, err := NewDetector(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nblibram: gitleaks init: %v\n", err)
		return nil
	}

	return NewSanitizer(detector)
}

func NewDetector(configPath string) (*detect.Detector, error) {
	if configPath == "" {
		return detect.NewDetectorDefaultConfig()
	}

	// Read default config first
	viper.SetConfigType("toml")
	if err := viper.ReadConfig(strings.NewReader(config.DefaultConfig)); err != nil {
		return nil, err
	}
	// Merge custom config on top
	viper.SetConfigFile(configPath)
	if err := viper.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", configPath, err)
	}

	var vc config.ViperConfig
	if err := viper.Unmarshal(&vc); err != nil {
		return nil, err
	}
	cfg, err := vc.Translate()
	if err != nil {
		return nil, err
	}
	return detect.NewDetector(cfg), nil
}
