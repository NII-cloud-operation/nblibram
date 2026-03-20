package notebook

import (
	"testing"
)

func TestSplitSourceLines(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", []string{""}},
		{"hello", []string{"hello"}},
		{"a\nb", []string{"a\n", "b"}},
		{"a\nb\n", []string{"a\n", "b\n", ""}},
		{"line1\nline2\nline3", []string{"line1\n", "line2\n", "line3"}},
	}
	for _, tt := range tests {
		got := SplitSourceLines(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("SplitSourceLines(%q): len=%d, want %d", tt.input, len(got), len(tt.want))
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("SplitSourceLines(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestNewCell(t *testing.T) {
	c := NewCell("code", "x = 1")
	if c.CellType != "code" {
		t.Errorf("expected code, got %s", c.CellType)
	}
	if len(c.Outputs) != 0 {
		t.Error("code cell should have empty outputs slice")
	}
	if c.Outputs == nil {
		t.Error("code cell outputs should be non-nil empty slice")
	}

	m := NewCell("markdown", "# Hello")
	if m.Outputs != nil {
		t.Error("markdown cell should have nil outputs")
	}
}
