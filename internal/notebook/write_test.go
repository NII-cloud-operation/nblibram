package notebook

import (
	"bytes"
	"encoding/json"
	"os"
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

func TestRoundTrip(t *testing.T) {
	original, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	nb, err := Read(testFile)
	if err != nil {
		t.Fatal(err)
	}

	// Write to buffer via stdout capture
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	if err := nb.Write("", false); err != nil {
		t.Fatal(err)
	}
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.Bytes()

	// Normalize: json.Indent both for consistent comparison
	var origPretty, outPretty bytes.Buffer
	json.Indent(&origPretty, original, "", " ")
	origPretty.WriteByte('\n')
	json.Indent(&outPretty, output, "", " ")
	outPretty.WriteByte('\n')

	if !bytes.Equal(origPretty.Bytes(), outPretty.Bytes()) {
		t.Errorf("round-trip produced different output\n--- original ---\n%s\n--- output ---\n%s",
			origPretty.String(), outPretty.String())
	}
}

func TestDeletePreservesOtherCells(t *testing.T) {
	nb, err := Read(testFile)
	if err != nil {
		t.Fatal(err)
	}

	// Save raw JSON of cells 1-5 before deleting cell 0
	preserved := make([]json.RawMessage, len(nb.CellsRaw)-1)
	copy(preserved, nb.CellsRaw[1:])

	// Delete cell 0
	nb.Cells = nb.Cells[1:]
	nb.CellsRaw = nb.CellsRaw[1:]

	// Verify remaining cells are byte-for-byte identical
	for i, raw := range nb.CellsRaw {
		if !bytes.Equal(raw, preserved[i]) {
			t.Errorf("cell %d raw JSON changed after deleting cell 0", i+1)
		}
	}
}

func TestMarshalNewCellRaw(t *testing.T) {
	c := NewCell("code", "x = 1")
	raw, err := MarshalNewCellRaw(c)
	if err != nil {
		t.Fatal(err)
	}

	// Parse back and verify required fields
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatal(err)
	}

	required := []string{"cell_type", "metadata", "source", "outputs", "execution_count"}
	for _, key := range required {
		if _, ok := parsed[key]; !ok {
			t.Errorf("new code cell missing required field %q", key)
		}
	}

	// execution_count should be null
	if string(parsed["execution_count"]) != "null" {
		t.Errorf("execution_count should be null, got %s", parsed["execution_count"])
	}

	// outputs should be empty array
	if string(parsed["outputs"]) != "[]" {
		t.Errorf("outputs should be [], got %s", parsed["outputs"])
	}
}

func TestPatchCellSource(t *testing.T) {
	nb, err := Read(testFile)
	if err != nil {
		t.Fatal(err)
	}

	original := nb.CellsRaw[1] // code cell
	newSource := NBSource{"patched = True\n"}
	patched, err := PatchCellSource(original, newSource)
	if err != nil {
		t.Fatal(err)
	}

	// Parse both and compare: only source should differ
	origFields, _ := parseOrderedObject(original)
	patchFields, _ := parseOrderedObject(patched)

	if len(origFields) != len(patchFields) {
		t.Fatalf("field count changed: %d -> %d", len(origFields), len(patchFields))
	}

	for i := range origFields {
		if origFields[i].Key != patchFields[i].Key {
			t.Errorf("key order changed at position %d: %s -> %s", i, origFields[i].Key, patchFields[i].Key)
		}
		if origFields[i].Key != "source" {
			// Non-source fields should be identical
			if !bytes.Equal(origFields[i].Value, patchFields[i].Value) {
				t.Errorf("field %q changed unexpectedly", origFields[i].Key)
			}
		}
	}

	// Verify new source is correct
	var cell Cell
	json.Unmarshal(patched, &cell)
	if cell.Source[0] != "patched = True\n" {
		t.Errorf("source not patched: %q", cell.Source[0])
	}
}
