package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"payment-go/internal/models"
	order2 "payment-go/internal/transport/webhook/order"
	"strings"
	"sync"
)

type IWebhookService interface {
	GetOrderCreatedWebhook(link *string, ord *models.Order) *url.URL
	GetOrderOnSuccessWebhook(link *string, ord *models.Order) *url.URL
	GetOrderOnFailureWebhook(link *string, ord *models.Order) *url.URL

	SendOrderCompleted(ord *models.Order, webhook *string)
	SendLinkCreated(link *models.PaymentLink, webhook *string)
}
type webhookService struct {
}

var whIns *webhookService
var whOnce = sync.Once{}

func WebhookService() IWebhookService {
	whOnce.Do(func() {
		whIns = &webhookService{}
	})
	return whIns
}

func (s *webhookService) GetOrderCreatedWebhook(link *string, ord *models.Order) *url.URL {
	sh := ord.Shop
	//if !sh.IsAvailable() || !sh.HostValidated {
	//	return nil
	//}

	// если вебхук передан в запросе
	return s.getWebhookOrDefault(link, sh.Host, sh.Webhooks.LinkCreated)
}

func (s *webhookService) GetOrderOnSuccessWebhook(link *string, ord *models.Order) *url.URL {
	sh := ord.Shop
	//if !sh.IsAvailable() || !sh.HostValidated {
	//	return nil
	//}

	// если вебхук передан в запросе
	return s.getWebhookOrDefault(link, sh.Host, sh.Webhooks.OnSuccess)
}

func (s *webhookService) GetOrderOnFailureWebhook(link *string, ord *models.Order) *url.URL {
	sh := ord.Shop
	//if !sh.IsAvailable() || !sh.HostValidated {
	//	return nil
	//}

	// если вебхук передан в запросе
	return s.getWebhookOrDefault(link, sh.Host, sh.Webhooks.OnFailure)
}

func (s *webhookService) getWebhookOrDefault(webhook *string, host string, def *string) *url.URL {
	if webhook != nil && len(*webhook) != 0 && (*webhook)[0] == '/' {
		link := host + *webhook
		u, err := url.Parse(link)
		if err != nil {
			return nil
		}
		return u
	}

	// дефолтный вебхук
	if def == nil {
		return nil
	}

	u, err := url.Parse(host + *def)
	if err != nil {
		return nil
	}

	return u
}

func (s *webhookService) SendLinkCreated(link *models.PaymentLink, webhook *string) {
	success := link.Status == models.StatusCompleted
	ord := &link.Order

	u := s.GetOrderCreatedWebhook(webhook, ord)
	if u == nil {
		return
	}

	// собираем информацию в json
	data, err := json.Marshal(&order2.WebhookOrderCreatedDto{
		Success:     success,
		OrderNumber: ord.Number.String(),
		PaymentLink: link.URL,
	})
	if err != nil {
		return
	}

	// отсылаем вебхук
	cl := http.Client{}
	reader := strings.NewReader(string(data))
	cl.Post(u.String(), "application/json", reader)
}

func (s *webhookService) SendOrderCompleted(ord *models.Order, webhook *string) {
	success := ord.Status == models.StatusCompleted

	// получаем нужный вебхук
	var u *url.URL = nil
	if success {
		u = s.GetOrderOnSuccessWebhook(webhook, ord)
	} else {
		u = s.GetOrderOnFailureWebhook(webhook, ord)
	}
	if u == nil {
		return
	}

	// собираем информацию в json
	data, err := json.Marshal(&order2.WebhookOrderCompletedDto{
		Success:     success,
		OrderNumber: ord.Number.String(),
		Amount:      ord.Amount,
	})
	if err != nil {
		fmt.Printf("%#v", data)
		return
	}

	// отсылаем вебхук
	cl := http.Client{}
	reader := strings.NewReader(string(data))
	_, err = cl.Post(u.String(), "application/json", reader)
}

func (s *webhookService) SendWithdrawFinished(wd *models.Withdraw) {
	webhook := wd.Shop.Webhooks.OnWithdrawUpdated
	if webhook == nil {
		return
	}

	// prepare url
	u, err := url.Parse(wd.Shop.Host + *webhook)
	if err != nil {
		return
	}

	// prepare dto
	dto := withdraw.FromWithdraw(wd)
	data, err := json.Marshal(&dto)
	if err != nil {
		return
	}

	// отсылаем вебхук
	cl := http.Client{}
	reader := strings.NewReader(string(data))
	_, err = cl.Post(u.String(), "application/json", reader)
}
