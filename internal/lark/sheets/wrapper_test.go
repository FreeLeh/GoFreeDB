package sheets

import (
	"context"
	"fmt"
	"go.uber.org/mock/gomock"
	"net/http"
	"testing"

	"github.com/FreeLeh/GoFreeDB/internal/models"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

const testSpreadsheetToken = "spreadsheet123"

func TestCreateSheet(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		defer gock.Off()

		ctrl := gomock.NewController(t)
		mockAuth := NewMockAccessTokenGetter(ctrl)
		wrapper := NewWrapper(mockAuth)

		mockAuth.EXPECT().AccessToken().
			Return("access_token", nil).
			Times(1)

		expectedReq := map[string]interface{}{
			"requests": []map[string]interface{}{
				{
					"addSheet": map[string]interface{}{
						"properties": map[string]interface{}{
							"title": "sheet",
						},
					},
				},
			},
		}

		gock.New(fmt.Sprintf(createSheetURL, testSpreadsheetToken)).
			Post("").
			JSON(expectedReq).
			Reply(http.StatusOK).
			JSON(map[string]interface{}{
				"code": apiStatusCodeOK,
			})

		err := wrapper.CreateSheet(context.Background(), testSpreadsheetToken, "sheet")
		assert.Nil(t, err)
	})

	t.Run("http_error", func(t *testing.T) {
		defer gock.Off()

		ctrl := gomock.NewController(t)
		mockAuth := NewMockAccessTokenGetter(ctrl)
		wrapper := NewWrapper(mockAuth)

		mockAuth.EXPECT().AccessToken().
			Return("access_token", nil).
			Times(1)

		gock.New(fmt.Sprintf(createSheetURL, testSpreadsheetToken)).
			Post("").
			Reply(http.StatusInternalServerError)

		err := wrapper.CreateSheet(context.Background(), "spreadsheet123", "sheet")
		assert.NotNil(t, err)
	})

	t.Run("api_error", func(t *testing.T) {
		defer gock.Off()

		ctrl := gomock.NewController(t)
		mockAuth := NewMockAccessTokenGetter(ctrl)
		wrapper := NewWrapper(mockAuth)

		mockAuth.EXPECT().AccessToken().
			Return("access_token", nil).
			Times(1)

		gock.New(fmt.Sprintf(createSheetURL, testSpreadsheetToken)).
			Post("").
			Reply(http.StatusOK).
			JSON(map[string]interface{}{
				"code": 500,
				"msg":  "internal error",
			})

		err := wrapper.CreateSheet(context.Background(), "spreadsheet123", "sheet")
		assert.NotNil(t, err)
	})
}

func TestInsertRows(t *testing.T) {
	t.Run("successful_overwrite", func(t *testing.T) {
		defer gock.Off()

		ctrl := gomock.NewController(t)
		mockAuth := NewMockAccessTokenGetter(ctrl)
		wrapper := NewWrapper(mockAuth)
		gock.InterceptClient(wrapper.httpClient)

		mockAuth.EXPECT().AccessToken().
			Return("access_token", nil).
			Times(1)
		defer gock.Off()

		expectedReq := map[string]interface{}{
			"valueRange": map[string]interface{}{
				"range":  "Sheet1!A1:B2",
				"values": [][]interface{}{{"1", "2"}, {"3", "4"}},
			},
		}

		gock.New(fmt.Sprintf(appendValuesURL, testSpreadsheetToken, appendModeOverwrite)).
			Post("").
			MatchParam("insertDataOption", string(appendModeOverwrite)).
			JSON(expectedReq).
			Reply(http.StatusOK).
			JSON(map[string]interface{}{
				"code": apiStatusCodeOK,
				"data": map[string]interface{}{
					"updates": map[string]interface{}{
						"updatedRange":   "Sheet1!A1:B2",
						"updatedRows":    2,
						"updatedColumns": 2,
						"updatedCells":   4,
					},
				},
			})

		values := [][]interface{}{{"1", "2"}, {"3", "4"}}
		res, err := wrapper.OverwriteRows(
			context.Background(),
			"spreadsheet123",
			models.NewA1RangeFromString("Sheet1!A1:B2"),
			values,
		)

		assert.Nil(t, err)
		assert.Equal(t, models.NewA1RangeFromString("Sheet1!A1:B2"), res.UpdatedRange)
		assert.Equal(t, int64(2), res.UpdatedRows)
		assert.Equal(t, int64(2), res.UpdatedColumns)
		assert.Equal(t, int64(4), res.UpdatedCells)
	})

	t.Run("http_error", func(t *testing.T) {
		defer gock.Off()

		ctrl := gomock.NewController(t)
		mockAuth := NewMockAccessTokenGetter(ctrl)
		wrapper := NewWrapper(mockAuth)
		gock.InterceptClient(wrapper.httpClient)

		mockAuth.EXPECT().AccessToken().
			Return("access_token", nil).
			Times(1)

		gock.New(fmt.Sprintf(appendValuesURL, testSpreadsheetToken, appendModeOverwrite)).
			Post("").
			Reply(http.StatusInternalServerError)

		values := [][]interface{}{{"1", "2"}}
		_, err := wrapper.OverwriteRows(
			context.Background(),
			"spreadsheet123",
			models.NewA1RangeFromString("Sheet1!A1"),
			values,
		)

		assert.NotNil(t, err)
	})

	t.Run("api_error", func(t *testing.T) {
		defer gock.Off()

		ctrl := gomock.NewController(t)
		mockAuth := NewMockAccessTokenGetter(ctrl)
		wrapper := NewWrapper(mockAuth)
		gock.InterceptClient(wrapper.httpClient)

		mockAuth.EXPECT().AccessToken().
			Return("access_token", nil).
			Times(1)

		gock.New(fmt.Sprintf(appendValuesURL, testSpreadsheetToken, appendModeOverwrite)).
			Post("").
			Reply(http.StatusOK).
			JSON(map[string]interface{}{
				"code": 500,
				"msg":  "internal error",
			})

		values := [][]interface{}{{"1", "2"}}
		_, err := wrapper.OverwriteRows(
			context.Background(),
			"spreadsheet123",
			models.NewA1RangeFromString("Sheet1!A1"),
			values,
		)

		assert.NotNil(t, err)
	})
}

