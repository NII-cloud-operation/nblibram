package outputs

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/nii-cloud/nblibram/internal/filter"
	nb "github.com/nii-cloud/nblibram/internal/notebook"
)

type outputFormat struct {
	kind string
	mime string
}

func Run(args []string) error {
	fs := flag.NewFlagSet("outputs", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	format := fs.String("format", "text", "output format: text, json, or raw")
	mime := fs.String("mime", "", "specific MIME type to extract")
	noFilter := fs.Bool("no-filter", false, "disable privacy filters")
	var queryFlags nb.MultiFlag
	fs.Var(&queryFlags, "query", "cell query ("+nb.QueryUsage+")")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(queryFlags) == 0 {
		return errors.New("outputs requires at least one --query")
	}

	notebook, err := nb.Read(*file)
	if err != nil {
		return err
	}

	qf, err := nb.ParseQueryFlags(queryFlags)
	if err != nil {
		return err
	}

	startIdx, err := nb.LocateStartCell(notebook, qf)
	if err != nil {
		return err
	}

	outFmt, err := parseOutputFormat(*format, *mime)
	if err != nil {
		return err
	}

	cell := notebook.Cells[startIdx]
	if sanitizer := filter.LoadDefault(*noFilter); sanitizer != nil {
		sanitizer.SanitizeCells([]nb.Cell{cell})
	}

	return emitOutputs(cell, outFmt)
}

func parseOutputFormat(format, mime string) (outputFormat, error) {
	switch format {
	case "text":
		return outputFormat{kind: "text", mime: mime}, nil
	case "json":
		return outputFormat{kind: "json", mime: mime}, nil
	case "raw":
		if mime == "" {
			return outputFormat{}, errors.New("raw format requires --mime")
		}
		return outputFormat{kind: "raw", mime: mime}, nil
	default:
		return outputFormat{}, fmt.Errorf("unsupported outputs format: %s", format)
	}
}

func emitOutputs(c nb.Cell, format outputFormat) error {
	if len(c.Outputs) == 0 {
		return errors.New("cell has no outputs")
	}
	switch format.kind {
	case "text":
		return emitTextOutputs(c, format.mime)
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(c.Outputs)
	case "raw":
		return emitRawMIME(c, format.mime)
	default:
		return errors.New("unknown output format")
	}
}

func emitTextOutputs(c nb.Cell, mime string) error {
	for _, out := range c.Outputs {
		if out.OutputType == "stream" {
			if mime != "" && mime != out.Name && mime != out.Stream {
				continue
			}
			fmt.Print(nb.CellText(nb.Cell{Source: out.Text}))
		}
		if out.OutputType == "error" {
			fmt.Printf("%s: %s\n", out.Ename, out.Evalue)
			for _, line := range out.Traceback {
				fmt.Println(line)
			}
		}
		if len(out.Data) > 0 {
			if mime == "" {
				if txt, ok := out.Data["text/plain"]; ok {
					fmt.Println(txt)
				}
				continue
			}
			if val, ok := out.Data[mime]; ok {
				fmt.Println(val)
			}
		}
	}
	return nil
}

func emitRawMIME(c nb.Cell, mime string) error {
	for _, out := range c.Outputs {
		if len(out.Data) == 0 {
			continue
		}
		val, ok := out.Data[mime]
		if !ok {
			continue
		}
		switch payload := val.(type) {
		case string:
			data, err := decodeBase64(payload)
			if err != nil {
				return err
			}
			if _, err := os.Stdout.Write(data); err != nil {
				return err
			}
		case []interface{}:
			for _, part := range payload {
				str, ok := part.(string)
				if !ok {
					return errors.New("unexpected non-string payload segment")
				}
				data, err := decodeBase64(str)
				if err != nil {
					return err
				}
				if _, err := os.Stdout.Write(data); err != nil {
					return err
				}
			}
		default:
			return errors.New("unsupported payload type for raw output")
		}
		return nil
	}
	return fmt.Errorf("no outputs contained MIME %s", mime)
}

func decodeBase64(val string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(val)
}
