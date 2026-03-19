package toc

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	nb "github.com/nii-cloud/nblibram/internal/notebook"
	"github.com/nii-cloud/nblibram/internal/render"
)

func Run(args []string) error {
	fs := flag.NewFlagSet("toc", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	words := fs.Int("words", 20, "number of preview words")
	format := fs.String("format", "md", "output format: md or json")
	noOutputs := fs.Bool("exclude-outputs", false, "exclude cell outputs from JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	notebook, err := nb.Read(*file)
	if err != nil {
		return err
	}

	headings := nb.CollectHeadings(notebook, *words)
	if len(headings) == 0 {
		return nil
	}

	switch *format {
	case "md":
		render.PrintHeadingsMarkdown(headings)
	case "json":
		headingCells := nb.ExtractHeadingCells(notebook)
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
