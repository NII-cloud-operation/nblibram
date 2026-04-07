package mutate

import (
	"bytes"
	"encoding/json"
	"os"
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

func TestInsert(t *testing.T) {
	out := captureStdout(t, func() {
		if err := RunInsert([]string{"-file", testFile, "-source", "new_cell = True", "-type", "code"}); err != nil {
			t.Fatal(err)
		}
	})
	var result nb.Notebook
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if len(result.Cells) != 7 {
		t.Errorf("expected 7 cells after insert, got %d", len(result.Cells))
	}
	last := result.Cells[len(result.Cells)-1]
	if last.Source[0] != "new_cell = True" {
		t.Errorf("last cell source = %q", last.Source[0])
	}
}

func TestInsertWithQuery(t *testing.T) {
	out := captureStdout(t, func() {
		if err := RunInsert([]string{"-file", testFile, "-query", "start:0", "-position", "after", "-source", "inserted", "-type", "code"}); err != nil {
			t.Fatal(err)
		}
	})
	var result nb.Notebook
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Cells) != 7 {
		t.Errorf("expected 7 cells, got %d", len(result.Cells))
	}
	if result.Cells[1].Source[0] != "inserted" {
		t.Errorf("cell 1 source = %q", result.Cells[1].Source[0])
	}
}

func TestUpdateWithHash(t *testing.T) {
	notebook, err := nb.Read(testFile)
	if err != nil {
		t.Fatal(err)
	}
	cell := notebook.Cells[3]
	hash := nb.ComputeCellHash(cell.CellType, nb.CellText(cell))

	out := captureStdout(t, func() {
		if err := RunUpdate([]string{"-file", testFile, "-query", "start:3", "-hash", hash, "-source", "updated = True"}); err != nil {
			t.Fatal(err)
		}
	})
	var result nb.Notebook
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatal(err)
	}
	if result.Cells[3].Source[0] != "updated = True" {
		t.Errorf("cell 3 source = %q", result.Cells[3].Source[0])
	}
}

func TestUpdateHashMismatch(t *testing.T) {
	err := RunUpdate([]string{"-file", testFile, "-query", "start:3", "-hash", "wronghash", "-source", "x"})
	if err == nil {
		t.Error("expected hash mismatch error")
	}
}

func TestDeleteWithHash(t *testing.T) {
	notebook, err := nb.Read(testFile)
	if err != nil {
		t.Fatal(err)
	}
	cell := notebook.Cells[0]
	hash := nb.ComputeCellHash(cell.CellType, nb.CellText(cell))

	out := captureStdout(t, func() {
		if err := RunDelete([]string{"-file", testFile, "-query", "start:0", "-hash", hash}); err != nil {
			t.Fatal(err)
		}
	})
	var result nb.Notebook
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Cells) != 5 {
		t.Errorf("expected 5 cells after delete, got %d", len(result.Cells))
	}
}

func TestDeleteHashMismatch(t *testing.T) {
	err := RunDelete([]string{"-file", testFile, "-query", "start:0", "-hash", "wronghash"})
	if err == nil {
		t.Error("expected hash mismatch error")
	}
}

func TestHash(t *testing.T) {
	out := captureStdout(t, func() {
		if err := RunHash([]string{"-file", testFile, "-query", "start:0"}); err != nil {
			t.Fatal(err)
		}
	})
	if !bytes.Contains([]byte(out), []byte("_hash")) {
		t.Error("expected _hash in output")
	}
}

func TestHashAll(t *testing.T) {
	out := captureStdout(t, func() {
		if err := RunHash([]string{"-file", testFile, "-all"}); err != nil {
			t.Fatal(err)
		}
	})
	var results []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 6 {
		t.Errorf("expected 6 hashes, got %d", len(results))
	}
}

func TestInsertAssignsMeme(t *testing.T) {
	out := captureStdout(t, func() {
		if err := RunInsert([]string{"-file", testFile, "-source", "x = 1", "-type", "code"}); err != nil {
			t.Fatal(err)
		}
	})
	var result nb.Notebook
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatal(err)
	}
	last := result.Cells[len(result.Cells)-1]
	if nb.GetMemeID(last) == "" {
		t.Error("inserted cell should have a MEME")
	}
}

func TestInsertNoMeme(t *testing.T) {
	out := captureStdout(t, func() {
		if err := RunInsert([]string{"-file", testFile, "-source", "x = 1", "-type", "code", "-no-meme"}); err != nil {
			t.Fatal(err)
		}
	})
	var result nb.Notebook
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatal(err)
	}
	last := result.Cells[len(result.Cells)-1]
	if nb.GetMemeID(last) != "" {
		t.Error("inserted cell should have no MEME with --no-meme")
	}
}

func TestInsertUpdatesPrevNext(t *testing.T) {
	// Insert between cell 2 (has MEME) and cell 3
	out := captureStdout(t, func() {
		if err := RunInsert([]string{"-file", testFile, "-query", "start:2", "-position", "after", "-source", "inserted", "-type", "markdown"}); err != nil {
			t.Fatal(err)
		}
	})
	var result nb.Notebook
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatal(err)
	}

	// New cell at index 3 should have prev=cell2's MEME
	newMeme := nb.GetCellMeme(result.Cells[3])
	if newMeme == nil {
		t.Fatal("new cell has no MEME")
	}
	if newMeme["previous"] != "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" {
		t.Errorf("new cell previous: %v", newMeme["previous"])
	}

	// Cell 2's next should point to new cell
	meme2 := nb.GetCellMeme(result.Cells[2])
	if meme2["next"] != nb.GetMemeID(result.Cells[3]) {
		t.Errorf("cell 2 next: %v, want %v", meme2["next"], nb.GetMemeID(result.Cells[3]))
	}
}

func TestDeleteUpdatesPrevNext(t *testing.T) {
	notebook, err := nb.Read(testFile)
	if err != nil {
		t.Fatal(err)
	}
	// Delete cell 3 (between cell 2 with MEME and cell 4 with MEME)
	cell := notebook.Cells[3]
	hash := nb.ComputeCellHash(cell.CellType, nb.CellText(cell))

	out := captureStdout(t, func() {
		if err := RunDelete([]string{"-file", testFile, "-query", "start:3", "-hash", hash}); err != nil {
			t.Fatal(err)
		}
	})
	var result nb.Notebook
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatal(err)
	}

	// After delete, cell 2 (idx 2) and cell 4 (now idx 3) are neighbors
	meme2 := nb.GetCellMeme(result.Cells[2])
	meme3 := nb.GetCellMeme(result.Cells[3])
	if meme2["next"] != nb.GetMemeID(result.Cells[3]) {
		t.Errorf("cell 2 next: %v, want %v", meme2["next"], nb.GetMemeID(result.Cells[3]))
	}
	if meme3["previous"] != nb.GetMemeID(result.Cells[2]) {
		t.Errorf("cell 3 previous: %v, want %v", meme3["previous"], nb.GetMemeID(result.Cells[2]))
	}
}
