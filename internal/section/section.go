package section

import (
	"errors"
	"flag"
	"fmt"

	nb "github.com/nii-cloud/nblibram/internal/notebook"
	"github.com/nii-cloud/nblibram/internal/render"
)

func Run(args []string) error {
	fs := flag.NewFlagSet("section", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	sets := fs.Int("sets", 1, "number of consecutive sections to return")
	format := fs.String("format", "md", "output format: md, json, or py")
	excludeOutputs := fs.Bool("exclude-outputs", false, "exclude cell outputs from JSON output")
	var queryFlags nb.MultiFlag
	fs.Var(&queryFlags, "query", "section query ("+nb.QueryUsage+")")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(queryFlags) == 0 {
		return errors.New("section requires at least one --query")
	}
	if *sets <= 0 {
		return errors.New("--sets must be >= 1")
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

	sections, err := CollectSections(notebook, startIdx, *sets)
	if err != nil {
		return err
	}

	return render.Sections(*format, sections, render.Options{ExcludeOutputs: *excludeOutputs})
}

func CollectSections(notebook *nb.Notebook, startIdx, count int) ([]render.SectionBlock, error) {
	if startIdx < 0 || startIdx >= len(notebook.Cells) {
		return nil, fmt.Errorf("start index %d out of range", startIdx)
	}

	idx := startIdx
	var sections []render.SectionBlock
	var level int
	for len(sections) < count && idx < len(notebook.Cells) {
		secStart, secEnd, lvl, err := nb.SectionBounds(notebook, idx)
		if err != nil {
			return nil, err
		}
		level = lvl
		sectionCells := nb.CloneCells(notebook.Cells[secStart:secEnd])
		sections = append(sections, render.SectionBlock{Cells: sectionCells})
		nextIdx := nb.FindNextPeerHeading(notebook, secEnd, level)
		if nextIdx < 0 {
			break
		}
		idx = nextIdx
	}
	return sections, nil
}
