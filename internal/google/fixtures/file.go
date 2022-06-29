package fixtures

import (
	"path/filepath"
	"runtime"
)

// PathToFixture returns the absolute path to the given fixtureFileName.
// This is a helper function to make it easier load some test fixture files.
func PathToFixture(fixtureFileName string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), fixtureFileName)
}
