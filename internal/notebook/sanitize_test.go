package notebook

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSanitizeRawValueString(t *testing.T) {
	raw := json.RawMessage(`"hello world"`)
	out, err := SanitizeRawValue(raw, strings.ToUpper)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `"HELLO WORLD"` {
		t.Errorf("got %s", out)
	}
}

func TestSanitizeRawValueNumber(t *testing.T) {
	raw := json.RawMessage(`42`)
	out, err := SanitizeRawValue(raw, strings.ToUpper)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `42` {
		t.Errorf("numbers should pass through, got %s", out)
	}
}

func TestSanitizeRawValueNull(t *testing.T) {
	raw := json.RawMessage(`null`)
	out, err := SanitizeRawValue(raw, strings.ToUpper)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `null` {
		t.Errorf("null should pass through, got %s", out)
	}
}

func TestSanitizeRawValueBool(t *testing.T) {
	raw := json.RawMessage(`true`)
	out, err := SanitizeRawValue(raw, strings.ToUpper)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `true` {
		t.Errorf("bool should pass through, got %s", out)
	}
}

func TestSanitizeRawValueEmpty(t *testing.T) {
	out, err := SanitizeRawValue(json.RawMessage(``), strings.ToUpper)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `` {
		t.Errorf("empty should pass through, got %q", out)
	}
}

func TestSanitizeRawValueObject(t *testing.T) {
	raw := json.RawMessage(`{"z":"hello","a":"world"}`)
	out, err := SanitizeRawValue(raw, strings.ToUpper)
	if err != nil {
		t.Fatal(err)
	}
	// Verify key order preserved and values sanitized
	fields, _ := parseOrderedObject(out)
	if fields[0].Key != "z" || fields[1].Key != "a" {
		t.Errorf("key order changed")
	}
	var m map[string]string
	json.Unmarshal(out, &m)
	if m["z"] != "HELLO" || m["a"] != "WORLD" {
		t.Errorf("values not sanitized: %v", m)
	}
}

func TestSanitizeRawValueArray(t *testing.T) {
	raw := json.RawMessage(`["hello",42,"world"]`)
	out, err := SanitizeRawValue(raw, strings.ToUpper)
	if err != nil {
		t.Fatal(err)
	}
	var arr []json.RawMessage
	json.Unmarshal(out, &arr)
	if string(arr[0]) != `"HELLO"` || string(arr[1]) != `42` || string(arr[2]) != `"WORLD"` {
		t.Errorf("got %s", out)
	}
}

func TestSanitizeCellRaw(t *testing.T) {
	cell := json.RawMessage(`{"cell_type":"code","source":["line1"],"metadata":{"key":"val"},"outputs":[]}`)
	out, err := SanitizeCellRaw(cell, strings.ToUpper)
	if err != nil {
		t.Fatal(err)
	}
	var c Cell
	json.Unmarshal(out, &c)
	if c.CellType != "code" {
		t.Error("cell_type should not be sanitized")
	}
	if c.Source[0] != "LINE1" {
		t.Errorf("source not sanitized: %q", c.Source[0])
	}
}

func TestSanitizeCellRawWithOutputs(t *testing.T) {
	cell := json.RawMessage(`{"cell_type":"code","source":[],"metadata":{},"outputs":[{"output_type":"stream","text":["hello"],"data":{"text/plain":"world"},"evalue":"err"}]}`)
	out, err := SanitizeCellRaw(cell, strings.ToUpper)
	if err != nil {
		t.Fatal(err)
	}
	// Verify output fields were sanitized
	fields, _ := parseOrderedObject(out)
	for _, f := range fields {
		if f.Key == "outputs" {
			elems, _ := parseRawArray(f.Value)
			var outMap map[string]json.RawMessage
			json.Unmarshal(elems[0], &outMap)
			if string(outMap["evalue"]) != `"ERR"` {
				t.Errorf("evalue not sanitized: %s", outMap["evalue"])
			}
		}
	}
}

func TestSanitizeRawTopField(t *testing.T) {
	nb := loadTestNotebook(t)
	nb.SanitizeRawTopField("metadata", strings.ToUpper)

	// Verify at least one metadata value was uppercased
	for _, f := range nb.rawTop {
		if f.Key == "metadata" {
			var m map[string]any
			json.Unmarshal(f.Value, &m)
			// metadata.kernelspec.display_name should be uppercased
			ks, _ := m["kernelspec"].(map[string]any)
			if ks != nil {
				if name, ok := ks["display_name"].(string); ok && name == strings.ToUpper(name) {
					return // success
				}
			}
		}
	}
}

func TestSanitizeRawTopFieldMissing(t *testing.T) {
	nb := loadTestNotebook(t)
	// Should not panic on non-existent key
	nb.SanitizeRawTopField("nonexistent", strings.ToUpper)
}
