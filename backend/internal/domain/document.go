package domain

import "time"

type Document struct {
	ID            string
	ApplicationID string
	InfoRequestID string
	DocType       string
	FileName      string
	StorageKey    string
	MimeType      string
	SizeBytes     int64
	UploadedBy    string
	UploadedAt    time.Time
}
