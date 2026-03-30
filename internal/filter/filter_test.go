package filter

import (
	"bytes"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	nb "github.com/nii-cloud/nblibram/internal/notebook"
)

func newTestSanitizer(t *testing.T) *Sanitizer {
	t.Helper()
	d, err := NewDetector("")
	if err != nil {
		t.Fatalf("NewDetector: %v", err)
	}
	return NewSanitizer(d)
}

func TestSanitizeDetectsToken(t *testing.T) {
	s := newTestSanitizer(t)
	input := `token = "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123"`
	result := s.Sanitize(input)
	if strings.Contains(result, "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123") {
		t.Error("secret was not sanitized")
	}
	if !strings.Contains(result, "[") {
		t.Error("expected replacement label in output")
	}
}

func TestSanitizeCleanText(t *testing.T) {
	s := newTestSanitizer(t)
	input := "print('hello world')"
	result := s.Sanitize(input)
	if result != input {
		t.Errorf("clean text was modified: %q", result)
	}
}

func TestSanitizeEquivalence(t *testing.T) {
	s := newTestSanitizer(t)
	secret := "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123"
	r1 := s.Sanitize(`a = "` + secret + `"`)
	r2 := s.Sanitize(`b = "` + secret + `"`)

	label1 := extractLabel(r1)
	label2 := extractLabel(r2)
	if label1 != label2 {
		t.Errorf("same secret got different labels: %s vs %s", label1, label2)
	}
}

func TestSanitizeIPv4(t *testing.T) {
	s := newTestSanitizer(t)
	input := `server = "10.0.0.5"`
	result := s.Sanitize(input)
	if strings.Contains(result, "10.0.0.5") {
		t.Fatal("SECURITY: IPv4 address was not sanitized")
	}
	if !strings.Contains(result, "[ipv4-address_") {
		t.Errorf("expected [ipv4-address_N] label, got %q", result)
	}
}

func TestSanitizeIPv4Equivalence(t *testing.T) {
	s := newTestSanitizer(t)
	r1 := s.Sanitize(`a = "10.0.0.5"`)
	r2 := s.Sanitize(`b = "192.168.1.100"`)
	r3 := s.Sanitize(`c = "10.0.0.5"`)

	l1 := extractLabel(r1)
	l2 := extractLabel(r2)
	l3 := extractLabel(r3)

	if l1 == l2 {
		t.Error("different IPs should get different labels")
	}
	if l1 != l3 {
		t.Errorf("same IP should get same label: %s vs %s", l1, l3)
	}
}

func TestSanitizeIPv4PrefixOverlap(t *testing.T) {
	s := newTestSanitizer(t)
	input := `hosts = ["192.168.0.1", "192.168.0.100", "192.168.0.1"]`
	result := s.Sanitize(input)

	// Both IPs must be sanitized
	if strings.Contains(result, "192.168.0.1") {
		t.Fatalf("SECURITY: IP address was not sanitized: %s", result)
	}

	// "192.168.0.100" must not become "[ipv4-address_N]00"
	if strings.Contains(result, "]00") {
		t.Fatalf("partial-match corruption: %s", result)
	}

	// Different IPs must get different labels
	labels := extractAllLabels(result)
	if len(labels) < 3 {
		t.Fatalf("expected at least 3 labels, got %d in %s", len(labels), result)
	}
	// labels[0] and labels[2] should match (same IP: 192.168.0.1)
	if labels[0] != labels[2] {
		t.Errorf("same IP got different labels: %s vs %s", labels[0], labels[2])
	}
	// labels[0] and labels[1] should differ (different IPs)
	if labels[0] == labels[1] {
		t.Errorf("different IPs got same label: %s", labels[0])
	}
}

func TestSanitizeDomain(t *testing.T) {
	s := newTestSanitizer(t)
	input := `url = "api.example.com"`
	result := s.Sanitize(input)
	if strings.Contains(result, "api.example.com") {
		t.Fatal("SECURITY: domain name was not sanitized")
	}
	if !strings.Contains(result, "[domain-name_") {
		t.Errorf("expected [domain-name_N] label, got %q", result)
	}
}

func TestSanitizeDifferentSecrets(t *testing.T) {
	s := newTestSanitizer(t)
	r1 := s.Sanitize(`a = "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123"`)
	r2 := s.Sanitize(`b = "ghp_X9Y8z7W6v5U4t3S2r1Q0p9O8n7M6l5K4j3I2"`)

	label1 := extractLabel(r1)
	label2 := extractLabel(r2)
	if label1 == label2 {
		t.Error("different secrets should get different labels")
	}
}

