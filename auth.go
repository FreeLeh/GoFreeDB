package freedb

import "github.com/FreeLeh/GoFreeDB/google/auth"

// GoogleAuthScopes specifies the list of Google Auth scopes required to run FreeDB implementations properly.
var (
	GoogleAuthScopes = auth.GoogleSheetsReadWrite
)
