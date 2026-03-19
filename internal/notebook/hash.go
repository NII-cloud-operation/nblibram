package notebook

import (
	"fmt"
	"unicode/utf16"
)

// djb2 hash over UTF-16 code units to match mynerva's TS charCodeAt() behavior.
func ComputeCellHash(cellType, source string) string {
	str := cellType + "\x00" + source
	hash := uint32(5381)
	for _, u := range utf16.Encode([]rune(str)) {
		hash = (hash * 33) ^ uint32(u)
	}
	return fmt.Sprintf("%x", hash)
}
