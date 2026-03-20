package outputs

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

const testFile = "../../testdata/basic.ipynb"

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

func TestOutputsText(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-query", "start:1", "-format", "text", "-no-filter"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "/home/user") {
		t.Errorf("expected '/home/user', got %q", out)
	}
}

func TestOutputsJSON(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-query", "start:1", "-format", "json", "-no-filter"}); err != nil {
			t.Fatal(err)
		}
	})
	var results []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 output, got %d", len(results))
	}
	if results[0]["output_type"] != "stream" {
		t.Errorf("expected stream, got %v", results[0]["output_type"])
	}
}

func TestOutputsNoOutputError(t *testing.T) {
	err := Run([]string{"-file", testFile, "-query", "start:0", "-format", "text", "-no-filter"})
	if err == nil {
		t.Error("expected error for cell with no outputs")
	}
}

func TestOutputsMissingQuery(t *testing.T) {
	err := Run([]string{"-file", testFile, "-format", "text"})
	if err == nil {
		t.Error("expected error when no query provided")
	}
}

func TestOutputsRawRequiresMime(t *testing.T) {
	err := Run([]string{"-file", testFile, "-query", "start:1", "-format", "raw", "-no-filter"})
	if err == nil {
		t.Error("expected error when raw format without --mime")
	}
}

func TestParseOutputFormat(t *testing.T) {
	tests := []struct {
		format string
		mime   string
		kind   string
		err    bool
	}{
		{"text", "", "text", false},
		{"json", "", "json", false},
		{"raw", "image/png", "raw", false},
		{"raw", "", "", true},
		{"invalid", "", "", true},
	}
	for _, tt := range tests {
		f, err := parseOutputFormat(tt.format, tt.mime)
		if tt.err && err == nil {
			t.Errorf("parseOutputFormat(%q, %q): expected error", tt.format, tt.mime)
		}
		if !tt.err && err != nil {
			t.Errorf("parseOutputFormat(%q, %q): %v", tt.format, tt.mime, err)
		}
		if !tt.err && f.kind != tt.kind {
			t.Errorf("parseOutputFormat(%q, %q): kind=%s, want %s", tt.format, tt.mime, f.kind, tt.kind)
		}
	}
}

func TestOutputsDefaultFilterSanitizes(t *testing.T) {
	// Cell 1 output contains "/home/user" which is clean,
	// but verify filter runs without error in default mode.
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-query", "start:1", "-format", "text"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "/home/user") {
		t.Errorf("expected clean output, got %q", out)
	}
}

func TestOutputsSecondCell(t *testing.T) {
	// Cell 3 has output "3\n"
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-query", "start:3", "-format", "text", "-no-filter"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "3") {
		t.Errorf("expected '3' in output, got %q", out)
	}
}
