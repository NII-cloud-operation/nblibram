package cells

import (
	"errors"
	"flag"
	"fmt"

	nb "github.com/nii-cloud/nblibram/internal/notebook"
	"github.com/nii-cloud/nblibram/internal/render"
)

func Run(args []string) error {
	fs := flag.NewFlagSet("cells", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	sets := fs.Int("sets", 1, "number of Markdown+code pairs")
	count := fs.Int("count", 0, "number of consecutive cells (overrides --sets)")
	format := fs.String("format", "md", "output format: md, json, or py")
	excludeOutputs := fs.Bool("exclude-outputs", false, "exclude cell outputs from JSON output")
	var queryFlags nb.MultiFlag
	fs.Var(&queryFlags, "query", "cell query ("+nb.QueryUsage+")")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(queryFlags) == 0 {
		return errors.New("cells requires at least one --query")
	}

	notebook, err := nb.Read(*file)
	if err != nil {
		return err
	}

	filter, err := nb.ParseQueryFlags(queryFlags)
	if err != nil {
		return err
	}

	startIdx, err := nb.LocateStartCell(notebook, filter)
	if err != nil {
		return err
	}

	var sections []render.SectionBlock
	if *count > 0 {
		sections, err = collectConsecutive(notebook, startIdx, *count)
	} else {
		if *sets <= 0 {
			return errors.New("--sets must be >= 1")
		}
		sections, err = CollectCellSets(notebook, startIdx, *sets)
	}
	if err != nil {
		return err
	}

	return render.Sections(*format, sections, render.Options{ExcludeOutputs: *excludeOutputs})
}

func collectConsecutive(notebook *nb.Notebook, startIdx, count int) ([]render.SectionBlock, error) {
	if startIdx < 0 || startIdx >= len(notebook.Cells) {
		return nil, fmt.Errorf("start index %d out of range", startIdx)
	}
	end := startIdx + count
	if end > len(notebook.Cells) {
		end = len(notebook.Cells)
	}
	return []render.SectionBlock{
		{Cells: nb.CloneCells(notebook.Cells[startIdx:end])},
	}, nil
}

func CollectCellSets(notebook *nb.Notebook, startIdx, count int) ([]render.SectionBlock, error) {
	if startIdx < 0 || startIdx >= len(notebook.Cells) {
		return nil, fmt.Errorf("start index %d out of range", startIdx)
	}
	idx := startIdx
	var sections []render.SectionBlock
	for len(sections) < count {
		if idx >= len(notebook.Cells) {
			return nil, errors.New("not enough Markdown+code sets")
		}
		if notebook.Cells[idx].CellType != "markdown" {
			return nil, fmt.Errorf("cell %d is not a markdown cell", idx)
		}
		end := idx + 1
		for end < len(notebook.Cells) && notebook.Cells[end].CellType == "code" {
			end++
		}
		sections = append(sections, render.SectionBlock{Cells: nb.CloneCells(notebook.Cells[idx:end])})
		idx = end
	}
	return sections, nil
}
