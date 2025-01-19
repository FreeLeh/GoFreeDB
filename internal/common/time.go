package common

import "time"

func CurrentTimeMs() int64 {
	return time.Now().UnixMilli()
}
