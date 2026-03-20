package cells

import (
	"bytes"
	"os"
	"strings"
	"testing"

	nb "github.com/nii-cloud/nblibram/internal/notebook"
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

func TestRunCellsCount(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-query", "start:0", "-count", "2", "-format", "md", "-no-filter"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "# Title") {
		t.Error("expected title cell")
	}
	if !strings.Contains(out, "import os") {
		t.Error("expected code cell")
	}
}

func TestRunCellsSets(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-query", "start:0", "-sets", "1", "-format", "md", "-no-filter"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "# Title") {
		t.Error("expected title cell")
	}
}

func TestRunCellsJSON(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-query", "start:0", "-count", "3", "-format", "json", "-no-filter"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, `"cells"`) {
		t.Error("expected JSON output")
	}
}

func TestRunCellsDefaultFilterSanitizesSecrets(t *testing.T) {
	// Cell 5 in basic.ipynb contains a GitHub token and an IP address.
	// Without --no-filter, Run must sanitize both by default.
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-query", "start:5", "-count", "1", "-format", "md"}); err != nil {
			t.Fatal(err)
		}
	})
	if strings.Contains(out, "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123") {
		t.Fatal("SECURITY: GitHub token was NOT sanitized in default mode")
	}
	if strings.Contains(out, "10.0.0.5") {
		t.Fatal("SECURITY: IPv4 address was NOT sanitized in default mode")
	}
}

func TestRunCellsNoFilterExposesSecrets(t *testing.T) {
	// With --no-filter, secrets should appear as-is.
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-query", "start:5", "-count", "1", "-format", "md", "-no-filter"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123") {
		t.Error("with --no-filter, secret should be present in output")
	}
}

func TestRunCellsDefaultFilterJSON(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-query", "start:5", "-count", "1", "-format", "json"}); err != nil {
			t.Fatal(err)
		}
	})
	if strings.Contains(out, "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123") {
		t.Fatal("SECURITY: secret was NOT sanitized in JSON output")
	}
}

func TestRunCellsDefaultFilterPython(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-query", "start:5", "-count", "1", "-format", "py"}); err != nil {
			t.Fatal(err)
		}
	})
	if strings.Contains(out, "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123") {
		t.Fatal("SECURITY: secret was NOT sanitized in Python output")
	}
}

func TestCollectConsecutive(t *testing.T) {
	notebook, err := nb.Read(testFile)
	if err != nil {
		t.Fatal(err)
	}
	sections, err := collectConsecutive(notebook, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 1 {
		t.Fatalf("expected 1 block, got %d", len(sections))
	}
	if len(sections[0].Cells) != 3 {
		t.Errorf("expected 3 cells, got %d", len(sections[0].Cells))
	}
}

func TestCollectCellSets(t *testing.T) {
	notebook, err := nb.Read(testFile)
	if err != nil {
		t.Fatal(err)
	}
	sections, err := CollectCellSets(notebook, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 1 {
		t.Fatalf("expected 1 set, got %d", len(sections))
	}
	// First set: markdown "Title" + code cell
	if len(sections[0].Cells) != 2 {
		t.Errorf("expected 2 cells in set, got %d", len(sections[0].Cells))
	}
}
