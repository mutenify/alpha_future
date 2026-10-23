package application

import (
	"context"
	"fmt"
	"io"
	"path"

	"smartcredit/internal/domain"
	"smartcredit/internal/usecase/workflow"
)

type Upload struct {
	DocType  string
	FileName string
	MimeType string
	Body     io.Reader
}

// UploadDocument — прикрепление к анкете (DRAFT) или ответ на запрос (INFO_REQUESTED).
// Ответ на запрос закрывает его и возвращает заявку в SUBMITTED или MANUAL_REVIEW.
func (s *Service) UploadDocument(ctx context.Context, p domain.Principal, id string, up Upload) (domain.Document, error) {
	app, err := s.load(ctx, p, id)
	if err != nil {
		return domain.Document{}, err
	}
	if err := requireStatus(app, domain.StatusDraft, domain.StatusInfoRequested); err != nil {
		return domain.Document{}, err
	}
	key := fmt.Sprintf("%s/%s/%s", app.ID, path.Base(up.DocType), path.Base(up.FileName))
	size, err := s.Storage.Save(ctx, key, up.Body)
	if err != nil {
		return domain.Document{}, err
	}
	var (
		doc    domain.Document
		change domain.StatusChange
	)
	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		open, err := s.Infos.Open(ctx, app.ID)
		if err != nil {
			return err
		}
		d := domain.Document{ApplicationID: app.ID, DocType: up.DocType, FileName: up.FileName,
			StorageKey: key, MimeType: up.MimeType, SizeBytes: size, UploadedBy: p.UserID}
		if open != nil {
			d.InfoRequestID = open.ID
		}
		if doc, err = s.Docs.Create(ctx, d); err != nil || open == nil {
			return err
		}
		return s.answerInfoRequest(ctx, p, &app, *open, &change)
	})
	if err != nil {
		return domain.Document{}, err
	}
	s.Flow.Publish(change)
	return doc, nil
}

func (s *Service) answerInfoRequest(ctx context.Context, p domain.Principal, app *domain.Application, r domain.InfoRequest, change *domain.StatusChange) error {
	if err := s.Infos.Fulfill(ctx, r.ID); err != nil {
		return err
	}
	to := r.ReturnStatus()
	c, err := s.Flow.Move(ctx, app, to, workflow.By(p), "Клиент дозагрузил документы")
	if err != nil {
		return err
	}
	*change = c
	if to == domain.StatusSubmitted {
		return s.Jobs.Enqueue(ctx, app.ID, domain.StepCompleteness)
	}
	return nil
}

func (s *Service) Documents(ctx context.Context, p domain.Principal, id string) ([]domain.Document, error) {
	if _, err := s.load(ctx, p, id); err != nil {
		return nil, err
	}
	return s.Docs.ListByApplication(ctx, id)
}
