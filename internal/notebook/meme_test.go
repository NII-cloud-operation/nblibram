package notebook

import (
	"testing"
)

func TestGetMemeID(t *testing.T) {
	nb := loadTestNotebook(t)

	if id := GetMemeID(nb.Cells[0]); id != "" {
		t.Errorf("cell 0: expected empty meme, got %q", id)
	}
	if id := GetMemeID(nb.Cells[2]); id != "aaaa-bbbb-cccc" {
		t.Errorf("cell 2: expected 'aaaa-bbbb-cccc', got %q", id)
	}
	if id := GetMemeID(nb.Cells[4]); id != "aaaa-bbbb-cccc-1-abcd" {
		t.Errorf("cell 4: expected 'aaaa-bbbb-cccc-1-abcd', got %q", id)
	}
}
