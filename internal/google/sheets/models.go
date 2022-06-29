package sheets

import "strings"

type A1Range struct {
	Original  string
	SheetName string
	FromCell  string
	ToCell    string
}

func NewA1Range(s string) A1Range {
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

type InsertRowsResult struct {
	UpdatedRange   A1Range
	UpdatedRows    int64
	UpdatedColumns int64
	UpdatedCells   int64
	InsertedValues [][]interface{}
}

type UpdateRowsResult struct {
	UpdatedRange   A1Range
	UpdatedRows    int64
	UpdatedColumns int64
	UpdatedCells   int64
	UpdatedValues  [][]interface{}
}
