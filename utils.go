package freeleh

import (
	"time"
)

func currentTimeMs() int64 {
	return time.Now().UnixMilli()
}

func getA1Range(sheetName string, rng string) string {
	return sheetName + "!" + rng
}
