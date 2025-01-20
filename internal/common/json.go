package common

import (
	"encoding/json"
	"log/slog"
)

func JSONEncodeNoError(v any) string {
	if v == nil {
		return ""
	}

	raw, err := json.Marshal(v)
	if err != nil {
		slog.Error("failed marshalling JSON", "v", v)
		return ""
	}
	return string(raw)
}
