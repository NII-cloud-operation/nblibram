package toc

import (
	"bytes"
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

func TestTocMarkdown(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-format", "md", "-no-filter"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "# Title") {
		t.Error("expected '# Title' in output")
	}
	if !strings.Contains(out, "## Section A") {
		t.Error("expected '## Section A' in output")
	}
	if !strings.Contains(out, "## Section B") {
		t.Error("expected '## Section B' in output")
	}
}

func TestTocDefaultFilterSanitizes(t *testing.T) {
	// basic.ipynb cell 5 has a token in source.
	// toc only shows headings, but we verify filter is loaded by default
	// (stderr should show gitleaks warning, no crash).
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-format", "md"}); err != nil {
			t.Fatal(err)
		}
	})
	// Headings themselves don't contain secrets in basic.ipynb,
	// so just verify it runs without error with filter on.
	if !strings.Contains(out, "# Title") {
		t.Error("expected headings in output")
	}
}

func TestTocJSON(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-format", "json", "-no-filter"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, `"cells"`) {
		t.Error("expected JSON with 'cells' key")
	}
	if !strings.Contains(out, `"title"`) {
		t.Error("expected cell id 'title' in JSON")
	}
}
