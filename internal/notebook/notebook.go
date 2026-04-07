package notebook

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Notebook struct {
	// Parsed fields (read-only view for querying)
	Cells    []Cell
	Metadata map[string]any
	NBFormat int
	Minor    int

	// Raw JSON preservation (for writing)
	rawTop   []orderedField
	CellsRaw []json.RawMessage
}

type Cell struct {
	CellType       string         `json:"cell_type"`
	Source         NBSource       `json:"source"`
	ID             string         `json:"id,omitempty"`
	Outputs        []Output       `json:"outputs,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	ExecutionCount *int           `json:"execution_count,omitempty"`
	Index          int            `json:"_index,omitempty"`
	Hash           string         `json:"_hash,omitempty"`
}

type Output struct {
	Name       string     `json:"name,omitempty"`
	OutputType string     `json:"output_type"`
	Text       NBSource   `json:"text,omitempty"`
	Data       OutputData `json:"data,omitempty"`
	Stream     string     `json:"stream,omitempty"`
	Ename      string     `json:"ename,omitempty"`
	Evalue     string     `json:"evalue,omitempty"`
	Traceback  []string   `json:"traceback,omitempty"`
	Metadata   OutputMeta `json:"metadata,omitempty"`
}

type OutputData map[string]any

type OutputMeta map[string]any

type NBSource []string

func (s *NBSource) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		*s = nil
		return nil
	}
	switch data[0] {
	case '[':
		var arr []string
		if err := json.Unmarshal(data, &arr); err != nil {
			return err
		}
		*s = NBSource(arr)
	case '"':
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		*s = NBSource([]string{str})
	default:
		return fmt.Errorf("unsupported source encoding: %s", string(data))
	}
	return nil
}

type Heading struct {
	Level   int    `json:"level"`
	Title   string `json:"title"`
	Preview string `json:"preview"`
}

func Read(path string) (*Notebook, error) {
	var data []byte
	var err error
	if path != "" {
		data, err = os.ReadFile(path)
	} else {
		data, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("no notebook data")
	}

	topFields, err := parseOrderedObject(data)
	if err != nil {
		return nil, fmt.Errorf("parsing notebook: %w", err)
	}

	var nb Notebook
	nb.rawTop = topFields

	for _, f := range topFields {
		switch f.Key {
		case "cells":
			nb.CellsRaw, err = parseRawArray(f.Value)
			if err != nil {
				return nil, fmt.Errorf("parsing cells array: %w", err)
			}
			nb.Cells = make([]Cell, len(nb.CellsRaw))
			for i, raw := range nb.CellsRaw {
				if err := json.Unmarshal(raw, &nb.Cells[i]); err != nil {
					return nil, fmt.Errorf("parsing cell %d: %w", i, err)
				}
				nb.Cells[i].Index = i
			}
		case "metadata":
			if err := json.Unmarshal(f.Value, &nb.Metadata); err != nil {
				return nil, fmt.Errorf("parsing metadata: %w", err)
			}
		case "nbformat":
			if err := json.Unmarshal(f.Value, &nb.NBFormat); err != nil {
				return nil, fmt.Errorf("parsing nbformat: %w", err)
			}
		case "nbformat_minor":
			if err := json.Unmarshal(f.Value, &nb.Minor); err != nil {
				return nil, fmt.Errorf("parsing nbformat_minor: %w", err)
			}
		}
	}

	return &nb, nil
}

type RawField struct {
	Key   string
	Value json.RawMessage
}

func (nb *Notebook) RawTopFields() []RawField {
	out := make([]RawField, len(nb.rawTop))
	for i, f := range nb.rawTop {
		out[i] = RawField{Key: f.Key, Value: f.Value}
	}
	return out
}

func CollectHeadings(nb *Notebook, previewWords int) []Heading {
	var result []Heading
	for _, c := range nb.Cells {
		if c.CellType != "markdown" {
			continue
		}
		result = append(result, HeadingsFromCell(c, previewWords)...)
	}
	return result
}

func HeadingsFromCell(c Cell, previewWords int) []Heading {
	var hs []Heading
	lines := c.Source
	for idx, raw := range lines {
		trimmed := strings.TrimLeft(raw, " ")
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		level := CountLeadingHashes(trimmed)
		if level == 0 {
			continue
		}
		title := strings.TrimSpace(trimmed[level:])
		preview := PreviewFromLines(lines[idx+1:], previewWords)
		hs = append(hs, Heading{Level: level, Title: title, Preview: preview})
	}
	return hs
}

func CountLeadingHashes(s string) int {
	count := 0
	for _, r := range s {
		if r == '#' {
			count++
			continue
		}
		break
	}
	return count
}

func PreviewFromLines(lines []string, limit int) string {
	if limit <= 0 {
		return ""
	}
	var words []string
	truncated := false
	for _, line := range lines {
		for _, w := range strings.Fields(line) {
			words = append(words, w)
			if len(words) == limit {
				truncated = true
				break
			}
		}
		if truncated {
			break
		}
	}
	if len(words) == 0 {
		return ""
	}
	preview := strings.Join(words, " ")
	if truncated {
		preview += " ..."
	}
	return preview
}

func FirstHeadingLevel(c Cell) (int, bool) {
	if c.CellType != "markdown" {
		return 0, false
	}
	for _, raw := range c.Source {
		trimmed := strings.TrimLeft(raw, " ")
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		level := CountLeadingHashes(trimmed)
		if level == 0 {
			continue
		}
		return level, true
	}
	return 0, false
}

func SectionBounds(nb *Notebook, startIdx int) (int, int, int, error) {
	cell := nb.Cells[startIdx]
	level, ok := FirstHeadingLevel(cell)
	if !ok {
		return 0, 0, 0, fmt.Errorf("cell %d is not a markdown heading", startIdx)
	}
	end := len(nb.Cells)
	for i := startIdx + 1; i < len(nb.Cells); i++ {
		lvl, ok := FirstHeadingLevel(nb.Cells[i])
		if ok && lvl <= level {
			end = i
			break
		}
	}
	return startIdx, end, level, nil
}

func FindNextPeerHeading(nb *Notebook, start int, level int) int {
	for i := start; i < len(nb.Cells); i++ {
		lvl, ok := FirstHeadingLevel(nb.Cells[i])
		if !ok {
			continue
		}
		if lvl == level {
			return i
		}
	}
	return -1
}

func CellText(c Cell) string {
	return strings.Join(c.Source, "")
}

func CloneCells(cells []Cell) []Cell {
	res := make([]Cell, len(cells))
	copy(res, cells)
	return res
}

func ExcludeOutputs(cells []Cell) []Cell {
	res := make([]Cell, len(cells))
	for i, c := range cells {
		res[i] = c
		res[i].Outputs = nil
	}
	return res
}

func ExtractHeadingCells(nb *Notebook) []Cell {
	var cells []Cell
	for _, c := range nb.Cells {
		if c.CellType != "markdown" {
			continue
		}
		for _, line := range c.Source {
			if strings.HasPrefix(strings.TrimLeft(line, " "), "#") {
				cells = append(cells, c)
				break
			}
		}
	}
	return cells
}
