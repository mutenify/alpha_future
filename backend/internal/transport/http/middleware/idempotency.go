package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"

	"smartcredit/internal/domain"
	"smartcredit/internal/transport/http/httpx"
	"smartcredit/internal/usecase/port"
)

// Idempotency: повтор с тем же Idempotency-Key отдаёт сохранённый ответ,
// тот же ключ с другим телом — 422. Без заголовка запрос идёт как обычно.
func Idempotency(repo port.IdempotencyRepository) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				httpx.Error(w, r, domain.Invalid("не удалось прочитать тело"))
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			hash := requestHash(r, body)
			userID := httpx.Principal(r.Context()).UserID

			stored, err := repo.Get(r.Context(), userID, key)
			switch {
			case err != nil:
				httpx.Error(w, r, err)
			case stored != nil && stored.RequestHash != hash:
				httpx.Error(w, r, domain.ErrIdempotencyMismatch)
			case stored != nil:
				replay(w, *stored)
			default:
				rec := &recorder{ResponseWriter: w, status: http.StatusOK}
				next.ServeHTTP(rec, r)
				if rec.status < http.StatusInternalServerError {
					resp := port.StoredResponse{RequestHash: hash, Status: rec.status, Body: rec.body.Bytes()}
					if err := repo.Save(r.Context(), userID, key, r.Method, r.URL.Path, resp); err != nil {
						slog.Error("save idempotency key", "err", err)
					}
				}
			}
		})
	}
}

func requestHash(r *http.Request, body []byte) string {
	h := sha256.Sum256(append([]byte(r.Method+" "+r.URL.Path+"\n"), body...))
	return hex.EncodeToString(h[:])
}

func replay(w http.ResponseWriter, s port.StoredResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Idempotent-Replay", "true")
	w.WriteHeader(s.Status)
	_, _ = w.Write(s.Body)
}

type recorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (r *recorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
