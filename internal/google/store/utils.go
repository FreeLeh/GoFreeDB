package store

import (
	"context"
	"time"

	"github.com/FreeLeh/GoFreeDB/internal/google/sheets"
)

func ensureSheets(wrapper sheetsWrapper, spreadsheetID string, sheetName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	return wrapper.CreateSheet(ctx, spreadsheetID, sheetName)
}

func findScratchpadLocation(
	wrapper sheetsWrapper,
	spreadsheetID string,
	scratchpadSheetName string,
) (sheets.A1Range, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	result, err := wrapper.OverwriteRows(
		ctx,
		spreadsheetID,
		scratchpadSheetName+"!"+defaultKVTableRange,
		[][]interface{}{{scratchpadBooked}},
	)
	if err != nil {
		return sheets.A1Range{}, err
	}
	return result.UpdatedRange, nil
}
