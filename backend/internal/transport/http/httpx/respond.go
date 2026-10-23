// Package httpx — общие помощники HTTP-слоя: ответы, ошибки, разбор запроса, принципал.
package httpx

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func OK(w http.ResponseWriter, v any) { JSON(w, http.StatusOK, v) }
