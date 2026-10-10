package smsservice

import (
	"context"
	"smsgateway/entity"
	"smsgateway/param/smsparam"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/richerror"
	"time"
)

func (s SMSService) ReportSMS(ctx context.Context, req smsparam.ReportSMSRequest) (smsparam.ReportSMSResponse, error) {
	const op = "SMSService.ReportSMS"

	page := req.Page
	if page < 1 {
		page = smsparam.DefaultReportPage
	}

	perPage := req.PerPage
	if perPage < 1 {
		perPage = smsparam.DefaultReportPerPage
	}

	if perPage > smsparam.MaxReportPerPage {
		perPage = smsparam.MaxReportPerPage
	}

	filter := entity.MessageFilter{
		UserID:         req.UserID,
		ReceptorNumber: req.Receptor,
		Status:         req.Status,
		Type:           req.SmsType,
		Limit:          perPage,
		Offset:         (page - 1) * perPage,
	}

	if req.From != "" {
		from, err := time.Parse(smsparam.ReportDateLayout, req.From)
		if err != nil {
			return smsparam.ReportSMSResponse{}, richerror.New(op).WithKind(richerror.KindInvalidInput).
				WithErr(err).WithMessage(errmsg.ErrorMsgInvalidInput)
		}

		filter.CreatedFrom = from.Unix()
	}

	if req.To != "" {
		to, err := time.Parse(smsparam.ReportDateLayout, req.To)
		if err != nil {
			return smsparam.ReportSMSResponse{}, richerror.New(op).WithKind(richerror.KindInvalidInput).
				WithErr(err).WithMessage(errmsg.ErrorMsgInvalidInput)
		}

		// the To date is inclusive, so filter up to the start of the next day
		filter.CreatedTo = to.AddDate(0, 0, 1).Unix()
	}

	messages, total, err := s.smsRepo.ListMessages(ctx, filter)
	if err != nil {
		return smsparam.ReportSMSResponse{}, err
	}

	items := make([]smsparam.ReportSMSItem, 0, len(messages))
	for _, message := range messages {
		items = append(items, smsparam.ReportSMSItem{
			ID:             message.ID,
			UserID:         message.UserID,
			SourceNumber:   message.SourceNumber,
			ReceptorNumber: message.ReceptorNumber,
			Content:        message.Content,
			Type:           message.Type,
			Status:         message.Status,
			FailedReason:   message.FailedReason,
			CreatedAt:      message.CreatedAt,
		})
	}

	return smsparam.ReportSMSResponse{
		Items: items,
		Pagination: smsparam.Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: (total + perPage - 1) / perPage,
		},
	}, nil
}
