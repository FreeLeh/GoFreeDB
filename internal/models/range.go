package models

import (
	"fmt"
	"strconv"
	"strings"
)

type A1Range struct {
	Original  string
	SheetName string
	FromCell  string
	ToCell    string
}

func (r A1Range) Range() string {
	return fmt.Sprintf("%s:%s", r.FromCell, r.ToCell)
}

func (r A1Range) NumCols() int {
	toCellColIdx := CellToColIdx(r.ToCell)
	fromCellColIdx := CellToColIdx(r.FromCell)

	if toCellColIdx >= fromCellColIdx {
		return toCellColIdx - fromCellColIdx + 1
	}
	return fromCellColIdx - toCellColIdx + 1
}

func (r A1Range) NumRows() int {
	fromIdx := strings.IndexAny(r.FromCell, digits)
	toIdx := strings.IndexAny(r.ToCell, digits)

	if fromIdx == -1 || toIdx == -1 {
		return 0
	}

	from, err := strconv.ParseInt(r.FromCell[fromIdx:], 10, 64)
	if err != nil {
		return 0
	}

	to, err := strconv.ParseInt(r.ToCell[toIdx:], 10, 64)
	if err != nil {
		return 0
	}

	if to >= from {
		return int(to - from + 1)
	}
	return int(from - to + 1)
}

func NewA1Range(sheetName string, rng string) A1Range {
	return NewA1RangeFromString(sheetName + "!" + rng)
}

func NewA1RangeFromString(s string) A1Range {
	exclamationIdx := strings.Index(s, "!")
	colonIdx := strings.Index(s, ":")

	if exclamationIdx == -1 {
		if colonIdx == -1 {
			return A1Range{
				Original:  s,
				SheetName: "",
				FromCell:  s,
				ToCell:    s,
			}
		} else {
			return A1Range{
				Original:  s,
				SheetName: "",
				FromCell:  s[:colonIdx],
				ToCell:    s[colonIdx+1:],
			}
		}
	} else {
		if colonIdx == -1 {
			return A1Range{
				Original:  s,
				SheetName: s[:exclamationIdx],
				FromCell:  s[exclamationIdx+1:],
				ToCell:    s[exclamationIdx+1:],
			}
		} else {
			return A1Range{
				Original:  s,
				SheetName: s[:exclamationIdx],
				FromCell:  s[exclamationIdx+1 : colonIdx],
				ToCell:    s[colonIdx+1:],
			}
		}
	}
}

type ColIdx struct {
	Name string
	Idx  int
}

type ColsMapping map[string]ColIdx

func (m ColsMapping) NameMap() map[string]string {
	result := make(map[string]string, 0)
	for col, val := range m {
		result[col] = val.Name
	}
	return result
}

const (
	alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits   = "0123456789"
)

func GenerateColumnMapping(columns []string) map[string]ColIdx {
	mapping := make(map[string]ColIdx, len(columns))
	for n, col := range columns {
		mapping[col] = ColIdx{
			Name: GenerateColumnName(n),
			Idx:  n,
		}
	}
	return mapping
}

func GenerateColumnName(n int) string {
	// This is not purely a Base26 conversion since the second char can start from "A" (or 0) again.
	// In a normal Base26 int to string conversion, the second char can only start from "B" (or 1).
	// Hence, we need to hack it by checking the first round separately from the subsequent round.
	// For the subsequent rounds, we need to subtract by 1 first or else it will always start from 1 (not 0).
	col := string(alphabet[n%26])
	n = n / 26

	for {
		if n <= 0 {
			break
		}

		n -= 1
		col = string(alphabet[n%26]) + col
		n = n / 26
	}

	return col
}

func CellToColIdx(cell string) int {
	digitIdx := strings.IndexAny(cell, digits)
	if digitIdx != -1 {
		cell = cell[:digitIdx]
	}
	cell = strings.ToUpper(cell)

	index := 0
	for i := 0; i < len(cell); i++ {
		index = index*26 + int(cell[i]-'A'+1)
	}
	return index
}
