package services

import (
	"fmt"
	"log"
	"payment-go/internal/config"
	"payment-go/internal/models"
	"payment-go/internal/providers"
	"payment-go/internal/repositories"
	"payment-go/internal/transport/bank"
	"payment-go/internal/utils/bank_api"
	"payment-go/internal/utils/proxy"
	"strconv"
	"strings"
	"sync"
	"time"
)

type IBankApiService interface {
	TaskCreateLink(ord *models.Order, cb func(link *models.PaymentLink))
	TaskCheckLink(link *models.PaymentLink, cb func(*models.PaymentLink))
}
type bankApiService struct {
}

var baIns IBankApiService
var baOnce = sync.Once{}

func BankApiService() IBankApiService {
	baOnce.Do(func() {
		baIns = &bankApiService{}
	})
	return baIns
}

func (s *bankApiService) TaskCreateLink(ord *models.Order, cb func(*models.PaymentLink)) {

	orderRepo := repositories.OrderRepository()
	linkRepo := repositories.PaymentLinkRepository()

	// создадим объект ссылки на оплату
	link := &models.PaymentLink{
		OrderID:  ord.ID,
		Order:    *ord,
		Amount:   ord.Amount,
		CardType: ord.GetCardType(),
		Status:   models.StatusNew,
	}

	// на случай ошибки
	taskFailed := func(err error) {
		err = PaymentLinkService().FinishLinkWithStatus(link, ord, models.StatusFailed)
		if err != nil {
			log.Printf("Error while finishing order #%d: %e", ord.ID, err)
		} else if cb != nil {
			cb(link)
		}
		SubscriptionService().RunSubscribers(link)
		SubscriptionService().RunSubscribers(ord)
	}

	if ord.Card.PhonePrefix == nil || ord.Card.PhoneNumber == nil {
		taskFailed(fmt.Errorf("got nil in card phone"))
	}
	phone := &config.Phone{
		Prefix: *ord.Card.PhonePrefix,
		Number: *ord.Card.PhoneNumber,
	}

	if err := linkRepo.Save(link); err != nil {
		taskFailed(err)
	}

	// запланируем таск на создание ссылки на оплату
	q := providers.TaskQueueProvider().GetQueue()
	q.AddTask(proxy.NewTask(func(pr proxy.IProxy) {
		// получение атрибутов
		b := bank_api.NewBank()
		atts, err := b.GetAttributes(pr)
		if err != nil {
			taskFailed(err)
			return
		}

		// получение checkMerchant
		requestDto := bank.NewCheckMerchantRequestDto(atts, phone)
		resp, err := b.CheckMerchant(pr, requestDto)
		if err != nil {
			taskFailed(err)
			return
		}

		// получение ссылки
		amount := fmt.Sprintf("%.2f", link.Amount)
		cardType, ok := config.GetConfig().Bank.CardTypeMapping[link.CardType]
		if !ok {
			taskFailed(fmt.Errorf("unknown card type"))
			return
		}
		linkDto := bank.NewLinkRequestDto(resp, amount, cardType)
		linkRes, err := b.GetLink(pr, linkDto)
		if err != nil {
			taskFailed(err)
			return
		}

		// меняем статус и сохраняем ссылку
		link.URL = linkRes.PaymentUrl
		link.Status = models.StatusPending
		if err = linkRepo.Save(link); err != nil {
			log.Println(err)
		}

		// меняем статус заказа
		ord.Status = models.StatusPending
		if err = orderRepo.Save(ord); err != nil {
			log.Println(err)
		}

		if cb != nil {
			cb(link)
		}
		go func() {
			//time.Sleep(3 * time.Second)
			SubscriptionService().RunSubscribers(link)
			SubscriptionService().RunSubscribers(ord)
		}()
	}))
}

func (s *bankApiService) TaskCheckLink(link *models.PaymentLink, cb func(*models.PaymentLink)) {

	ord := &link.Order

	paymentFailed := func(err error) {
		fmt.Println("paymentFailed()", err)

		err = PaymentLinkService().FinishLinkWithStatus(link, ord, models.StatusFailed)
		if err != nil {
			log.Printf("Error while finishing order #%d: %e", ord.ID, err)
		} else if cb != nil {
			cb(link)
		}
	}

	paymentCompleted := func() {
		fmt.Println("paymentCompleted")

		err := PaymentLinkService().FinishLinkWithStatus(link, ord, models.StatusCompleted)
		if err != nil {
			log.Println("Error while finishing order #"+strconv.Itoa(int(ord.ID)), err)
		} else if cb != nil {
			cb(link)
		}
	}

	q := providers.TaskQueueProvider().GetQueue()

	var taskGenerator func(attempt int) proxy.IHttpTask
	taskGenerator = func(attempt int) proxy.IHttpTask {
		return proxy.NewTask(func(pr proxy.IProxy) {
			if attempt > config.CheckLinkMaxAttempts {
				fmt.Println("reason attempts")
				paymentFailed(fmt.Errorf("attempt"))
				return
			}

			// если ссылка устарела
			if time.Since(link.UpdatedAt) > config.CheckLinkTimeout {
				fmt.Println("reason timeout")
				paymentFailed(fmt.Errorf("timeout"))
				return
			}

			doNextIteration := func() {
				time.Sleep(config.CheckLinkInterval)
				q.AddTask(taskGenerator(attempt + 1))
			}

			// переходим по ссылке
			res, err := pr.Get(link.URL)
			if err != nil {
				fmt.Println("pr.Get link.URL do next iteration")
				fmt.Println(err)
				go doNextIteration()
				return
			}
			// проверяем наличие редиректа
			redirect := res.Header.Get("Location")
			if len(redirect) < 5 {
				go doNextIteration()
				return
			}

			// разобьём редирект по /
			redirectParts := strings.Split(redirect, "/")
			l := len(redirectParts)
			if l < 5 {
				fmt.Println("redirectParts < 5 do next iteration", redirectParts)
				go doNextIteration()
				return
			}
			// достанем id транзакции
			transactionId := redirectParts[l-1]
			if len(transactionId) < 10 {
				fmt.Println("transaction id < 10", transactionId)
				go doNextIteration()
				return
			}

			// проверим транзакцию
			b := bank_api.NewBank()
			completed, err := b.CheckLink(pr, transactionId)
			if err != nil {
				fmt.Println("check link err")
				fmt.Println(err)
				go doNextIteration()
				return
			}
			if completed {
				fmt.Println("paymentCompleted")
				paymentCompleted()
			} else {
				fmt.Println("paymentFailed")
				paymentFailed(fmt.Errorf("not completed"))
			}
		})
	}

	go func() {
		time.Sleep(config.CheckLinkInterval)
		q.AddTask(taskGenerator(0))
	}()
}
