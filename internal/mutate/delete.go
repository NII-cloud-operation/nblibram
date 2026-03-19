package mutate

import (
	"errors"
	"flag"
	"fmt"

	nb "github.com/nii-cloud/nblibram/internal/notebook"
)

func RunDelete(args []string) error {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	hash := fs.String("hash", "", "expected cell hash (required, optimistic lock)")
	inPlace := fs.Bool("i", false, "modify file in place")
	var queryFlags nb.MultiFlag
	fs.Var(&queryFlags, "query", "cell to delete ("+nb.QueryUsage+")")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if len(queryFlags) == 0 {
		return errors.New("delete requires at least one --query")
	}
	if *hash == "" {
		return errors.New("delete requires --hash for optimistic locking")
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

	cell := notebook.Cells[idx]
	currentHash := nb.ComputeCellHash(cell.CellType, nb.CellText(cell))
	if currentHash != *hash {
		return fmt.Errorf("hash mismatch: cell has been modified (expected %s, got %s)", *hash, currentHash)
	}

	notebook.Cells = append(notebook.Cells[:idx], notebook.Cells[idx+1:]...)

	return notebook.Write(*file, *inPlace)
}
