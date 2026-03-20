package notebook

import (
	"testing"
)

const testFile = "../../testdata/basic.ipynb"

func loadTestNotebook(t *testing.T) *Notebook {
	t.Helper()
	nb, err := Read(testFile)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	return nb
}

func TestRead(t *testing.T) {
	nb := loadTestNotebook(t)
	if len(nb.Cells) != 6 {
		t.Fatalf("expected 6 cells, got %d", len(nb.Cells))
	}
	if nb.Cells[0].CellType != "markdown" {
		t.Errorf("cell 0: expected markdown, got %s", nb.Cells[0].CellType)
	}
	if nb.Cells[1].CellType != "code" {
		t.Errorf("cell 1: expected code, got %s", nb.Cells[1].CellType)
	}
	if nb.Cells[0].ID != "title" {
		t.Errorf("cell 0: expected id 'title', got %s", nb.Cells[0].ID)
	}
	if nb.Cells[0].Index != 0 || nb.Cells[5].Index != 5 {
		t.Errorf("indices not set correctly")
	}
}

func TestCellText(t *testing.T) {
	nb := loadTestNotebook(t)
	text := CellText(nb.Cells[0])
	if text != "# Title\n\nIntroduction text." {
		t.Errorf("unexpected cell text: %q", text)
	}
}

func TestCollectHeadings(t *testing.T) {
	nb := loadTestNotebook(t)
	headings := CollectHeadings(nb, 5)
	if len(headings) != 3 {
		t.Fatalf("expected 3 headings, got %d", len(headings))
	}
	if headings[0].Level != 1 || headings[0].Title != "Title" {
		t.Errorf("heading 0: %+v", headings[0])
	}
	if headings[1].Level != 2 || headings[1].Title != "Section A" {
		t.Errorf("heading 1: %+v", headings[1])
	}
}

func TestSectionBounds(t *testing.T) {
	nb := loadTestNotebook(t)
	start, end, level, err := SectionBounds(nb, 0)
	if err != nil {
		t.Fatal(err)
	}
	if start != 0 || end != 6 || level != 1 {
		t.Errorf("expected 0..6 level 1, got %d..%d level %d", start, end, level)
	}

	start, end, level, err = SectionBounds(nb, 2)
	if err != nil {
		t.Fatal(err)
	}
	if start != 2 || end != 4 || level != 2 {
		t.Errorf("expected 2..4 level 2, got %d..%d level %d", start, end, level)
	}
}

func TestExtractHeadingCells(t *testing.T) {
	nb := loadTestNotebook(t)
	cells := ExtractHeadingCells(nb)
	if len(cells) != 3 {
		t.Fatalf("expected 3 heading cells, got %d", len(cells))
	}
}

func TestReadInvalidPath(t *testing.T) {
	_, err := Read("nonexistent.ipynb")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestCloneCells(t *testing.T) {
	nb := loadTestNotebook(t)
	cloned := CloneCells(nb.Cells[:2])
	if len(cloned) != 2 {
		t.Fatalf("expected 2 cloned cells, got %d", len(cloned))
	}
	cloned[0].CellType = "modified"
	if nb.Cells[0].CellType == "modified" {
		t.Error("clone modified original")
	}
}

func TestExcludeOutputs(t *testing.T) {
	nb := loadTestNotebook(t)
	stripped := ExcludeOutputs(nb.Cells[:2])
	if stripped[1].Outputs != nil {
		t.Error("outputs should be nil")
	}
	if nb.Cells[1].Outputs == nil {
		t.Error("original should not be modified")
	}
}

func TestFindNextPeerHeading(t *testing.T) {
	nb := loadTestNotebook(t)
	next := FindNextPeerHeading(nb, 3, 2)
	if next != 4 {
		t.Errorf("expected 4, got %d", next)
	}
	none := FindNextPeerHeading(nb, 5, 1)
	if none != -1 {
		t.Errorf("expected -1, got %d", none)
	}
}

func TestFirstHeadingLevel(t *testing.T) {
	nb := loadTestNotebook(t)
	level, ok := FirstHeadingLevel(nb.Cells[0])
	if !ok || level != 1 {
		t.Errorf("expected level 1, got %d, ok=%v", level, ok)
	}
	_, ok = FirstHeadingLevel(nb.Cells[1])
	if ok {
		t.Error("code cell should not have heading")
	}
}

func TestCountLeadingHashes(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"# H1", 1},
		{"## H2", 2},
		{"### H3", 3},
		{"no heading", 0},
		{"", 0},
	}
	for _, tt := range tests {
		got := CountLeadingHashes(tt.input)
		if got != tt.want {
			t.Errorf("CountLeadingHashes(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestPreviewFromLines(t *testing.T) {
	lines := []string{"hello world\n", "foo bar baz\n"}
	p := PreviewFromLines(lines, 3)
	if p != "hello world foo ..." {
		t.Errorf("unexpected preview: %q", p)
	}
	p = PreviewFromLines(lines, 0)
	if p != "" {
		t.Error("expected empty preview for limit=0")
	}
	p = PreviewFromLines(nil, 5)
	if p != "" {
		t.Error("expected empty preview for nil lines")
	}
}
