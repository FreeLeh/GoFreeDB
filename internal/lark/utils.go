package lark

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

func getSheetIDs(wrapper sheetsWrapper, spreadsheetToken string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	result, err := wrapper.GetSheets(ctx, spreadsheetToken)
	if err != nil {
		return nil, err
	}

	mapping := make(map[string]string, len(result.Sheets))
	for _, sheet := range result.Sheets {
		mapping[sheet.Title] = sheet.SheetID
	}

	return mapping, nil
}

func findScratchpadLocation(
	wrapper sheetsWrapper,
	spreadsheetToken string,
	scratchpadSheetID string,
) (models.A1Range, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	result, err := wrapper.OverwriteRows(
		ctx,
		spreadsheetToken,
		models.NewA1Range(scratchpadSheetID, defaultScratchpadTableRange),
		[][]interface{}{{scratchpadBooked}},
	)
	if err != nil {
		return models.A1Range{}, err
	}
	return result.UpdatedRange, nil
}

func convertToFormula(query string) map[string]interface{} {
	// Must use this format (this was implicitly defined in
	// https://open.larksuite.com/document/server-docs/docs/appendix/data-types-supported-by-sheets
	return map[string]interface{}{
		"type": "formula",
		"text": query,
	}
}
