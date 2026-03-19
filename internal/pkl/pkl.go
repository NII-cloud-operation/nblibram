package pkl

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	nb "github.com/nii-cloud/nblibram/internal/notebook"
)

func Run(args []string) error {
	fs := flag.NewFlagSet("pkl", flag.ExitOnError)
	file := fs.String("file", "", "path to .pkl file (required)")
	format := fs.String("format", "json", "output format: json or text")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" {
		return errors.New("pkl requires --file")
	}

	rec, err := nb.ReadPklRecord(*file)
	if err != nil {
		return err
	}

	switch *format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(rec)
	case "text":
		return emitPklText(rec)
	default:
		return fmt.Errorf("unsupported format: %s", *format)
	}
}

func emitPklText(rec *nb.PklRecord) error {
	var content map[string]interface{}
	if err := json.Unmarshal(rec.Content, &content); err != nil {
		return err
	}

	switch rec.MsgType {
	case "stream":
		if text, ok := content["text"].(string); ok {
			fmt.Print(text)
		}
	case "execute_result", "display_data":
		if data, ok := content["data"].(map[string]interface{}); ok {
			if text, ok := data["text/plain"].(string); ok {
				fmt.Println(text)
			}
		}
	case "error":
		if ename, ok := content["ename"].(string); ok {
			evalue, _ := content["evalue"].(string)
			fmt.Printf("%s: %s\n", ename, evalue)
		}
		if tb, ok := content["traceback"].([]interface{}); ok {
			for _, line := range tb {
				if s, ok := line.(string); ok {
					fmt.Println(s)
				}
			}
		}
	default:
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(content)
	}
	return nil
}