func TestSanitizeCells(t *testing.T) {
	s := newTestSanitizer(t)
	cells := []nb.Cell{
		{
			CellType: "code",
			Source:   nb.NBSource{`token = "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123"`},
		},
	}
	s.SanitizeCells(cells)
	if strings.Contains(cells[0].Source[0], "ghp_") {
		t.Error("cell source was not sanitized")
	}
}

func TestSanitizeCellsOutputData(t *testing.T) {
	s := newTestSanitizer(t)
	secret := `token = "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123"`
	cells := []nb.Cell{
		{
			CellType: "code",
			Source:   nb.NBSource{"print(1)"},
			Outputs: []nb.Output{
				{
					OutputType: "execute_result",
					Data: nb.OutputData{
						"text/plain": secret,
						"text/html":  []interface{}{secret},
					},
				},
			},
		},
	}
	s.SanitizeCells(cells)
	data := cells[0].Outputs[0].Data
	if str, ok := data["text/plain"].(string); ok && strings.Contains(str, "ghp_") {
		t.Fatal("SECURITY: output data text/plain was not sanitized")
	}
	if arr, ok := data["text/html"].([]interface{}); ok {
		if str, ok := arr[0].(string); ok && strings.Contains(str, "ghp_") {
			t.Fatal("SECURITY: output data text/html array was not sanitized")
		}
	}
}

func TestSanitizeCellsMetadata(t *testing.T) {
	s := newTestSanitizer(t)
	secret := `token = "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123"`
	cells := []nb.Cell{
		{
			CellType: "code",
			Source:   nb.NBSource{"x = 1"},
			Metadata: map[string]any{
				"note": secret,
				"nested": map[string]interface{}{
					"key": secret,
				},
			},
		},
	}
	s.SanitizeCells(cells)
	if str, ok := cells[0].Metadata["note"].(string); ok && strings.Contains(str, "ghp_") {
		t.Fatal("SECURITY: cell metadata was not sanitized")
	}
	nested := cells[0].Metadata["nested"].(map[string]interface{})
	if str, ok := nested["key"].(string); ok && strings.Contains(str, "ghp_") {
		t.Fatal("SECURITY: nested cell metadata was not sanitized")
	}
}

func TestSanitizeCellsOutputs(t *testing.T) {
	s := newTestSanitizer(t)
	// Use full assignment context so gitleaks keyword prefilter triggers
	secret := `token = "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123"`
	cells := []nb.Cell{
		{
			CellType: "code",
			Source:   nb.NBSource{"print(1)"},
			Outputs: []nb.Output{
				{
					OutputType: "stream",
					Text:       nb.NBSource{secret},
				},
				{
					OutputType: "error",
					Evalue:     secret,
					Traceback:  []string{secret},
				},
			},
		},
	}
	s.SanitizeCells(cells)
	if strings.Contains(cells[0].Outputs[0].Text[0], "ghp_") {
		t.Error("output text was not sanitized")
	}
	if strings.Contains(cells[0].Outputs[1].Evalue, "ghp_") {
		t.Error("evalue was not sanitized")
	}
	if strings.Contains(cells[0].Outputs[1].Traceback[0], "ghp_") {
		t.Error("traceback was not sanitized")
	}
}

func TestSanitizeHeadings(t *testing.T) {
	s := newTestSanitizer(t)
	headings := []nb.Heading{
		{Level: 1, Title: `Setup token = "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123"`, Preview: "some preview"},
	}
	s.SanitizeHeadings(headings)
	if strings.Contains(headings[0].Title, "ghp_") {
		t.Error("heading title was not sanitized")
	}
}

func TestLoadDefaultNoFilter(t *testing.T) {
	s := LoadDefault(true)
	if s != nil {
		t.Error("expected nil when noFilter=true")
	}
}

func TestLoadDefaultWithFilter(t *testing.T) {
	s := LoadDefault(false)
	if s == nil {
		t.Error("expected non-nil sanitizer")
	}
}

func TestNewDetectorDefault(t *testing.T) {
	d, err := NewDetector("")
	if err != nil {
		t.Fatal(err)
	}
	if d == nil {
		t.Error("expected non-nil detector")
	}
}

