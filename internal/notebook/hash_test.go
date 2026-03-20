package notebook

import (
	"testing"
)

func TestComputeCellHash(t *testing.T) {
	// Verified against TypeScript computeCellHash implementation
	tests := []struct {
		cellType string
		source   string
		want     string
	}{
		{"markdown", "# Hello", "af1c61a3"},
		{"code", "x = 1", "67ad8f1c"},
		{"code", "", "a8c2768"},
	}

	for _, tt := range tests {
		got := ComputeCellHash(tt.cellType, tt.source)
		if got != tt.want {
			t.Errorf("ComputeCellHash(%q, %q) = %s, want %s", tt.cellType, tt.source, got, tt.want)
		}
	}
}

func TestComputeCellHashConsistency(t *testing.T) {
	h1 := ComputeCellHash("code", "print('hello')")
	h2 := ComputeCellHash("code", "print('hello')")
	if h1 != h2 {
		t.Errorf("same input produced different hashes: %s vs %s", h1, h2)
	}
	h3 := ComputeCellHash("code", "print('world')")
	if h1 == h3 {
		t.Error("different input produced same hash")
	}
}
