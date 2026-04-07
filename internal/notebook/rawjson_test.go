package notebook

import (
	"encoding/json"
	"testing"
)

func TestParseOrderedObjectKeyOrder(t *testing.T) {
	input := `{"z":1,"a":2,"m":3}`
	fields, err := parseOrderedObject([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	wantKeys := []string{"z", "a", "m"}
	if len(fields) != len(wantKeys) {
		t.Fatalf("got %d fields, want %d", len(fields), len(wantKeys))
	}
	for i, f := range fields {
		if f.Key != wantKeys[i] {
			t.Errorf("field %d: key=%q, want %q", i, f.Key, wantKeys[i])
		}
	}
}

func TestParseOrderedObjectNested(t *testing.T) {
	input := `{"outer":{"b":2,"a":1}}`
	fields, err := parseOrderedObject([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 1 || fields[0].Key != "outer" {
		t.Fatalf("unexpected: %v", fields)
	}
	inner, err := parseOrderedObject(fields[0].Value)
	if err != nil {
		t.Fatal(err)
	}
	if inner[0].Key != "b" || inner[1].Key != "a" {
		t.Errorf("nested key order lost: %s, %s", inner[0].Key, inner[1].Key)
	}
}

func TestParseOrderedObjectEscapedKey(t *testing.T) {
	input := `{"key\"with\\escape":true}`
	fields, err := parseOrderedObject([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if fields[0].Key != `key"with\escape` {
		t.Errorf("key = %q", fields[0].Key)
	}
}

func TestParseOrderedObjectUnicodeKey(t *testing.T) {
	input := `{"日本語キー":"値","emoji🎉":"data"}`
	fields, err := parseOrderedObject([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if fields[0].Key != "日本語キー" || fields[1].Key != "emoji🎉" {
		t.Errorf("unicode keys: %q, %q", fields[0].Key, fields[1].Key)
	}
}

func TestParseOrderedObjectEmpty(t *testing.T) {
	fields, err := parseOrderedObject([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 0 {
		t.Errorf("expected 0 fields, got %d", len(fields))
	}
}

func TestParseOrderedObjectNotObject(t *testing.T) {
	_, err := parseOrderedObject([]byte(`[1,2,3]`))
	if err == nil {
		t.Error("expected error for array input")
	}
}

func TestParseOrderedObjectInvalid(t *testing.T) {
	_, err := parseOrderedObject([]byte(`{invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestMarshalOrderedObjectRoundTrip(t *testing.T) {
	input := `{"z":1,"a":"hello","nested":{"x":true}}`
	fields, err := parseOrderedObject([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	out, err := marshalOrderedObject(fields)
	if err != nil {
		t.Fatal(err)
	}
	// Re-parse to verify structural equivalence and key order
	fields2, err := parseOrderedObject(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != len(fields2) {
		t.Fatalf("field count: %d vs %d", len(fields), len(fields2))
	}
	for i := range fields {
		if fields[i].Key != fields2[i].Key {
			t.Errorf("key %d: %q vs %q", i, fields[i].Key, fields2[i].Key)
		}
	}
}

func TestParseRawArrayBasic(t *testing.T) {
	input := json.RawMessage(`[1,"two",{"k":"v"},[4,5]]`)
	elems, err := parseRawArray(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(elems) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(elems))
	}
}

func TestParseRawArrayEmpty(t *testing.T) {
	elems, err := parseRawArray(json.RawMessage(`[]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(elems) != 0 {
		t.Errorf("expected 0 elements, got %d", len(elems))
	}
}

func TestParseRawArrayNotArray(t *testing.T) {
	_, err := parseRawArray(json.RawMessage(`{"key":"value"}`))
	if err == nil {
		t.Error("expected error for object input")
	}
}

func TestParseRawArrayInvalid(t *testing.T) {
	_, err := parseRawArray(json.RawMessage(`[invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestMarshalRawArrayRoundTrip(t *testing.T) {
	input := json.RawMessage(`[1,"two",{"k":"v"}]`)
	elems, err := parseRawArray(input)
	if err != nil {
		t.Fatal(err)
	}
	out := marshalRawArray(elems)

	var orig, result []json.RawMessage
	json.Unmarshal(input, &orig)
	json.Unmarshal(out, &result)
	if len(orig) != len(result) {
		t.Fatalf("length: %d vs %d", len(orig), len(result))
	}
}

func TestMarshalRawArrayEmpty(t *testing.T) {
	out := marshalRawArray(nil)
	if string(out) != "[]" {
		t.Errorf("expected [], got %s", out)
	}
}

func TestDeepNestedRoundTrip(t *testing.T) {
	input := `{"a":{"b":{"c":{"d":"deep"}}}}`
	fields, err := parseOrderedObject([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	out, err := marshalOrderedObject(fields)
	if err != nil {
		t.Fatal(err)
	}
	// Verify the deep value survived
	var m map[string]any
	json.Unmarshal(out, &m)
	a := m["a"].(map[string]any)
	b := a["b"].(map[string]any)
	c := b["c"].(map[string]any)
	if c["d"] != "deep" {
		t.Errorf("deep nested value lost")
	}
}
