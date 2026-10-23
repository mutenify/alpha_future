package httpx

import (
	"encoding/json"
	"net/http"
	"strconv"

	"smartcredit/internal/domain"
)

const maxBody = 1 << 20

func Decode(r *http.Request, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, maxBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return domain.Invalid("некорректное тело запроса: " + err.Error())
	}
	return nil
}

func QueryInt(r *http.Request, name string) int64 {
	v, _ := strconv.ParseInt(r.URL.Query().Get(name), 10, 64)
	return v
}
