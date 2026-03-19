package notebook

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const QueryUsage = "start:N, match:REGEX, contains:TEXT, id:ID, meme:UUID"

type MultiFlag []string

func (m *MultiFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *MultiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

type QueryFilter struct {
	Start    *int
	Match    []*regexp.Regexp
	Contains []string
	ID       *string
	Meme     *string
}

func ParseQueryFlags(values []string) (QueryFilter, error) {
	var filter QueryFilter
	for _, raw := range values {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) != 2 {
			return filter, fmt.Errorf("invalid query: %s", raw)
		}
		key := parts[0]
		value := parts[1]
		switch key {
		case "start":
			if filter.Start != nil {
				return filter, errors.New("start query specified multiple times")
			}
			idx, err := strconv.Atoi(value)
			if err != nil {
				return filter, fmt.Errorf("invalid start index: %s", value)
			}
			filter.Start = &idx
		case "match":
			re, err := regexp.Compile(value)
			if err != nil {
				return filter, fmt.Errorf("invalid regex in match: %s", err)
			}
			filter.Match = append(filter.Match, re)
		case "contains":
			filter.Contains = append(filter.Contains, value)
		case "id":
			if filter.ID != nil {
				return filter, errors.New("id query specified multiple times")
			}
			filter.ID = &value
		case "meme":
			if filter.Meme != nil {
				return filter, errors.New("meme query specified multiple times")
			}
			filter.Meme = &value
		default:
			return filter, fmt.Errorf("unsupported query key: %s", key)
		}
	}
	return filter, nil
}

func LocateStartCell(nb *Notebook, filter QueryFilter) (int, error) {
	if filter.Start != nil {
		idx := *filter.Start
		if idx < 0 || idx >= len(nb.Cells) {
			return 0, fmt.Errorf("start index %d out of range", idx)
		}
		if !MatchesCell(nb.Cells[idx], idx, filter) {
			return 0, fmt.Errorf("cell %d does not satisfy remaining queries", idx)
		}
		return idx, nil
	}
	for idx, c := range nb.Cells {
		if MatchesCell(c, idx, filter) {
			return idx, nil
		}
	}
	return 0, errors.New("no cell matched the query")
}

func MatchesCell(c Cell, idx int, filter QueryFilter) bool {
	if filter.ID != nil {
		if c.ID == "" || c.ID != *filter.ID {
			return false
		}
	}
	if filter.Meme != nil {
		meme := GetMemeID(c)
		pattern := *filter.Meme
		if strings.HasSuffix(pattern, "*") {
			if !strings.HasPrefix(meme, pattern[:len(pattern)-1]) {
				return false
			}
		} else if meme != pattern {
			return false
		}
	}
	text := CellText(c)
	for _, re := range filter.Match {
		if !re.MatchString(text) {
			return false
		}
	}
	for _, s := range filter.Contains {
		if !strings.Contains(text, s) {
			return false
		}
	}
	return true
}