func TestBatchUpdateRows(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		defer gock.Off()

		ctrl := gomock.NewController(t)
		mockAuth := NewMockAccessTokenGetter(ctrl)
		wrapper := NewWrapper(mockAuth)
		gock.InterceptClient(wrapper.httpClient)

		mockAuth.EXPECT().AccessToken().
			Return("access_token", nil).
			Times(1)
		defer gock.Off()

		expectedReq := []map[string]interface{}{
			{
				"range":  "Sheet1!A1:B2",
				"values": [][]interface{}{{"1", "2"}, {"3", "4"}},
			},
			{
				"range":  "Sheet1!C1:D2",
				"values": [][]interface{}{{"5", "6"}, {"7", "8"}},
			},
		}

		gock.New(fmt.Sprintf(batchUpdateRowsURL, testSpreadsheetToken)).
			Post("").
			JSON(expectedReq).
			Reply(http.StatusOK).
			JSON(map[string]interface{}{
				"code": apiStatusCodeOK,
			})

		requests := []BatchUpdateRowsRequest{
			{
				A1Range: models.NewA1RangeFromString("Sheet1!A1:B2"),
				Values:  [][]interface{}{{"1", "2"}, {"3", "4"}},
			},
			{
				A1Range: models.NewA1RangeFromString("Sheet1!C1:D2"),
				Values:  [][]interface{}{{"5", "6"}, {"7", "8"}},
			},
		}

		err := wrapper.BatchUpdateRows(context.Background(), "spreadsheet123", requests)
		assert.Nil(t, err)
	})

	t.Run("http_error", func(t *testing.T) {
		defer gock.Off()

		ctrl := gomock.NewController(t)
		mockAuth := NewMockAccessTokenGetter(ctrl)
		wrapper := NewWrapper(mockAuth)
		gock.InterceptClient(wrapper.httpClient)

		mockAuth.EXPECT().AccessToken().
			Return("access_token", nil).
			Times(1)

		gock.New(fmt.Sprintf(batchUpdateRowsURL, testSpreadsheetToken)).
			Post("").
			Reply(http.StatusInternalServerError)

		requests := []BatchUpdateRowsRequest{
			{
				A1Range: models.NewA1RangeFromString("Sheet1!A1"),
				Values:  [][]interface{}{{"1"}},
			},
		}

		err := wrapper.BatchUpdateRows(context.Background(), "spreadsheet123", requests)
		assert.NotNil(t, err)
	})

	t.Run("api_error", func(t *testing.T) {
		defer gock.Off()

		ctrl := gomock.NewController(t)
		mockAuth := NewMockAccessTokenGetter(ctrl)
		wrapper := NewWrapper(mockAuth)
		gock.InterceptClient(wrapper.httpClient)

		mockAuth.EXPECT().AccessToken().
			Return("access_token", nil).
			Times(1)

		gock.New(fmt.Sprintf(batchUpdateRowsURL, testSpreadsheetToken)).
			Post("").
			Reply(http.StatusOK).
			JSON(map[string]interface{}{
				"code": 500,
				"msg":  "internal error",
			})

		requests := []BatchUpdateRowsRequest{
			{
				A1Range: models.NewA1RangeFromString("Sheet1!A1"),
				Values:  [][]interface{}{{"1"}},
			},
		}

		err := wrapper.BatchUpdateRows(context.Background(), "spreadsheet123", requests)
		assert.NotNil(t, err)
	})
}

