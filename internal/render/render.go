package render

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	nb "github.com/nii-cloud/nblibram/internal/notebook"
)

type SectionBlock struct {
	Cells []nb.Cell
}

type Options struct {
	ExcludeOutputs bool
}

func Sections(format string, sections []SectionBlock, opts Options) error {
	switch format {
	case "md":
		PrintSectionsMarkdown(sections)
	case "json":
		cells := FlattenSectionCells(sections)
		if opts.ExcludeOutputs {
			cells = nb.ExcludeOutputs(cells)
		}
		payload := map[string]any{
			"cells": cells,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(payload); err != nil {
			return err
		}
	case "py":
		PrintSectionsPython(sections)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
	return nil
}

func PrintSectionsMarkdown(sections []SectionBlock) {
	for idx, section := range sections {
		if idx > 0 {
			fmt.Println("---")
		}
		for _, c := range section.Cells {
			switch c.CellType {
			case "markdown":
				text := nb.CellText(c)
				fmt.Print(text)
				if !strings.HasSuffix(text, "\n") {
					fmt.Println()
				}
				fmt.Println()
			case "code":
				fmt.Println("```")
				code := nb.CellText(c)
				fmt.Print(code)
				if !strings.HasSuffix(code, "\n") {
					fmt.Println()
				}
				fmt.Println("```")
				fmt.Println()
			default:
				fmt.Print(nb.CellText(c))
				fmt.Println()
			}
		}
	}
}

func PrintSectionsPython(sections []SectionBlock) {
	for idx, section := range sections {
		if idx > 0 {
			fmt.Println("# ---")
		}
		for _, c := range section.Cells {
			switch c.CellType {
			case "markdown":
				WriteMarkdownAsComments(c)
			case "code":
				code := nb.CellText(c)
				fmt.Print(code)
				if !strings.HasSuffix(code, "\n") {
					fmt.Println()
				}
				fmt.Println()
			default:
				fmt.Printf("# %s\n\n", nb.CellText(c))
			}
		}
	}
}

func WriteMarkdownAsComments(c nb.Cell) {
	for _, line := range c.Source {
		line = strings.TrimRight(line, "\n")
		if strings.TrimSpace(line) == "" {
			fmt.Println("#")
			continue
		}
		fmt.Printf("# %s\n", line)
	}
	fmt.Println()
}

func PrintHeadingsMarkdown(headings []nb.Heading) {
	for _, h := range headings {
		fmt.Printf("%s %s\n", strings.Repeat("#", h.Level), h.Title)
		fmt.Println()
		if h.Preview != "" {
			fmt.Printf("%s\n", h.Preview)
		}
		fmt.Println()
	}
}

func FlattenSectionCells(sections []SectionBlock) []nb.Cell {
	var cells []nb.Cell
	for _, section := range sections {
		cells = append(cells, section.Cells...)
	}
	return cells
}
