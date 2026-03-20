package audit

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zricethezav/gitleaks/v8/detect"

	"github.com/nii-cloud/nblibram/internal/filter"
	nb "github.com/nii-cloud/nblibram/internal/notebook"
)

type finding struct {
	CellIndex int    `json:"cell_index"`
	CellType  string `json:"cell_type"`
	RuleID    string `json:"rule_id"`
	Secret    string `json:"secret"`
	Location  string `json:"location"`
}

func Run(args []string) error {
	fs := flag.NewFlagSet("audit", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	format := fs.String("format", "text", "output format: text or json")
	if err := fs.Parse(args); err != nil {
		return err
	}

	notebook, err := nb.Read(*file)
	if err != nil {
		return err
	}

	detector, err := filter.NewDetector("")
	if err != nil {
		return err
	}

	var findings []finding

	for i, cell := range notebook.Cells {
		source := strings.Join(cell.Source, "")
		for _, f := range detector.Detect(detect.Fragment{
			Raw:      source,
			FilePath: "notebook.ipynb",
		}) {
			findings = append(findings, finding{
				CellIndex: i,
				CellType:  cell.CellType,
				RuleID:    f.RuleID,
				Secret:    f.Secret,
				Location:  "source",
			})
		}

		for _, out := range cell.Outputs {
			text := strings.Join(out.Text, "")
			for _, f := range detector.Detect(detect.Fragment{
				Raw:      text,
				FilePath: "notebook.ipynb",
			}) {
				findings = append(findings, finding{
					CellIndex: i,
					CellType:  cell.CellType,
					RuleID:    f.RuleID,
					Secret:    f.Secret,
					Location:  "output",
				})
			}
		}
	}

	switch *format {
	case "text":
		if len(findings) == 0 {
			fmt.Fprintln(os.Stderr, "No leaks detected.")
			return nil
		}
		for _, f := range findings {
			fmt.Fprintf(os.Stdout, "cell %d (%s) [%s] %s: %s\n",
				f.CellIndex, f.CellType, f.Location, f.RuleID, f.Secret)
		}
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(findings); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported format: %s", *format)
	}

	if len(findings) > 0 {
		os.Exit(1)
	}
	return nil
}
