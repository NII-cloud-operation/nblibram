package notebook

import (
	"testing"
)

func TestParseQueryFlags(t *testing.T) {
	tests := []struct {
		input []string
		check func(QueryFilter) bool
	}{
		{[]string{"start:3"}, func(f QueryFilter) bool { return f.Start != nil && *f.Start == 3 }},
		{[]string{"id:abc"}, func(f QueryFilter) bool { return f.ID != nil && *f.ID == "abc" }},
		{[]string{"contains:hello"}, func(f QueryFilter) bool { return len(f.Contains) == 1 && f.Contains[0] == "hello" }},
		{[]string{"match:^import"}, func(f QueryFilter) bool { return len(f.Match) == 1 }},
		{[]string{"meme:uuid-1234"}, func(f QueryFilter) bool { return f.Meme != nil && *f.Meme == "uuid-1234" }},
	}

	for _, tt := range tests {
		f, err := ParseQueryFlags(tt.input)
		if err != nil {
			t.Errorf("ParseQueryFlags(%v): %v", tt.input, err)
			continue
		}
		if !tt.check(f) {
			t.Errorf("ParseQueryFlags(%v): check failed", tt.input)
		}
	}
}

func TestParseQueryFlagsErrors(t *testing.T) {
	bad := [][]string{
		{"nocolon"},
		{"start:abc"},
		{"match:[invalid"},
		{"start:1", "start:2"},
		{"id:a", "id:b"},
		{"unknown:x"},
	}
	for _, input := range bad {
		_, err := ParseQueryFlags(input)
		if err == nil {
			t.Errorf("expected error for %v", input)
		}
	}
}

func TestLocateStartCell(t *testing.T) {
	nb := loadTestNotebook(t)

	idx, err := LocateStartCell(nb, QueryFilter{Start: intPtr(0)})
	if err != nil || idx != 0 {
		t.Errorf("start:0 => idx=%d, err=%v", idx, err)
	}

	id := "code2"
	idx, err = LocateStartCell(nb, QueryFilter{ID: &id})
	if err != nil || idx != 3 {
		t.Errorf("id:code2 => idx=%d, err=%v", idx, err)
	}

	idx, err = LocateStartCell(nb, QueryFilter{Contains: []string{"x = 1"}})
	if err != nil || idx != 3 {
		t.Errorf("contains:'x = 1' => idx=%d, err=%v", idx, err)
	}
}

func TestLocateStartCellMeme(t *testing.T) {
	nb := loadTestNotebook(t)

	meme := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	idx, err := LocateStartCell(nb, QueryFilter{Meme: &meme})
	if err != nil || idx != 2 {
		t.Errorf("meme exact => idx=%d, err=%v", idx, err)
	}

	memeWild := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11*"
	idx, err = LocateStartCell(nb, QueryFilter{Meme: &memeWild})
	if err != nil || idx != 2 {
		t.Errorf("meme wildcard => idx=%d, err=%v", idx, err)
	}
}

func intPtr(i int) *int { return &i }
