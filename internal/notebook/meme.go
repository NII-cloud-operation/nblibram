package notebook

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// GetMemeID extracts the lc_cell_meme.current UUID from cell metadata.
func GetMemeID(c Cell) string {
	meme, ok := c.Metadata["lc_cell_meme"]
	if !ok {
		return ""
	}
	memeMap, ok := meme.(map[string]interface{})
	if !ok {
		return ""
	}
	current, ok := memeMap["current"].(string)
	if !ok {
		return ""
	}
	return current
}

// GetCellMeme returns the full lc_cell_meme object from cell metadata.
func GetCellMeme(c Cell) map[string]any {
	meme, ok := c.Metadata["lc_cell_meme"]
	if !ok {
		return nil
	}
	memeMap, ok := meme.(map[string]interface{})
	if !ok {
		return nil
	}
	return memeMap
}

// GenerateMEME generates a new UUID v1 string for MEME identification.
func GenerateMEME() string {
	return uuid.Must(uuid.NewUUID()).String()
}

// PatchCellMeme patches the lc_cell_meme field in a cell's raw JSON metadata.
// It preserves all other metadata fields and their ordering.
func PatchCellMeme(raw json.RawMessage, memeData map[string]any) (json.RawMessage, error) {
	cellFields, err := parseOrderedObject(raw)
	if err != nil {
		return nil, err
	}

	metaIdx := -1
	for i, f := range cellFields {
		if f.Key == "metadata" {
			metaIdx = i
			break
		}
	}
	if metaIdx == -1 {
		return nil, fmt.Errorf("cell has no metadata field")
	}

	metaFields, err := parseOrderedObject(cellFields[metaIdx].Value)
	if err != nil {
		return nil, err
	}

	memeValue, err := json.Marshal(memeData)
	if err != nil {
		return nil, err
	}

	found := false
	for i, f := range metaFields {
		if f.Key == "lc_cell_meme" {
			metaFields[i].Value = json.RawMessage(memeValue)
			found = true
			break
		}
	}
	if !found {
		metaFields = append(metaFields, orderedField{Key: "lc_cell_meme", Value: json.RawMessage(memeValue)})
	}

	newMeta, err := marshalOrderedObject(metaFields)
	if err != nil {
		return nil, err
	}
	cellFields[metaIdx].Value = newMeta

	return marshalOrderedObject(cellFields)
}

// UpdateNeighborMemes updates prev/next links for cells at the given indices.
func UpdateNeighborMemes(nb *Notebook, indices []int) error {
	for _, i := range indices {
		c := &nb.Cells[i]
		memeObj := GetCellMeme(*c)
		if memeObj == nil {
			continue
		}

		var prevMeme, nextMeme any
		if i > 0 {
			prevMeme = GetMemeID(nb.Cells[i-1])
			if prevMeme == "" {
				prevMeme = nil
			}
		}
		if i < len(nb.Cells)-1 {
			nextMeme = GetMemeID(nb.Cells[i+1])
			if nextMeme == "" {
				nextMeme = nil
			}
		}

		oldPrev, _ := memeObj["previous"]
		oldNext, _ := memeObj["next"]

		if !memeValEqual(oldPrev, prevMeme) || !memeValEqual(oldNext, nextMeme) {
			// Record history if there was a previous state
			if oldPrev != nil || oldNext != nil {
				entry := map[string]any{
					"current":  memeObj["current"],
					"previous": oldPrev,
					"next":     oldNext,
				}
				history, _ := memeObj["history"].([]any)
				history = append(history, entry)
				memeObj["history"] = history
			}
		}

		memeObj["previous"] = prevMeme
		memeObj["next"] = nextMeme
		c.Metadata["lc_cell_meme"] = memeObj

		patched, err := PatchCellMeme(nb.CellsRaw[i], memeObj)
		if err != nil {
			return err
		}
		nb.CellsRaw[i] = patched
	}
	return nil
}

func memeValEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	aStr, aOk := a.(string)
	bStr, bOk := b.(string)
	if aOk && bOk {
		return aStr == bStr
	}
	return false
}
