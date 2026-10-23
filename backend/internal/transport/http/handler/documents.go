package handler

import (
	"net/http"

	"smartcredit/internal/domain"
	"smartcredit/internal/transport/http/dto"
	"smartcredit/internal/transport/http/httpx"
	"smartcredit/internal/usecase/application"
)

const maxUpload = 20 << 20

func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		httpx.Error(w, r, domain.Invalid("нужна multipart-форма с полями file и docType"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.Error(w, r, domain.Invalid("нет файла"))
		return
	}
	defer file.Close()
	doc, err := h.Apps.UploadDocument(r.Context(), httpx.Principal(r.Context()), r.PathValue("id"), application.Upload{
		DocType:  r.FormValue("docType"),
		FileName: header.Filename,
		MimeType: header.Header.Get("Content-Type"),
		Body:     file,
	})
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, dto.FromDocument(doc))
}

func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	docs, err := h.Apps.Documents(r.Context(), httpx.Principal(r.Context()), r.PathValue("id"))
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromDocuments(docs))
}
