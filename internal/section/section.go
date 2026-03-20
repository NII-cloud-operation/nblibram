package section

import (
	"errors"
	"flag"
	"fmt"

	"github.com/nii-cloud/nblibram/internal/filter"
	nb "github.com/nii-cloud/nblibram/internal/notebook"
	"github.com/nii-cloud/nblibram/internal/render"
)

func Run(args []string) error {
	fs := flag.NewFlagSet("section", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	sets := fs.Int("sets", 1, "number of consecutive sections to return")
	format := fs.String("format", "md", "output format: md, json, or py")
	excludeOutputs := fs.Bool("exclude-outputs", false, "exclude cell outputs from JSON output")
	noFilter := fs.Bool("no-filter", false, "disable privacy filters")
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

	f, err := nb.ParseQueryFlags(queryFlags)
	if err != nil {
		return err
	}

	startIdx, err := nb.LocateStartCell(notebook, f)
	if err != nil {
		return err
	}

	sections, err := CollectSections(notebook, startIdx, *sets)
	if err != nil {
		return err
	}

	if sanitizer := filter.LoadDefault(*noFilter); sanitizer != nil {
		for i := range sections {
			sanitizer.SanitizeCells(sections[i].Cells)
		}
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
