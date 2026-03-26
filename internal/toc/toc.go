package toc

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/nii-cloud/nblibram/internal/filter"
	nb "github.com/nii-cloud/nblibram/internal/notebook"
	"github.com/nii-cloud/nblibram/internal/render"
)

func Run(args []string) error {
	fs := flag.NewFlagSet("toc", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	words := fs.Int("words", 20, "number of preview words")
	format := fs.String("format", "md", "output format: md or json")
	noOutputs := fs.Bool("exclude-outputs", false, "exclude cell outputs from JSON output")
	noFilter := fs.Bool("no-filter", false, "disable privacy filters")
	if err := fs.Parse(args); err != nil {
		return err
	}

	notebook, err := nb.Read(*file)
	if err != nil {
		return err
	}

	sanitizer := filter.LoadDefault(*noFilter)

	headings := nb.CollectHeadings(notebook, *words)
	if len(headings) == 0 {
		return nil
	}

	switch *format {
	case "md":
		if sanitizer != nil {
			sanitizer.SanitizeHeadings(headings)
		}
		render.PrintHeadingsMarkdown(headings)
	case "json":
		headingCells := nb.ExtractHeadingCells(notebook)
		for i := range headingCells {
			headingCells[i].Hash = nb.ComputeCellHash(headingCells[i].CellType, nb.CellText(headingCells[i]))
		}
		if sanitizer != nil {
			sanitizer.SanitizeCells(headingCells)
		}
		if *noOutputs {
			headingCells = nb.ExcludeOutputs(headingCells)
		}
		payload := map[string]any{
			"cells": headingCells,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(payload); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported format: %s", *format)
	}

	return nil
}
