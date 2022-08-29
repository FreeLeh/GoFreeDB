package sheets

import (
	"context"
	"net/http"
	"testing"

	"github.com/FreeLeh/GoFreeDB/google/auth"
	"github.com/FreeLeh/GoFreeDB/internal/google/fixtures"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestCreateSpreadsheet(t *testing.T) {
	path := fixtures.PathToFixture("service_account.json")

	auth, err := auth.NewServiceFromFile(path, []string{}, auth.ServiceConfig{})
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

	auth, err := auth.NewServiceFromFile(path, []string{}, auth.ServiceConfig{})
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

	auth, err := auth.NewServiceFromFile(path, []string{}, auth.ServiceConfig{})
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
		assert.Equal(t, NewA1Range("Sheet1!A1:B3"), res.UpdatedRange)
		assert.Equal(t, int64(2), res.UpdatedRows)
		assert.Equal(t, int64(2), res.UpdatedColumns)
		assert.Equal(t, int64(4), res.UpdatedCells)
		assert.Equal(t, values, res.InsertedValues)
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
		assert.Equal(t, NewA1Range(""), res.UpdatedRange)
		assert.Equal(t, int64(0), res.UpdatedRows)
		assert.Equal(t, int64(0), res.UpdatedColumns)
		assert.Equal(t, int64(0), res.UpdatedCells)
		assert.Equal(t, [][]interface{}(nil), res.InsertedValues)
	})
}

func TestUpdateRows(t *testing.T) {
	path := fixtures.PathToFixture("service_account.json")

	auth, err := auth.NewServiceFromFile(path, []string{}, auth.ServiceConfig{})
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
		assert.Equal(t, NewA1Range("Sheet1!A1:B3"), res.UpdatedRange)
		assert.Equal(t, int64(2), res.UpdatedRows)
		assert.Equal(t, int64(2), res.UpdatedColumns)
		assert.Equal(t, int64(4), res.UpdatedCells)
		assert.Equal(t, values, res.UpdatedValues)
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
		assert.Equal(t, NewA1Range(""), res.UpdatedRange)
		assert.Equal(t, int64(0), res.UpdatedRows)
		assert.Equal(t, int64(0), res.UpdatedColumns)
		assert.Equal(t, int64(0), res.UpdatedCells)
		assert.Equal(t, [][]interface{}(nil), res.UpdatedValues)
	})
}

func TestBatchUpdateRows(t *testing.T) {
	path := fixtures.PathToFixture("service_account.json")

	auth, err := auth.NewServiceFromFile(path, []string{}, auth.ServiceConfig{})
	assert.Nil(t, err, "should not have any error instantiating a new service account client")

	wrapper, err := NewWrapper(auth)
	assert.Nil(t, err, "should not have any error instantiating a new sheets wrapper")

	gock.InterceptClient(auth.HTTPClient())

	t.Run("successful", func(t *testing.T) {
		expectedReqBody := map[string]interface{}{
			"includeValuesInResponse":   true,
			"responseValueRenderOption": responseValueRenderFormatted,
			"valueInputOption":          valueInputUserEntered,
			"data": []map[string]interface{}{
				{
					"majorDimension": majorDimensionRows,
					"range":          "Sheet1!A1:A2",
					"values": [][]interface{}{
						{"VA1"},
						{"VA2"},
					},
				},
				{
					"majorDimension": majorDimensionRows,
					"range":          "Sheet1!B1:B2",
					"values": [][]interface{}{
						{"VB1"},
						{"VB2"},
					},
				},
			},
		}
		resp := map[string]interface{}{
			"spreadsheetId":       "123",
			"totalUpdatedRows":    4,
			"totalUpdatedColumns": 2,
			"totalUpdatedCells":   4,
			"totalUpdatedSheets":  1,
			"responses": []map[string]interface{}{
				{
					"spreadsheetId":  "123",
					"updatedRange":   "Sheet1!A1:A2",
					"updatedRows":    2,
					"updatedColumns": 1,
					"updatedCells":   2,
					"updatedData": map[string]interface{}{
						"range":          "Sheet1!A1:A2",
						"majorDimension": majorDimensionRows,
						"values": [][]interface{}{
							{"VA1"},
							{"VA2"},
						},
					},
				},
				{
					"spreadsheetId":  "123",
					"updatedRange":   "Sheet1!B1:B2",
					"updatedRows":    2,
					"updatedColumns": 1,
					"updatedCells":   2,
					"updatedData": map[string]interface{}{
						"range":          "Sheet1!B1:B2",
						"majorDimension": majorDimensionRows,
						"values": [][]interface{}{
							{"VB1"},
							{"VB2"},
						},
					},
				},
			},
		}

		gock.New("https://sheets.googleapis.com").
			Post("/v4/spreadsheets/123/values:batchUpdate").
			JSON(expectedReqBody).
			Reply(http.StatusOK).
			JSON(resp)

		requests := []BatchUpdateRowsRequest{
			{
				A1Range: "Sheet1!A1:A2",
				Values:  [][]interface{}{{"VA1"}, {"VA2"}},
			},
			{
				A1Range: "Sheet1!B1:B2",
				Values:  [][]interface{}{{"VB1"}, {"VB2"}},
			},
		}
		res, err := wrapper.BatchUpdateRows(context.Background(), "123", requests)

		expected := BatchUpdateRowsResult{
			{
				UpdatedRange:   NewA1Range("Sheet1!A1:A2"),
				UpdatedRows:    2,
				UpdatedColumns: 1,
				UpdatedCells:   2,
				UpdatedValues:  [][]interface{}{{"VA1"}, {"VA2"}},
			},
			{
				UpdatedRange:   NewA1Range("Sheet1!B1:B2"),
				UpdatedRows:    2,
				UpdatedColumns: 1,
				UpdatedCells:   2,
				UpdatedValues:  [][]interface{}{{"VB1"}, {"VB2"}},
			},
		}
		assert.Nil(t, err, "should not have any error inserting rows")
		assert.Equal(t, expected, res)
	})

	t.Run("http500", func(t *testing.T) {
		expectedReqBody := map[string]interface{}{
			"includeValuesInResponse":   true,
			"responseValueRenderOption": responseValueRenderFormatted,
			"valueInputOption":          valueInputUserEntered,
			"data": []map[string]interface{}{
				{
					"majorDimension": majorDimensionRows,
					"range":          "Sheet1!A1:A2",
					"values": [][]interface{}{
						{"VA1"},
						{"VA2"},
					},
				},
				{
					"majorDimension": majorDimensionRows,
					"range":          "Sheet1!B1:B2",
					"values": [][]interface{}{
						{"VB1"},
						{"VB2"},
					},
				},
			},
		}

		gock.New("https://sheets.googleapis.com").
			Put("/v4/spreadsheets/123/values:batchUpdate").
			JSON(expectedReqBody).
			Reply(http.StatusInternalServerError)

		requests := []BatchUpdateRowsRequest{
			{
				A1Range: "Sheet1!A1:A2",
				Values:  [][]interface{}{{"VA1"}, {"VA2"}},
			},
			{
				A1Range: "Sheet1!B1:B2",
				Values:  [][]interface{}{{"VB1"}, {"VB2"}},
			},
		}
		res, err := wrapper.BatchUpdateRows(context.Background(), "123", requests)

		assert.NotNil(t, err, "should have error inserting a new row")
		assert.Equal(t, BatchUpdateRowsResult{}, res)
	})
}

