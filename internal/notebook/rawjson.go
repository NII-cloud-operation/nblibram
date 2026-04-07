package notebook

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type orderedField struct {
	Key   string
	Value json.RawMessage
}

func parseOrderedObject(data []byte) ([]orderedField, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if t != json.Delim('{') {
		return nil, fmt.Errorf("expected '{', got %v", t)
	}

	var fields []orderedField
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("expected string key, got %T", keyTok)
		}
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return nil, fmt.Errorf("decoding value for key %q: %w", key, err)
		}
		fields = append(fields, orderedField{Key: key, Value: value})
	}
	return fields, nil
}

func marshalOrderedObject(fields []orderedField) (json.RawMessage, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, f := range fields {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyJSON, err := json.Marshal(f.Key)
		if err != nil {
			return nil, err
		}
		buf.Write(keyJSON)
		buf.WriteByte(':')
		buf.Write(f.Value)
	}
	buf.WriteByte('}')
	return json.RawMessage(buf.Bytes()), nil
}

func parseRawArray(data json.RawMessage) ([]json.RawMessage, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if t != json.Delim('[') {
		return nil, fmt.Errorf("expected '[', got %v", t)
	}

	var elements []json.RawMessage
	for dec.More() {
		var elem json.RawMessage
		if err := dec.Decode(&elem); err != nil {
			return nil, err
		}
		elements = append(elements, elem)
	}
	return elements, nil
}

func marshalRawArray(elements []json.RawMessage) json.RawMessage {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, e := range elements {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.Write(e)
	}
	buf.WriteByte(']')
	return json.RawMessage(buf.Bytes())
}
