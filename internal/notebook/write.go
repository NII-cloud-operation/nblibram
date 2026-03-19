package notebook

import (
	"encoding/json"
	"os"
	"strings"
)

func (nb *Notebook) Write(path string, inPlace bool) error {
	for i := range nb.Cells {
		nb.Cells[i].Index = 0
	}

	enc := json.NewEncoder(os.Stdout)
	if inPlace && path != "" {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		enc = json.NewEncoder(f)
	}
	enc.SetIndent("", " ")
	enc.SetEscapeHTML(false)
	return enc.Encode(nb)
}

func NewCell(cellType, source string) Cell {
	c := Cell{
		CellType: cellType,
		Source:   NBSource(SplitSourceLines(source)),
		Metadata: map[string]any{},
	}
	if cellType == "code" {
		c.Outputs = []Output{}
	}
	return c
}

func SplitSourceLines(source string) []string {
	if source == "" {
		return []string{""}
	}
	var lines []string
	remaining := source
	for {
		idx := strings.Index(remaining, "\n")
		if idx < 0 {
			lines = append(lines, remaining)
			break
		}
		lines = append(lines, remaining[:idx+1])
		remaining = remaining[idx+1:]
	}
	return lines
}