func TestClear(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		defer gock.Off()

		ctrl := gomock.NewController(t)
		mockAuth := NewMockAccessTokenGetter(ctrl)
		wrapper := NewWrapper(mockAuth)
		gock.InterceptClient(wrapper.httpClient)

		mockAuth.EXPECT().AccessToken().
			Return("access_token", nil).
			Times(1)
		defer gock.Off()

		expectedReq := []map[string]interface{}{
			{
				"range":  "Sheet1!A1:B2",
				"values": [][]interface{}{{"", ""}, {"", ""}},
			},
		}

		gock.New(fmt.Sprintf(batchUpdateRowsURL, testSpreadsheetToken)).
			Post("").
			JSON(expectedReq).
			Reply(http.StatusOK).
			JSON(map[string]interface{}{
				"code": apiStatusCodeOK,
			})

		ranges := []models.A1Range{
			models.NewA1RangeFromString("Sheet1!A1:B2"),
		}

		err := wrapper.Clear(context.Background(), "spreadsheet123", ranges)
		assert.Nil(t, err)
	})

	t.Run("http_error", func(t *testing.T) {
		defer gock.Off()

		ctrl := gomock.NewController(t)
		mockAuth := NewMockAccessTokenGetter(ctrl)
		wrapper := NewWrapper(mockAuth)
		gock.InterceptClient(wrapper.httpClient)

		mockAuth.EXPECT().AccessToken().
			Return("access_token", nil).
			Times(1)

		gock.New(fmt.Sprintf(batchUpdateRowsURL, testSpreadsheetToken)).
			Post("").
			Reply(http.StatusInternalServerError)

		ranges := []models.A1Range{
			models.NewA1RangeFromString("Sheet1!A1"),
		}

		err := wrapper.Clear(context.Background(), "spreadsheet123", ranges)
		assert.NotNil(t, err)
	})

	t.Run("api_error", func(t *testing.T) {
		defer gock.Off()

		ctrl := gomock.NewController(t)
		mockAuth := NewMockAccessTokenGetter(ctrl)
		wrapper := NewWrapper(mockAuth)
		gock.InterceptClient(wrapper.httpClient)

		mockAuth.EXPECT().AccessToken().
			Return("access_token", nil).
			Times(1)

		gock.New(fmt.Sprintf(batchUpdateRowsURL, testSpreadsheetToken)).
			Post("").
			Reply(http.StatusOK).
			JSON(map[string]interface{}{
				"code": 500,
				"msg":  "internal error",
			})

		ranges := []models.A1Range{
			models.NewA1RangeFromString("Sheet1!A1"),
		}

		err := wrapper.Clear(context.Background(), "spreadsheet123", ranges)
		assert.NotNil(t, err)
	})
}

func TestGetSheets(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		defer gock.Off()

		ctrl := gomock.NewController(t)
		mockAuth := NewMockAccessTokenGetter(ctrl)
		wrapper := NewWrapper(mockAuth)
		gock.InterceptClient(wrapper.httpClient)

		mockAuth.EXPECT().AccessToken().
			Return("access_token", nil).
			Times(1)
		defer gock.Off()

		gock.New(fmt.Sprintf(getSheetsURL, testSpreadsheetToken)).
			Get("").
			Reply(http.StatusOK).
			JSON(map[string]interface{}{
				"code": apiStatusCodeOK,
				"data": map[string]interface{}{
					"sheets": []map[string]interface{}{
						{
							"sheet_id": "sheet1",
							"title":    "Sheet1",
						},
					},
				},
			})

		res, err := wrapper.GetSheets(context.Background(), "spreadsheet123")
		assert.Nil(t, err)
		assert.Equal(t, 1, len(res.Sheets))
		assert.Equal(t, "sheet1", res.Sheets[0].SheetID)
		assert.Equal(t, "Sheet1", res.Sheets[0].Title)
	})

	t.Run("http_error", func(t *testing.T) {
		defer gock.Off()

		ctrl := gomock.NewController(t)
		mockAuth := NewMockAccessTokenGetter(ctrl)
		wrapper := NewWrapper(mockAuth)
		gock.InterceptClient(wrapper.httpClient)

		mockAuth.EXPECT().AccessToken().
			Return("access_token", nil).
			Times(1)

		gock.New(fmt.Sprintf(getSheetsURL, testSpreadsheetToken)).
			Get("").
			Reply(http.StatusInternalServerError)

		_, err := wrapper.GetSheets(context.Background(), "spreadsheet123")
		assert.NotNil(t, err)
	})

	t.Run("api_error", func(t *testing.T) {
		defer gock.Off()

		ctrl := gomock.NewController(t)
		mockAuth := NewMockAccessTokenGetter(ctrl)
		wrapper := NewWrapper(mockAuth)
		gock.InterceptClient(wrapper.httpClient)

		mockAuth.EXPECT().AccessToken().
			Return("access_token", nil).
			Times(1)

		gock.New(fmt.Sprintf(getSheetsURL, testSpreadsheetToken)).
			Get("").
			Reply(http.StatusOK).
			JSON(map[string]interface{}{
				"code": 500,
				"msg":  "internal error",
			})

		_, err := wrapper.GetSheets(context.Background(), "spreadsheet123")
		assert.NotNil(t, err)
	})
}
