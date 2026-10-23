package dto

import (
	"time"

	"smartcredit/internal/domain"
)

type InfoRequest struct {
	ID          string            `json:"id"`
	Origin      domain.InfoOrigin `json:"origin"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
}

func FromInfoRequest(r domain.InfoRequest) InfoRequest {
	return InfoRequest{r.ID, r.Origin, r.Description, r.Status}
}

type Document struct {
	ID         string    `json:"id"`
	DocType    string    `json:"docType"`
	FileName   string    `json:"fileName"`
	UploadedAt time.Time `json:"uploadedAt"`
}

func FromDocument(d domain.Document) Document {
	return Document{d.ID, d.DocType, d.FileName, d.UploadedAt}
}

func FromDocuments(ds []domain.Document) []Document { return mapAll(ds, FromDocument) }
