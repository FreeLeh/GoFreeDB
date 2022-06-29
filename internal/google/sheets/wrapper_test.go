package sheets

import (
	"context"
	"net/http"
	"testing"

	"github.com/FreeLeh/GoFreeLeh/google/auth"
	"github.com/FreeLeh/GoFreeLeh/internal/google/fixtures"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestCreateSpreadsheet(t *testing.T) {
	path := fixtures.PathToFixture("service_account.json")

	auth, err := auth.NewService(path, []string{}, auth.ServiceConfig{})
	assert.Nil(t, err, "should not have any error instantiating a new service account client")

	wrapper, err := NewWrapper(auth)
	assert.Nil(t, err, "should not have any error instantiating a new sheets wrapper")

	gock.InterceptClient(auth.HTTPClient())

	t.Run("successful", func(t *testing.T) {
		expectedReqBody := map[string]map[string]string{
			"properties": {
				"title": "title",
			},
		}
		resp := map[string]string{"spreadsheetId": "123"}

		gock.New("https://sheets.googleapis.com").
			Post("/v4/spreadsheets").
			JSON(expectedReqBody).
			Reply(http.StatusOK).
			JSON(resp)

		sid, err := wrapper.CreateSpreadsheet(context.Background(), "title")
		assert.Equal(t, "123", sid, "returned spreadsheetID does not match with the mocked spreadsheetID")
		assert.Nil(t, err, "should not have any error creating a new spreadsheet")
	})

	t.Run("http500", func(t *testing.T) {
		expectedReqBody := map[string]map[string]string{
			"properties": {
				"title": "title",
			},
		}

		gock.New("https://sheets.googleapis.com").
			Post("/v4/spreadsheets").
			JSON(expectedReqBody).
			Reply(http.StatusInternalServerError)

		sid, err := wrapper.CreateSpreadsheet(context.Background(), "title")
		assert.Equal(t, "", sid, "returned spreadsheetID should be empty as there is HTTP error")
		assert.NotNil(t, err, "should have an error when creating the spreadsheet as there is HTTP error")
	})

	t.Run("empty_title", func(t *testing.T) {
		expectedReqBody := map[string]map[string]string{
			"properties": {
				"title": "title",
			},
		}
		resp := map[string]string{"spreadsheetId": "123"}

		gock.New("https://sheets.googleapis.com").
			Post("/v4/spreadsheets").
			JSON(expectedReqBody).
			Reply(http.StatusOK).
			JSON(resp)

		sid, err := wrapper.CreateSpreadsheet(context.Background(), "title")
		assert.Equal(t, "123", sid, "returned spreadsheetID does not match with the mocked spreadsheetID")
		assert.Nil(t, err, "should not have any error creating a new spreadsheet")
	})
}

func TestCreateSheet(t *testing.T) {
	path := fixtures.PathToFixture("service_account.json")

	auth, err := auth.NewService(path, []string{}, auth.ServiceConfig{})
	assert.Nil(t, err, "should not have any error instantiating a new service account client")

	wrapper, err := NewWrapper(auth)
	assert.Nil(t, err, "should not have any error instantiating a new sheets wrapper")

	gock.InterceptClient(auth.HTTPClient())

	t.Run("successful", func(t *testing.T) {
		expectedReqBody := map[string][]map[string]map[string]map[string]string{
			"requests": {
				{
					"addSheet": {
						"properties": {
							"title": "sheet",
						},
					},
				},
			},
		}
		resp := map[string]interface{}{
			"spreadsheetId": "123",
			"replies": []map[string]map[string]map[string]string{
				{
					"addSheet": {
						"properties": {
							"title": "sheet",
						},
					},
				},
			},
		}

		gock.New("https://sheets.googleapis.com").
			Post("/v4/spreadsheets/123:batchUpdate").
			JSON(expectedReqBody).
			Reply(http.StatusOK).
			JSON(resp)

		err := wrapper.CreateSheet(context.Background(), "123", "sheet")
		assert.Nil(t, err, "should not have any error creating a new sheet")
	})

	t.Run("http500", func(t *testing.T) {
		expectedReqBody := map[string][]map[string]map[string]map[string]string{
			"requests": {
				{
					"addSheet": {
						"properties": {
							"title": "sheet",
						},
					},
				},
			},
		}

		gock.New("https://sheets.googleapis.com").
			Post("/v4/spreadsheets/123:batchUpdate").
			JSON(expectedReqBody).
			Reply(http.StatusInternalServerError)

		err := wrapper.CreateSheet(context.Background(), "123", "sheet")
		assert.NotNil(t, err, "should have an error when creating a sheet as there is HTTP error")
	})

	t.Run("empty_title", func(t *testing.T) {
		expectedReqBody := map[string][]map[string]map[string]map[string]string{
			"requests": {
				{
					"addSheet": {"properties": {}},
				},
			},
		}
		resp := map[string]interface{}{
			"spreadsheetId": "123",
			"replies": []map[string]map[string]map[string]string{
				{
					"addSheet": {
						"properties": {
							"title": "untitled",
						},
					},
				},
			},
		}

		gock.New("https://sheets.googleapis.com").
			Post("/v4/spreadsheets/123:batchUpdate").
			JSON(expectedReqBody).
			Reply(http.StatusOK).
			JSON(resp)

		err := wrapper.CreateSheet(context.Background(), "123", "")
		assert.Nil(t, err, "should not have any error creating a new sheet")
	})
}

func TestInsertAppendRows(t *testing.T) {
	path := fixtures.PathToFixture("service_account.json")

	auth, err := auth.NewService(path, []string{}, auth.ServiceConfig{})
	assert.Nil(t, err, "should not have any error instantiating a new service account client")

	wrapper, err := NewWrapper(auth)
	assert.Nil(t, err, "should not have any error instantiating a new sheets wrapper")

	gock.InterceptClient(auth.HTTPClient())

	t.Run("successful", func(t *testing.T) {
		expectedParams := map[string]string{
			"includeValuesInResponse":   "true",
			"responseValueRenderOption": responseValueRenderFormatted,
			"insertDataOption":          string(appendModeOverwrite),
			"valueInputOption":          valueInputUserEntered,
		}
		expectedReqBody := map[string]interface{}{
			"majorDimension": majorDimensionRows,
			"range":          "Sheet1!A1:A2",
			"values": [][]interface{}{
				{"1", "2"},
				{"3", "4"},
			},
		}
		resp := map[string]interface{}{
			"spreadsheetId": "123",
			"tableRange":    "Sheet1!A1:A2",
			"updates": map[string]interface{}{
				"spreadsheetId":  "123",
				"updatedRange":   "Sheet1!A1:B3",
				"updatedRows":    2,
				"updatedColumns": 2,
				"updatedCells":   4,
				"updatedData": map[string]interface{}{
					"range":          "Sheet1!A1:B3",
					"majorDimension": majorDimensionRows,
					"values": [][]interface{}{
						{"1", "2"},
						{"3", "4"},
					},
				},
			},
		}

		gock.New("https://sheets.googleapis.com").
			Post("/v4/spreadsheets/123/values/Sheet1!A1:A2:append").
			MatchParams(expectedParams).
			JSON(expectedReqBody).
			Reply(http.StatusOK).
			JSON(resp)

		values := [][]interface{}{{"1", "2"}, {"3", "4"}}
		res, err := wrapper.insertRows(context.Background(), "123", "Sheet1!A1:A2", values, appendModeOverwrite)

		assert.Nil(t, err, "should not have any error inserting rows")
		assert.Equal(t, res.UpdatedRange, NewA1Range("Sheet1!A1:B3"))
		assert.Equal(t, res.UpdatedRows, int64(2))
		assert.Equal(t, res.UpdatedColumns, int64(2))
		assert.Equal(t, res.UpdatedCells, int64(4))
		assert.Equal(t, res.InsertedValues, values)
	})

	t.Run("http500", func(t *testing.T) {
		expectedParams := map[string]string{
			"includeValuesInResponse":   "true",
			"responseValueRenderOption": responseValueRenderFormatted,
			"insertDataOption":          string(appendModeOverwrite),
			"valueInputOption":          valueInputUserEntered,
		}
		expectedReqBody := map[string]interface{}{
			"majorDimension": majorDimensionRows,
			"range":          "Sheet1!A1:A2",
			"values": [][]interface{}{
				{"1", "2"},
				{"3", "4"},
			},
		}

		gock.New("https://sheets.googleapis.com").
			Post("/v4/spreadsheets/123/values/Sheet1!A1:A2:append").
			MatchParams(expectedParams).
			JSON(expectedReqBody).
			Reply(http.StatusInternalServerError)

		values := [][]interface{}{{"1", "2"}, {"3", "4"}}
		res, err := wrapper.insertRows(context.Background(), "123", "Sheet1!A1:A2", values, appendModeOverwrite)

		assert.NotNil(t, err, "should have error inserting a new row")
		assert.Equal(t, res.UpdatedRange, NewA1Range(""))
		assert.Equal(t, res.UpdatedRows, int64(0))
		assert.Equal(t, res.UpdatedColumns, int64(0))
		assert.Equal(t, res.UpdatedCells, int64(0))
		assert.Equal(t, res.InsertedValues, [][]interface{}(nil))
	})
}

func TestUpdateRows(t *testing.T) {
	path := fixtures.PathToFixture("service_account.json")

	auth, err := auth.NewService(path, []string{}, auth.ServiceConfig{})
	assert.Nil(t, err, "should not have any error instantiating a new service account client")

	wrapper, err := NewWrapper(auth)
	assert.Nil(t, err, "should not have any error instantiating a new sheets wrapper")

	gock.InterceptClient(auth.HTTPClient())

	t.Run("successful", func(t *testing.T) {
		expectedParams := map[string]string{
			"includeValuesInResponse":   "true",
			"responseValueRenderOption": responseValueRenderFormatted,
			"valueInputOption":          valueInputUserEntered,
		}
		expectedReqBody := map[string]interface{}{
			"majorDimension": majorDimensionRows,
			"range":          "Sheet1!A1:A2",
			"values": [][]interface{}{
				{"1", "2"},
				{"3", "4"},
			},
		}
		resp := map[string]interface{}{
			"spreadsheetId":  "123",
			"updatedRange":   "Sheet1!A1:B3",
			"updatedRows":    2,
			"updatedColumns": 2,
			"updatedCells":   4,
			"updatedData": map[string]interface{}{
				"range":          "Sheet1!A1:B3",
				"majorDimension": majorDimensionRows,
				"values": [][]interface{}{
					{"1", "2"},
					{"3", "4"},
				},
			},
		}

		gock.New("https://sheets.googleapis.com").
			Put("/v4/spreadsheets/123/values/Sheet1!A1:A2").
			MatchParams(expectedParams).
			JSON(expectedReqBody).
			Reply(http.StatusOK).
			JSON(resp)

		values := [][]interface{}{{"1", "2"}, {"3", "4"}}
		res, err := wrapper.UpdateRows(context.Background(), "123", "Sheet1!A1:A2", values)

		assert.Nil(t, err, "should not have any error inserting rows")
		assert.Equal(t, res.UpdatedRange, NewA1Range("Sheet1!A1:B3"))
		assert.Equal(t, res.UpdatedRows, int64(2))
		assert.Equal(t, res.UpdatedColumns, int64(2))
		assert.Equal(t, res.UpdatedCells, int64(4))
		assert.Equal(t, res.UpdatedValues, values)
	})

	t.Run("http500", func(t *testing.T) {
		expectedParams := map[string]string{
			"includeValuesInResponse":   "true",
			"responseValueRenderOption": responseValueRenderFormatted,
			"valueInputOption":          valueInputUserEntered,
		}
		expectedReqBody := map[string]interface{}{
			"majorDimension": majorDimensionRows,
			"range":          "Sheet1!A1:A2",
			"values": [][]interface{}{
				{"1", "2"},
				{"3", "4"},
			},
		}

		gock.New("https://sheets.googleapis.com").
			Put("/v4/spreadsheets/123/values/Sheet1!A1:A2").
			MatchParams(expectedParams).
			JSON(expectedReqBody).
			Reply(http.StatusInternalServerError)

		values := [][]interface{}{{"1", "2"}, {"3", "4"}}
		res, err := wrapper.UpdateRows(context.Background(), "123", "Sheet1!A1:A2", values)

		assert.NotNil(t, err, "should have error inserting a new row")
		assert.Equal(t, res.UpdatedRange, NewA1Range(""))
		assert.Equal(t, res.UpdatedRows, int64(0))
		assert.Equal(t, res.UpdatedColumns, int64(0))
		assert.Equal(t, res.UpdatedCells, int64(0))
		assert.Equal(t, res.UpdatedValues, [][]interface{}(nil))
	})
}

func TestClear(t *testing.T) {
	path := fixtures.PathToFixture("service_account.json")

	auth, err := auth.NewService(path, []string{}, auth.ServiceConfig{})
	assert.Nil(t, err, "should not have any error instantiating a new service account client")

	wrapper, err := NewWrapper(auth)
	assert.Nil(t, err, "should not have any error instantiating a new sheets wrapper")

	gock.InterceptClient(auth.HTTPClient())

	t.Run("successful", func(t *testing.T) {
		expectedReqBody := map[string][]string{
			"ranges": {"Sheet1!A1:B3", "Sheet1!B4:C5"},
		}
		resp := map[string]interface{}{
			"spreadsheetId": "123",
			"clearedRanges": []string{"Sheet1!A1:B3", "Sheet1!B4:C5"},
		}

		gock.Observe(gock.DumpRequest)
		gock.New("https://sheets.googleapis.com").
			Post("/v4/spreadsheets/123/values:batchClear").
			JSON(expectedReqBody).
			Reply(http.StatusOK).
			JSON(resp)

		res, err := wrapper.Clear(context.Background(), "123", []string{"Sheet1!A1:B3", "Sheet1!B4:C5"})
		assert.Nil(t, err, "should not have any error clearing rows")
		assert.Equal(t, res, []string{"Sheet1!A1:B3", "Sheet1!B4:C5"})
	})

	t.Run("http500", func(t *testing.T) {
		expectedReqBody := map[string][]string{
			"ranges": {"Sheet1!A1:B3", "Sheet1!B4:C5"},
		}

		gock.Observe(gock.DumpRequest)
		gock.New("https://sheets.googleapis.com").
			Post("/v4/spreadsheets/123/values:batchClear").
			JSON(expectedReqBody).
			Reply(http.StatusInternalServerError)

		res, err := wrapper.Clear(context.Background(), "123", []string{"Sheet1!A1:B3", "Sheet1!B4:C5"})
		assert.NotNil(t, err, "should have error clearing rows")
		assert.Equal(t, res, []string(nil))
	})
}
