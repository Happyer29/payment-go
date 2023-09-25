package services

import (
	"fmt"
	"log"
	"payment-go/internal/config"
	"payment-go/internal/models"
	"payment-go/internal/repositories"
	"payment-go/internal/transport/model/order"
	"sync"
)

type IEventService interface {
	OrderCreated(ord *models.Order, dto *order.CreateOrderDto)
	OrderCompleted(ord *models.Order)
	OrderUpdated(old *models.Order, new *models.Order)

	CardBalanceIncreased(crd *models.Card)
	NoCardsAvailable(paymentMethod *string)

}
type eventService struct {
}

var evIns *eventService
var evOnce = sync.Once{}

func EventService() IEventService {
	evOnce.Do(func() {
		evIns = &eventService{}
	})
	return evIns
}

func (s *eventService) OrderCompleted(ord *models.Order) {
	go WebhookService().SendOrderCompleted(ord, nil)

	if ord == nil || ord.Status != models.StatusCompleted {
		return
	}

	link := config.GetConfig().AppHost + fmt.Sprintf("/admin/order/%d/edit", ord.ID)
	text := fmt.Sprintf(
		"Order #%d in shop #%d is completed.\nOrder amount: %.02f.\nUsed card: %s.\nLink: %s",
		ord.ID, ord.ShopID, ord.Amount, ord.Card.GetIdentifier(), link,
	)

}

func (s *eventService) OrderUpdated(old *models.Order, new *models.Order) {
	// если статус заказа был изменён
	if !old.IsFinished() && new.IsFinished() {
		go s.OrderCompleted(new)
	}
}

func (s *eventService) OrderCreated(ord *models.Order, dto *order.CreateOrderDto) {
	fmt.Sprintf("Event.OrderCreated %#v\n", ord.ID)

	// если для оплаты необходимо сгенерировать ссылку
	if ord.HaveLink() {

		// поставим задачу на фоновое создание ссылки
		BankApiService().TaskCreateLink(ord, func(link *models.PaymentLink) {

			// если передан вебхук, то отсылаем на него инфу о создании ссылки
			WebhookService().SendLinkCreated(link, nil)

			if link.Status == models.StatusFailed {
				return
			}

			// если ссылка создалась, запускаем таск на проверку ссылки
			BankApiService().TaskCheckLink(link, func(link *models.PaymentLink) {
				WebhookService().SendOrderCompleted(ord, nil)
			})
		})
	} else {
		ord.Status = models.StatusPending
		err := repositories.OrderRepository().Save(ord)
		if err != nil {
			log.Println("Event.OrderCreated: no-link order save error.", err)
		}
	}

}

func (s *eventService) CardBalanceIncreased(crd *models.Card) {
	info, err := repositories.CardRepository().GetCardInfo(crd.ID)
	if err != nil {
		return
	}

	// если надо заблокировать карту
	if info.Balance >= config.CardDisableAmount && crd.Status == models.CardStatusEnabled {
		crd.Status = models.CardStatusDisabled
		err = repositories.CardRepository().Save(crd)
		if err == nil {
			text := fmt.Sprintf(
				"Card #%d will be deactivated soon. Card balance: %.2f",
				crd.ID, info.Balance,
			)
		}
	}
}

func (s *eventService) NoCardsAvailable(paymentMethod *string) {
	var text string
	if paymentMethod != nil {
		desc := *paymentMethod
		if *paymentMethod == config.PaymentMethodBankTransfer {
			desc = "Card"
		} else if *paymentMethod == config.PaymentMethodKapitalBank {
			desc = "Kapital Bank link"
		}
		text = "No cards available for payment method: " + desc + " !!!"
	} else {
		text = "No cards available !!!"
	}
}

