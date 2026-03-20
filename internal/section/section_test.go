package section

import (
	"bytes"
	"os"
	"strings"
	"testing"

	nb "github.com/nii-cloud/nblibram/internal/notebook"
	"github.com/nii-cloud/nblibram/internal/render"
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

func TestRunSection(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-query", "start:2", "-format", "md", "-no-filter"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "## Section A") {
		t.Error("expected '## Section A'")
	}
	if !strings.Contains(out, "x = 1") {
		t.Error("expected code from Section A")
	}
	if strings.Contains(out, "## Section B") {
		t.Error("Section B should not be included")
	}
}

func TestRunSectionDefaultFilterSanitizesSecrets(t *testing.T) {
	// Section B (cell 4-5) contains a GitHub token in cell 5.
	out := captureStdout(t, func() {
		if err := Run([]string{"-file", testFile, "-query", "start:4", "-format", "md"}); err != nil {
			t.Fatal(err)
		}
	})
	if strings.Contains(out, "ghp_R2D2c3POxWfYz9AqBmNjKlHgTsUvXw0123") {
		t.Fatal("SECURITY: secret was NOT sanitized in default mode")
	}
}

func TestCollectSections(t *testing.T) {
	notebook, err := nb.Read(testFile)
	if err != nil {
		t.Fatal(err)
	}
	sections, err := CollectSections(notebook, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
	cells1 := render.FlattenSectionCells(sections[:1])
	if len(cells1) != 2 {
		t.Errorf("section A: expected 2 cells, got %d", len(cells1))
	}
}
