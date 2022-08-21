package auth

// Scopes encapsulates a list of Google resources scopes to request during authentication step.
type Scopes []string

var (
	GoogleSheetsReadOnly  Scopes = []string{"https://www.googleapis.com/auth/spreadsheets.readonly"}
	GoogleSheetsWriteOnly Scopes = []string{"https://www.googleapis.com/auth/spreadsheets"}
	GoogleSheetsReadWrite Scopes = []string{"https://www.googleapis.com/auth/spreadsheets"}
)