func TestClear(t *testing.T) {
	path := fixtures.PathToFixture("service_account.json")

	auth, err := auth.NewServiceFromFile(path, []string{}, auth.ServiceConfig{})
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

		gock.New("https://sheets.googleapis.com").
			Post("/v4/spreadsheets/123/values:batchClear").
			JSON(expectedReqBody).
			Reply(http.StatusInternalServerError)

		res, err := wrapper.Clear(context.Background(), "123", []string{"Sheet1!A1:B3", "Sheet1!B4:C5"})
		assert.NotNil(t, err, "should have error clearing rows")
		assert.Equal(t, res, []string(nil))
	})
}

func TestQueryRows(t *testing.T) {
	path := fixtures.PathToFixture("service_account.json")

	auth, err := auth.NewServiceFromFile(path, []string{}, auth.ServiceConfig{})
	assert.Nil(t, err, "should not have any error instantiating a new service account client")

	wrapper, err := NewWrapper(auth)
	assert.Nil(t, err, "should not have any error instantiating a new sheets wrapper")

	gock.InterceptClient(auth.HTTPClient())

	t.Run("successful", func(t *testing.T) {
		expectedParams := map[string]string{
			"sheet":   "s1",
			"tqx":     "responseHandler:freedb",
			"tq":      "select A, B",
			"headers": "1",
		}
		resp := `
			/*O_o*/
			freedb({"version":"0.6","reqId":"0","status":"ok","sig":"141753603","table":{"cols":[{"id":"A","label":"","type":"string"},{"id":"B","label":"","type":"number","pattern":"General"}],"rows":[{"c":[{"v":"k1"},{"v":103.51,"f":"103.51"}]},{"c":[{"v":"k2"},{"v":111.0,"f":"111"}]},{"c":[{"v":"k3"},{"v":123.0,"f":"123"}]}],"parsedNumHeaders":0}})
		`

		gock.New("https://docs.google.com").
			Get("/spreadsheets/d/spreadsheetID/gviz/tq").
			MatchParams(expectedParams).
			Reply(http.StatusOK).
			BodyString(resp)

		res, err := wrapper.QueryRows(
			context.Background(),
			"spreadsheetID",
			"s1",
			"select A, B",
			true,
		)
		assert.Nil(t, err)

		expected := QueryRowsResult{Rows: [][]interface{}{
			{"k1", 103.51},
			{"k2", int64(111)},
			{"k3", int64(123)},
		}}
		assert.Equal(t, expected, res)
	})
}
