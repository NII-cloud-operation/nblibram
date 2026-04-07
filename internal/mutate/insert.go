package mutate

import (
	"encoding/json"
	"errors"
	"flag"

	nb "github.com/nii-cloud/nblibram/internal/notebook"
)

func RunInsert(args []string) error {
	fs := flag.NewFlagSet("insert", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	cellType := fs.String("type", "code", "cell type: code or markdown")
	source := fs.String("source", "", "cell source content")
	sourceFile := fs.String("source-file", "", "read cell source from file")
	position := fs.String("position", "after", "insert position: before or after")
	inPlace := fs.Bool("i", false, "modify file in place")
	var queryFlags nb.MultiFlag
	fs.Var(&queryFlags, "query", "insertion point ("+nb.QueryUsage+")")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *cellType != "code" && *cellType != "markdown" {
		return errors.New("--type must be code or markdown")
	}
	if *position != "before" && *position != "after" {
		return errors.New("--position must be before or after")
	}

	cellSource, err := resolveSource(*source, *sourceFile)
	if err != nil {
		return err
	}

	notebook, err := nb.Read(*file)
	if err != nil {
		return err
	}

	newCell := nb.NewCell(*cellType, cellSource)
	newCellRaw, err := nb.MarshalNewCellRaw(newCell)
	if err != nil {
		return err
	}

	if len(queryFlags) == 0 {
		notebook.Cells = append(notebook.Cells, newCell)
		notebook.CellsRaw = append(notebook.CellsRaw, newCellRaw)
	} else {
		filter, err := nb.ParseQueryFlags(queryFlags)
		if err != nil {
			return err
		}
		idx, err := nb.LocateStartCell(notebook, filter)
		if err != nil {
			return err
		}
		insertIdx := idx
		if *position == "after" {
			insertIdx = idx + 1
		}
		// insertIdx can be len(Cells) when inserting after the last cell
		cells := make([]nb.Cell, 0, len(notebook.Cells)+1)
		cells = append(cells, notebook.Cells[:insertIdx]...)
		cells = append(cells, newCell)
		cells = append(cells, notebook.Cells[insertIdx:]...)
		notebook.Cells = cells

		raws := make([]json.RawMessage, 0, len(notebook.CellsRaw)+1)
		raws = append(raws, notebook.CellsRaw[:insertIdx]...)
		raws = append(raws, newCellRaw)
		raws = append(raws, notebook.CellsRaw[insertIdx:]...)
		notebook.CellsRaw = raws
	}

	return notebook.Write(*file, *inPlace)
}
