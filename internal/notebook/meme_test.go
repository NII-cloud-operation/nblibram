package notebook

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGetMemeID(t *testing.T) {
	nb := loadTestNotebook(t)

	if id := GetMemeID(nb.Cells[0]); id != "" {
		t.Errorf("cell 0: expected empty meme, got %q", id)
	}
	if id := GetMemeID(nb.Cells[2]); id != "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" {
		t.Errorf("cell 2: got %q", id)
	}
	if id := GetMemeID(nb.Cells[4]); id != "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11-1-abcd" {
		t.Errorf("cell 4: got %q", id)
	}
}

func TestGenerateMEME(t *testing.T) {
	id := GenerateMEME()
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		t.Errorf("expected 5-segment UUID, got %d: %q", len(parts), id)
	}
}

func TestPatchCellMeme(t *testing.T) {
	nb := loadTestNotebook(t)

	// Cell 2 has existing lc_cell_meme
	raw := nb.CellsRaw[2]
	newMeme := map[string]any{"current": "new-uuid-here", "previous": nil, "next": nil}
	patched, err := PatchCellMeme(raw, newMeme)
	if err != nil {
		t.Fatal(err)
	}

	var cell Cell
	json.Unmarshal(patched, &cell)
	if GetMemeID(cell) != "new-uuid-here" {
		t.Errorf("meme not patched: %q", GetMemeID(cell))
	}

	// Verify key order preserved
	origFields, _ := parseOrderedObject(raw)
	patchFields, _ := parseOrderedObject(patched)
	for i := range origFields {
		if origFields[i].Key != patchFields[i].Key {
			t.Errorf("key order changed at %d: %s -> %s", i, origFields[i].Key, patchFields[i].Key)
		}
	}

	// Cell 0 has no lc_cell_meme — test adding new
	patched0, err := PatchCellMeme(nb.CellsRaw[0], map[string]any{"current": "fresh-uuid"})
	if err != nil {
		t.Fatal(err)
	}
	var cell0 Cell
	json.Unmarshal(patched0, &cell0)
	if GetMemeID(cell0) != "fresh-uuid" {
		t.Errorf("new meme not set: %q", GetMemeID(cell0))
	}
}

func TestUpdateNeighborMemes(t *testing.T) {
	nb := loadTestNotebook(t)

	// Give cell 1 a MEME so we can test prev/next
	nb.Cells[1].Metadata["lc_cell_meme"] = map[string]any{"current": "cell1-meme"}
	nb.CellsRaw[1], _ = PatchCellMeme(nb.CellsRaw[1], map[string]any{"current": "cell1-meme"})

	// Update cell 1 and its neighbors (0 has no MEME, 2 has)
	UpdateNeighborMemes(nb, []int{1, 2})

	meme1 := GetCellMeme(nb.Cells[1])
	if meme1["previous"] != nil {
		t.Errorf("cell 1 previous: want nil (cell 0 has no MEME), got %v", meme1["previous"])
	}
	if meme1["next"] != "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" {
		t.Errorf("cell 1 next: want cell 2's MEME, got %v", meme1["next"])
	}

	meme2 := GetCellMeme(nb.Cells[2])
	if meme2["previous"] != "cell1-meme" {
		t.Errorf("cell 2 previous: want cell1-meme, got %v", meme2["previous"])
	}

	// Verify CellsRaw was patched
	var c1 Cell
	json.Unmarshal(nb.CellsRaw[1], &c1)
	m1 := GetCellMeme(c1)
	if m1["next"] != "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" {
		t.Errorf("CellsRaw[1] not patched")
	}
}

func TestUpdateNeighborMemesHistory(t *testing.T) {
	nb := loadTestNotebook(t)

	// Cell 2 has MEME with no prev/next yet
	// Set initial prev/next
	meme2 := GetCellMeme(nb.Cells[2])
	meme2["previous"] = "old-prev"
	meme2["next"] = "old-next"
	nb.Cells[2].Metadata["lc_cell_meme"] = meme2
	nb.CellsRaw[2], _ = PatchCellMeme(nb.CellsRaw[2], meme2)

	// Now update — prev/next will change, should record history
	UpdateNeighborMemes(nb, []int{2})

	updated := GetCellMeme(nb.Cells[2])
	history, ok := updated["history"].([]any)
	if !ok || len(history) == 0 {
		t.Fatal("expected history entry")
	}
	entry := history[0].(map[string]any)
	if entry["previous"] != "old-prev" {
		t.Errorf("history previous: %v", entry["previous"])
	}
	if entry["next"] != "old-next" {
		t.Errorf("history next: %v", entry["next"])
	}
}
