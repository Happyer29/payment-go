package services

import (
	"fmt"
	"log"
	"payment-go/internal/models"
	"payment-go/internal/repositories"
	"sync"
	"time"
)

type IPaymentLinkService interface {
	LoadPendingLinks() []error
	FinishLinkWithStatus(link *models.PaymentLink, ord *models.Order, status string) error
}
type paymentLinkService struct {
}

var plIns IPaymentLinkService
var plOnce = sync.Once{}

func PaymentLinkService() IPaymentLinkService {
	plOnce.Do(func() {
		plIns = &paymentLinkService{}
	})
	return plIns
}

func (s *paymentLinkService) LoadPendingLinks() []error {
	links, err := repositories.PaymentLinkRepository().GetPending()
	if err != nil {
		return []error{err}
	}

	if len(links) == 0 {
		fmt.Println("No pending links to load")
	}

	var errors []error
	for _, link := range links {
		BankApiService().TaskCheckLink(link, nil)
	}
	return errors
}

func (s *paymentLinkService) FinishLinkWithStatus(link *models.PaymentLink, ord *models.Order, status string) error {
	if !models.IsStatusValid(status) {
		return fmt.Errorf("invalid status on finishing payment link #%d for order #%d", link.ID, ord.ID)
	}

	now := uint(time.Now().Unix())
	if err := OrderService().FinishOrderWithStatus(ord, status, now); err != nil {
		return err
	}

	link.Status = status
	link.DatePaid = &now
	if err := repositories.PaymentLinkRepository().Save(link); err != nil {
		log.Println(err)
	}

	return nil
}
