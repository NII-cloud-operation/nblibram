package mutate

// neighborsOf returns the indices of a cell and its immediate neighbors.
func neighborsOf(idx, length int) []int {
	var indices []int
	if idx > 0 {
		indices = append(indices, idx-1)
	}
	indices = append(indices, idx)
	if idx < length-1 {
		indices = append(indices, idx+1)
	}
	return indices
}

// neighborsOfDeleted returns the indices of cells adjacent to a deleted position.
func neighborsOfDeleted(deletedIdx, newLength int) []int {
	var indices []int
	prev := deletedIdx - 1
	next := deletedIdx // after removal, the cell at deletedIdx is the former next
	if prev >= 0 {
		indices = append(indices, prev)
	}
	if next < newLength {
		indices = append(indices, next)
	}
	return indices
}
