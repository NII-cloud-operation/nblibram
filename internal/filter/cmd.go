package filter

import (
	"flag"
	"io"
	"os"

	nb "github.com/nii-cloud/nblibram/internal/notebook"
)

func Run(args []string) error {
	fs := flag.NewFlagSet("filter", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	inPlace := fs.Bool("i", false, "modify file in place")
	if err := fs.Parse(args); err != nil {
		return err
	}

	sanitizer := LoadDefault(false)
	if sanitizer == nil {
		return passThrough(*file)
	}

	notebook, err := nb.Read(*file)
	if err != nil {
		return err
	}

	sanitize := sanitizer.Sanitize

	// Sanitize each cell's raw JSON
	for i, raw := range notebook.CellsRaw {
		notebook.CellsRaw[i], err = nb.SanitizeCellRaw(raw, sanitize)
		if err != nil {
			return err
		}
	}

	// Sanitize notebook-level metadata
	notebook.SanitizeRawTopField("metadata", sanitize)

	return notebook.Write(*file, *inPlace)
}

func passThrough(file string) error {
	var input io.Reader = os.Stdin
	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		input = f
	}
	_, err := io.Copy(os.Stdout, input)
	return err
}
