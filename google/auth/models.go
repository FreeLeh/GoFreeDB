package auth

type Scopes []string

var (
	GoogleSheetsReadOnly  Scopes = []string{"https://www.googleapis.com/auth/spreadsheets.readonly"}
	GoogleSheetsWriteOnly Scopes = []string{"https://www.googleapis.com/auth/spreadsheets"}
	GoogleSheetsReadWrite Scopes = []string{"https://www.googleapis.com/auth/spreadsheets"}
)
