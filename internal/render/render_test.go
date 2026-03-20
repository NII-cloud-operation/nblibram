package render

import (
	"bytes"
	"os"
	"strings"
	"testing"

	nb "github.com/nii-cloud/nblibram/internal/notebook"
)

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

func testSections() []SectionBlock {
	return []SectionBlock{
		{
			Cells: []nb.Cell{
				{CellType: "markdown", Source: nb.NBSource{"## Heading\n"}},
				{CellType: "code", Source: nb.NBSource{"x = 1\n"}},
			},
		},
	}
}

func TestSectionsMarkdown(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Sections("md", testSections(), Options{}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "## Heading") {
		t.Error("missing heading")
	}
	if !strings.Contains(out, "```") {
		t.Error("missing code fence")
	}
	if !strings.Contains(out, "x = 1") {
		t.Error("missing code")
	}
}

func TestSectionsPython(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Sections("py", testSections(), Options{}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "# ## Heading") {
		t.Error("markdown should be commented")
	}
	if !strings.Contains(out, "x = 1") {
		t.Error("missing code")
	}
}

func TestSectionsJSON(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Sections("json", testSections(), Options{}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, `"cells"`) {
		t.Error("missing cells key")
	}
}

func TestSectionsExcludeOutputs(t *testing.T) {
	sections := []SectionBlock{
		{
			Cells: []nb.Cell{
				{
					CellType: "code",
					Source:   nb.NBSource{"print(1)"},
					Outputs:  []nb.Output{{OutputType: "stream", Text: nb.NBSource{"1\n"}}},
				},
			},
		},
	}
	out := captureStdout(t, func() {
		if err := Sections("json", sections, Options{ExcludeOutputs: true}); err != nil {
			t.Fatal(err)
		}
	})
	if strings.Contains(out, "stream") {
		t.Error("outputs should be excluded")
	}
}

func TestPrintHeadingsMarkdown(t *testing.T) {
	headings := []nb.Heading{
		{Level: 1, Title: "Top", Preview: "intro text"},
		{Level: 2, Title: "Sub", Preview: ""},
	}
	out := captureStdout(t, func() {
		PrintHeadingsMarkdown(headings)
	})
	if !strings.Contains(out, "# Top") {
		t.Error("missing level 1 heading")
	}
	if !strings.Contains(out, "## Sub") {
		t.Error("missing level 2 heading")
	}
	if !strings.Contains(out, "intro text") {
		t.Error("missing preview")
	}
}
