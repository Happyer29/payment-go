package services

import (
	"errors"
	"payment-go/internal/models"
	"payment-go/internal/repositories"
	"payment-go/internal/transport/model/bank_message"
	"sync"
	"time"
)

type IBankMessageService interface {
	ProcessApproval(dto *bank_message.BankMessageApprovalDto, approved bool) error
}
type bankMessageService struct {
}

var bmIns *bankMessageService
var bmOnce = sync.Once{}

var ErrMessageNotFound = errors.New("message not found")
var ErrUnableToApprove = errors.New("unable to approve")
var ErrUnableToDecline = errors.New("unable to decline")
var ErrWhileSaving = errors.New("error while saving")
var ErrWhileFinishingOrder = errors.New("error while finishing order")
var ErrOrderNotFound = errors.New("order not found")
var ErrInvalidAmount = errors.New("invalid amount passed")

func BankMessageService() IBankMessageService {
	bmOnce.Do(func() {
		bmIns = &bankMessageService{}
	})
	return bmIns
}

func (s *bankMessageService) ProcessApproval(dto *bank_message.BankMessageApprovalDto, approved bool) error {
	// get message
	msg, err := repositories.BankMessageRepository().FindById(dto.MessageID)
	if err != nil {
		return ErrMessageNotFound
	}

	// get order
	ord, err := repositories.OrderRepository().FindById(msg.OrderID)
	if err != nil {
		return ErrOrderNotFound
	}

	if approved {
		if !msg.Approve() {
			return ErrUnableToApprove
		}
		if dto.Amount > 0 {
			ord.Amount = dto.Amount
		} else {
			return ErrInvalidAmount
		}
	} else if !msg.Decline() {
		return ErrUnableToDecline
	}
	if err := repositories.BankMessageRepository().Save(msg); err != nil {
		return ErrWhileSaving
	}

	var status string
	if approved {
		ord.Amount = dto.Amount
		status = models.StatusCompleted
	} else {
		status = models.StatusFailed
	}
	if err := OrderService().FinishOrderWithStatus(ord, status, uint(time.Now().Unix())); err != nil {
		return ErrWhileFinishingOrder
	}

	return nil
}
