package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"smartcredit/internal/transport/http/dto"
	"smartcredit/internal/transport/http/httpx"
)

// StreamEvents — SSE со сменой статуса. Доступ проверяется так же, как в GET /applications/{id}.
func (h *Handler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := h.Apps.Get(r.Context(), httpx.Principal(r.Context()), id); err != nil {
		httpx.Error(w, r, err)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		httpx.Error(w, r, fmt.Errorf("streaming unsupported"))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	flusher.Flush()

	ch, cancel := h.Events.Subscribe(id)
	defer cancel()
	for {
		select {
		case <-r.Context().Done():
			return
		case c := <-ch:
			data, _ := json.Marshal(dto.FromChange(c))
			fmt.Fprintf(w, "event: status\ndata: %s\n\n", data)
			flusher.Flush()
		}
	}
}
