package store

import (
	"context"
	"github.com/FreeLeh/GoFreeDB/internal/models"
	"time"
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
) (models.A1Range, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	result, err := wrapper.OverwriteRows(
		ctx,
		spreadsheetID,
		models.NewA1Range(scratchpadSheetName, defaultKVTableRange),
		[][]interface{}{{scratchpadBooked}},
	)
	if err != nil {
		return models.A1Range{}, err
	}
	return result.UpdatedRange, nil
}
