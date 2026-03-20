package filter

import (
	"encoding/json"
	"flag"
	"io"
	"os"
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
		// No filters to apply, pass through
		return passThrough(*file)
	}

	var input io.Reader = os.Stdin
	if *file != "" {
		f, err := os.Open(*file)
		if err != nil {
			return err
		}
		defer f.Close()
		input = f
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(input).Decode(&raw); err != nil {
		return err
	}

	sanitizeRawNotebook(sanitizer, raw)

	var output io.Writer = os.Stdout
	if *inPlace && *file != "" {
		f, err := os.Create(*file)
		if err != nil {
			return err
		}
		defer f.Close()
		output = f
	}

	enc := json.NewEncoder(output)
	enc.SetIndent("", " ")
	enc.SetEscapeHTML(false)
	return enc.Encode(raw)
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

func sanitizeRawNotebook(s *Sanitizer, raw map[string]interface{}) {
	if md, ok := raw["metadata"]; ok {
		raw["metadata"] = sanitizeValue(s, md)
	}

	cellsRaw, ok := raw["cells"].([]interface{})
	if !ok {
		return
	}

	for _, cellRaw := range cellsRaw {
		cell, ok := cellRaw.(map[string]interface{})
		if !ok {
			continue
		}
		if src, ok := cell["source"]; ok {
			cell["source"] = sanitizeSource(s, src)
		}
		if md, ok := cell["metadata"]; ok {
			cell["metadata"] = sanitizeValue(s, md)
		}
		outputsRaw, ok := cell["outputs"].([]interface{})
		if !ok {
			continue
		}
		for _, outRaw := range outputsRaw {
			out, ok := outRaw.(map[string]interface{})
			if !ok {
				continue
			}
			if txt, ok := out["text"]; ok {
				out["text"] = sanitizeSource(s, txt)
			}
			if data, ok := out["data"]; ok {
				out["data"] = sanitizeValue(s, data)
			}
			if tb, ok := out["traceback"].([]interface{}); ok {
				for i, line := range tb {
					if str, ok := line.(string); ok {
						tb[i] = s.Sanitize(str)
					}
				}
			}
			if ev, ok := out["evalue"].(string); ok {
				out["evalue"] = s.Sanitize(ev)
			}
		}
	}
}

func sanitizeSource(s *Sanitizer, src interface{}) interface{} {
	switch v := src.(type) {
	case string:
		return s.Sanitize(v)
	case []interface{}:
		for i, item := range v {
			if str, ok := item.(string); ok {
				v[i] = s.Sanitize(str)
			}
		}
		return v
	default:
		return src
	}
}

func sanitizeValue(s *Sanitizer, v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		return s.Sanitize(val)
	case map[string]interface{}:
		for k, v := range val {
			val[k] = sanitizeValue(s, v)
		}
		return val
	case []interface{}:
		for i, v := range val {
			val[i] = sanitizeValue(s, v)
		}
		return val
	default:
		return v
	}
}
