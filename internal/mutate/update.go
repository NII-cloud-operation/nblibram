package mutate

import (
	"errors"
	"flag"
	"fmt"

	nb "github.com/nii-cloud/nblibram/internal/notebook"
)

func RunUpdate(args []string) error {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	source := fs.String("source", "", "new cell source content")
	sourceFile := fs.String("source-file", "", "read new cell source from file")
	hash := fs.String("hash", "", "expected cell hash (required, optimistic lock)")
	inPlace := fs.Bool("i", false, "modify file in place")
	var queryFlags nb.MultiFlag
	fs.Var(&queryFlags, "query", "cell to update ("+nb.QueryUsage+")")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if len(queryFlags) == 0 {
		return errors.New("update requires at least one --query")
	}
	if *hash == "" {
		return errors.New("update requires --hash for optimistic locking")
	}

	cellSource, err := resolveSource(*source, *sourceFile)
	if err != nil {
		return err
	}

	notebook, err := nb.Read(*file)
	if err != nil {
		return err
	}

	filter, err := nb.ParseQueryFlags(queryFlags)
	if err != nil {
		return err
	}

	idx, err := nb.LocateStartCell(notebook, filter)
	if err != nil {
		return err
	}

	cell := &notebook.Cells[idx]
	currentHash := nb.ComputeCellHash(cell.CellType, nb.CellText(*cell))
	if currentHash != *hash {
		return fmt.Errorf("hash mismatch: cell has been modified (expected %s, got %s)", *hash, currentHash)
	}

	newSource := nb.NBSource(nb.SplitSourceLines(cellSource))
	cell.Source = newSource

	patched, err := nb.PatchCellSource(notebook.CellsRaw[idx], newSource)
	if err != nil {
		return err
	}
	notebook.CellsRaw[idx] = patched

	return notebook.Write(*file, *inPlace)
}