func TestFilterRunWithEmptyNotebook(t *testing.T) {
	nbJSON := `{"cells":[],"nbformat":4,"nbformat_minor":5}`
	old := os.Stdin
	r, w, _ := os.Pipe()
	w.WriteString(nbJSON)
	w.Close()
	os.Stdin = r

	out := captureStdout(t, func() {
		if err := Run([]string{}); err != nil {
			t.Fatal(err)
		}
	})

	os.Stdin = old

	if !strings.Contains(out, "cells") {
		t.Error("expected output with cells")
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestFilterRunWithSanitization(t *testing.T) {
	nbJSON := map[string]interface{}{
		"cells": []interface{}{
			map[string]interface{}{
				"cell_type": "code",
				"source":    []interface{}{`token = "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123"`},
				"outputs":   []interface{}{},
			},
		},
		"nbformat":       4,
		"nbformat_minor": 5,
	}
	data, _ := json.Marshal(nbJSON)

	old := os.Stdin
	r, w, _ := os.Pipe()
	w.Write(data)
	w.Close()
	os.Stdin = r

	out := captureStdout(t, func() {
		if err := Run([]string{}); err != nil {
			t.Fatal(err)
		}
	})

	os.Stdin = old

	if strings.Contains(out, "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123") {
		t.Error("secret was not filtered")
	}
}

func TestSanitizeRawNotebook(t *testing.T) {
	s := newTestSanitizer(t)
	secret := `token = "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123"`
	raw := map[string]interface{}{
		"metadata": map[string]interface{}{"info": secret},
		"cells": []interface{}{
			map[string]interface{}{
				"cell_type": "code",
				"source":    []interface{}{secret},
				"metadata":  map[string]interface{}{"tag": secret},
				"outputs": []interface{}{
					map[string]interface{}{
						"output_type": "stream",
						"text":        []interface{}{secret},
					},
					map[string]interface{}{
						"output_type": "error",
						"evalue":      secret,
						"traceback":   []interface{}{secret},
					},
					map[string]interface{}{
						"output_type": "execute_result",
						"data":        map[string]interface{}{"text/plain": secret},
					},
				},
			},
		},
	}
	sanitizeRawNotebook(s, raw)

	// Check notebook metadata
	md := raw["metadata"].(map[string]interface{})
	if strings.Contains(md["info"].(string), "ghp_") {
		t.Fatal("SECURITY: notebook metadata not sanitized")
	}

	cells := raw["cells"].([]interface{})
	cell := cells[0].(map[string]interface{})

	// Check source
	src := cell["source"].([]interface{})
	if strings.Contains(src[0].(string), "ghp_") {
		t.Fatal("SECURITY: cell source not sanitized")
	}

	// Check cell metadata
	cmeta := cell["metadata"].(map[string]interface{})
	if strings.Contains(cmeta["tag"].(string), "ghp_") {
		t.Fatal("SECURITY: cell metadata not sanitized")
	}

	// Check outputs
	outputs := cell["outputs"].([]interface{})

	// stream text
	stream := outputs[0].(map[string]interface{})
	stxt := stream["text"].([]interface{})
	if strings.Contains(stxt[0].(string), "ghp_") {
		t.Fatal("SECURITY: stream text not sanitized")
	}

	// error evalue + traceback
	errOut := outputs[1].(map[string]interface{})
	if strings.Contains(errOut["evalue"].(string), "ghp_") {
		t.Fatal("SECURITY: evalue not sanitized")
	}
	tb := errOut["traceback"].([]interface{})
	if strings.Contains(tb[0].(string), "ghp_") {
		t.Fatal("SECURITY: traceback not sanitized")
	}

	// execute_result data
	execOut := outputs[2].(map[string]interface{})
	data := execOut["data"].(map[string]interface{})
	if strings.Contains(data["text/plain"].(string), "ghp_") {
		t.Fatal("SECURITY: output data not sanitized")
	}
}

func TestRunFilterFile(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", "../../testdata/basic.ipynb"}); err != nil {
			t.Fatal(err)
		}
	})
	if strings.Contains(out, "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123") {
		t.Fatal("SECURITY: secret not sanitized in filter output")
	}
	if !strings.Contains(out, "cells") {
		t.Error("expected valid notebook JSON")
	}
}

func TestRunFilterInPlace(t *testing.T) {
	// Copy testdata to temp file
	data, err := os.ReadFile("../../testdata/basic.ipynb")
	if err != nil {
		t.Fatal(err)
	}
	tmp, err := os.CreateTemp("", "nblibram-test-*.ipynb")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.Write(data)
	tmp.Close()

	if err := Run([]string{"-file", tmp.Name(), "-i"}); err != nil {
		t.Fatal(err)
	}

	result, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(result), "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123") {
		t.Fatal("SECURITY: secret not sanitized in in-place filter")
	}
}

func extractLabel(s string) string {
	start := strings.Index(s, "[")
	end := strings.Index(s, "]")
	if start < 0 || end < 0 {
		return ""
	}
	return s[start : end+1]
}

func extractAllLabels(s string) []string {
	re := regexp.MustCompile(`\[[a-zA-Z0-9_-]+_\d+\]`)
	return re.FindAllString(s, -1)
}
