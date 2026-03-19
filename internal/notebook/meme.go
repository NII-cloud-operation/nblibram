package notebook

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
