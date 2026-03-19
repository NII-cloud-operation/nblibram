package mutate

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	nb "github.com/nii-cloud/nblibram/internal/notebook"
)

func RunHash(args []string) error {
	fs := flag.NewFlagSet("hash", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	all := fs.Bool("all", false, "hash all cells")
	var queryFlags nb.MultiFlag
	fs.Var(&queryFlags, "query", "cell to hash ("+nb.QueryUsage+")")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if !*all && len(queryFlags) == 0 {
		return fmt.Errorf("hash requires --query or --all")
	}

	notebook, err := nb.Read(*file)
	if err != nil {
		return err
	}

	type cellHash struct {
		Index    int    `json:"_index"`
		ID       string `json:"id,omitempty"`
		CellType string `json:"cell_type"`
		Hash     string `json:"_hash"`
	}

	var results []cellHash

	if *all {
		for i, c := range notebook.Cells {
			results = append(results, cellHash{
				Index:    i,
				ID:       c.ID,
				CellType: c.CellType,
				Hash:     nb.ComputeCellHash(c.CellType, nb.CellText(c)),
			})
		}
	} else {
		filter, err := nb.ParseQueryFlags(queryFlags)
		if err != nil {
			return err
		}
		idx, err := nb.LocateStartCell(notebook, filter)
		if err != nil {
			return err
		}
		c := notebook.Cells[idx]
		results = append(results, cellHash{
			Index:    idx,
			ID:       c.ID,
			CellType: c.CellType,
			Hash:     nb.ComputeCellHash(c.CellType, nb.CellText(c)),
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}
