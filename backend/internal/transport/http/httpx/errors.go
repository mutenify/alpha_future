package httpx

import (
	"errors"
	"log/slog"
	"net/http"

	"smartcredit/internal/domain"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error переводит доменную ошибку в HTTP-ответ формата components/schemas/Error.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	status, code := classify(err)
	msg := err.Error()
	if status == http.StatusInternalServerError {
		slog.ErrorContext(r.Context(), "request failed", "path", r.URL.Path, "err", err)
		msg = "внутренняя ошибка"
	}
	JSON(w, status, errorBody{Code: code, Message: msg})
}

func classify(err error) (int, string) {
	var ve domain.ValidationError
	switch {
	case errors.As(err, &ve):
		return http.StatusBadRequest, "VALIDATION_ERROR"
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, "UNAUTHORIZED"
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN"
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, domain.ErrInvalidTransition):
		return http.StatusConflict, "INVALID_TRANSITION"
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, "CONFLICT"
	case errors.Is(err, domain.ErrIdempotencyMismatch):
		return http.StatusUnprocessableEntity, "IDEMPOTENCY_MISMATCH"
	case errors.Is(err, domain.ErrSourceUnavailable):
		return http.StatusServiceUnavailable, "SOURCE_UNAVAILABLE"
	}
	return http.StatusInternalServerError, "INTERNAL"
}
