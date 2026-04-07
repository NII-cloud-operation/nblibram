package notebook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func (nb *Notebook) Write(path string, inPlace bool) error {
	// Reconstruct top-level JSON: replace "cells" with current CellsRaw,
	// pass all other fields through unchanged.
	out := make([]orderedField, len(nb.rawTop))
	for i, f := range nb.rawTop {
		if f.Key == "cells" {
			out[i] = orderedField{
				Key:   "cells",
				Value: marshalRawArray(nb.CellsRaw),
			}
		} else {
			out[i] = f
		}
	}

	compact, err := marshalOrderedObject(out)
	if err != nil {
		return err
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, compact, "", " "); err != nil {
		return err
	}
	pretty.WriteByte('\n')

	if inPlace && path != "" {
		return os.WriteFile(path, pretty.Bytes(), 0644)
	}
	_, err = os.Stdout.Write(pretty.Bytes())
	return err
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

// MarshalNewCellRaw builds a nbformat-compliant raw JSON for a newly created cell.
func MarshalNewCellRaw(c Cell) (json.RawMessage, error) {
	fields := []orderedField{
		{Key: "cell_type", Value: mustMarshal(c.CellType)},
	}
	if c.ID != "" {
		fields = append(fields, orderedField{Key: "id", Value: mustMarshal(c.ID)})
	}
	fields = append(fields, orderedField{Key: "metadata", Value: mustMarshal(c.Metadata)})
	fields = append(fields, orderedField{Key: "source", Value: mustMarshal(c.Source)})
	if c.CellType == "code" {
		fields = append(fields, orderedField{Key: "outputs", Value: mustMarshal(c.Outputs)})
		fields = append(fields, orderedField{Key: "execution_count", Value: mustMarshal(c.ExecutionCount)})
	}
	return marshalOrderedObject(fields)
}

func mustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
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

// PatchCellSource replaces the "source" field in raw cell JSON, preserving all other fields and key order.
func PatchCellSource(raw json.RawMessage, newSource NBSource) (json.RawMessage, error) {
	fields, err := parseOrderedObject(raw)
	if err != nil {
		return nil, err
	}

	sourceJSON, err := json.Marshal(newSource)
	if err != nil {
		return nil, err
	}

	found := false
	for i, f := range fields {
		if f.Key == "source" {
			fields[i].Value = json.RawMessage(sourceJSON)
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("cell has no source field")
	}

	return marshalOrderedObject(fields)
}
