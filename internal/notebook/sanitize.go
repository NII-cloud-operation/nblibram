package notebook

import (
	"bytes"
	"encoding/json"
)

// SanitizeRawValue recursively sanitizes all string values in raw JSON,
// preserving key order, field presence, and all non-string values exactly.
func SanitizeRawValue(raw json.RawMessage, sanitize func(string) string) (json.RawMessage, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return raw, nil
	}
	switch trimmed[0] {
	case '"':
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return nil, err
		}
		return json.Marshal(sanitize(s))
	case '{':
		fields, err := parseOrderedObject(trimmed)
		if err != nil {
			return nil, err
		}
		for i, f := range fields {
			fields[i].Value, err = SanitizeRawValue(f.Value, sanitize)
			if err != nil {
				return nil, err
			}
		}
		return marshalOrderedObject(fields)
	case '[':
		elements, err := parseRawArray(trimmed)
		if err != nil {
			return nil, err
		}
		for i, e := range elements {
			elements[i], err = SanitizeRawValue(e, sanitize)
			if err != nil {
				return nil, err
			}
		}
		return marshalRawArray(elements), nil
	default:
		return raw, nil
	}
}

// SanitizeCellRaw applies targeted sanitization to a cell's raw JSON.
// Only source, metadata, and output fields are sanitized; structural fields
// (cell_type, id, execution_count) pass through unchanged.
func SanitizeCellRaw(raw json.RawMessage, sanitize func(string) string) (json.RawMessage, error) {
	fields, err := parseOrderedObject(raw)
	if err != nil {
		return nil, err
	}
	for i, f := range fields {
		switch f.Key {
		case "source":
			fields[i].Value, err = SanitizeRawValue(f.Value, sanitize)
		case "metadata":
			fields[i].Value, err = SanitizeRawValue(f.Value, sanitize)
		case "outputs":
			fields[i].Value, err = sanitizeOutputsRaw(f.Value, sanitize)
		}
		if err != nil {
			return nil, err
		}
	}
	return marshalOrderedObject(fields)
}

// SanitizeRawTopField sanitizes a top-level field in the notebook's raw JSON.
func (nb *Notebook) SanitizeRawTopField(key string, sanitize func(string) string) {
	for i, f := range nb.rawTop {
		if f.Key == key {
			sanitized, err := SanitizeRawValue(f.Value, sanitize)
			if err == nil {
				nb.rawTop[i].Value = sanitized
			}
			return
		}
	}
}

func sanitizeOutputsRaw(raw json.RawMessage, sanitize func(string) string) (json.RawMessage, error) {
	outputs, err := parseRawArray(raw)
	if err != nil {
		return nil, err
	}
	for i, outRaw := range outputs {
		fields, err := parseOrderedObject(outRaw)
		if err != nil {
			return nil, err
		}
		for j, f := range fields {
			switch f.Key {
			case "text", "data", "traceback":
				fields[j].Value, err = SanitizeRawValue(f.Value, sanitize)
			case "evalue":
				fields[j].Value, err = SanitizeRawValue(f.Value, sanitize)
			}
			if err != nil {
				return nil, err
			}
		}
		outputs[i], err = marshalOrderedObject(fields)
		if err != nil {
			return nil, err
		}
	}
	return marshalRawArray(outputs), nil
}
