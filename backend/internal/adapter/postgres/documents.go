package postgres

import (
	"context"

	"smartcredit/internal/domain"
)

type Documents struct{ *DB }

const docCols = `id, application_id, coalesce(info_request_id::text, ''), doc_type, file_name, storage_key,
	mime_type, size_bytes, uploaded_by, uploaded_at`

func (r Documents) Create(ctx context.Context, d domain.Document) (domain.Document, error) {
	return scanDoc(r.q(ctx).QueryRow(ctx, `
		INSERT INTO documents (application_id, info_request_id, doc_type, file_name, storage_key, mime_type, size_bytes, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING `+docCols,
		d.ApplicationID, nullable(d.InfoRequestID), d.DocType, d.FileName, d.StorageKey, d.MimeType, d.SizeBytes, d.UploadedBy))
}

func (r Documents) ListByApplication(ctx context.Context, appID string) ([]domain.Document, error) {
	rows, err := r.q(ctx).Query(ctx, `SELECT `+docCols+` FROM documents WHERE application_id = $1 ORDER BY uploaded_at`, appID)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanDoc)
}

func scanDoc(row scanner) (domain.Document, error) {
	var d domain.Document
	err := row.Scan(&d.ID, &d.ApplicationID, &d.InfoRequestID, &d.DocType, &d.FileName, &d.StorageKey,
		&d.MimeType, &d.SizeBytes, &d.UploadedBy, &d.UploadedAt)
	return d, err
}
