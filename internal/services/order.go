package services

import (
	"fmt"
	"log"
	"payment-go/internal/models"
	"payment-go/internal/repositories"
	"payment-go/internal/transport/analytics/card"
	card2 "payment-go/internal/transport/model/card"
	"payment-go/internal/transport/model/order"
	"sync"
)

type IOrderService interface {
	Create(dto *order.CreateOrderDto) (*models.Order, error)
	FinishOrderWithStatus(ord *models.Order, status string, moment uint) error
	GetOrderStatus(orderNumber string) (string, error)
	GetPaymentInfo(orderNumber string) (order.IOrderPaymentInfoDto, error)
	GetTotals(dto *card.GetTotalsDto) ([]*repositories.TotalsResultDto, error)
	UpdateFromDto(dto *order.UpdateOrderDto) (*models.Order, error)
}
type orderService struct {
}

var orderIns IOrderService
var orderOnce = sync.Once{}

func OrderService() IOrderService {
	orderOnce.Do(func() {
		orderIns = &orderService{}
	})
	return orderIns
}

func (s *orderService) Create(dto *order.CreateOrderDto) (*models.Order, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	if err := PaymentService().ValidateOrderBeforeCreate(dto); err != nil {
		return nil, err
	}

	// authorize dto
	sh, err := ShopService().AuthorizeDto(dto)
	if err != nil {
		return nil, err
	}

	if !sh.IsAvailable() {
		return nil, fmt.Errorf("shop is inactive")
	}

	ord := &models.Order{
		Payload:       dto.Payload,
		Amount:        dto.Amount,
		PaymentMethod: dto.PaymentMethod,
		//Card:          *crd,
		Shop:     *sh,
		DatePaid: nil,
		Status:   models.StatusNew,
	}

	// подберём нужную карту
	crd, err := CardService().ChooseCard(ord)
	if err != nil {
		return nil, err
	}
	ord.Card = *crd.GetCard()

	if err := repositories.OrderRepository().Save(ord); err != nil {
		// снимем блокировку с карты
		crd.Unlock()
		return nil, fmt.Errorf("error while saving")
	}
	// запишем id заказа в карту
	crd.SetOrderId(ord.ID)

	// запустим события
	go EventService().OrderCreated(ord, dto)

	return ord, nil
}

func (s *orderService) UpdateFromDto(dto *order.UpdateOrderDto) (*models.Order, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	ord, err := repositories.OrderRepository().FindById(dto.ID)
	if err != nil {
		return nil, err
	}

	newOrd := *ord
	newOrd.Status = dto.Status
	err = repositories.OrderRepository().Save(&newOrd)
	if err != nil {
		return ord, err
	}

	// запустим события
	EventService().OrderUpdated(ord, &newOrd)

	return ord, nil
}

func (s *orderService) GetOrderStatus(orderNumber string) (string, error) {
	ord, err := repositories.OrderRepository().FindByNumber(orderNumber)
	if err != nil {
		return "", err
	}
	return ord.Status, nil
}

func (s *orderService) GetTotals(dto *card.GetTotalsDto) ([]*repositories.TotalsResultDto, error) {
	if dto == nil {
		return nil, fmt.Errorf("got invalid dto")
	}
	if err := dto.Validate(); err != nil {
		return nil, err
	}
	return repositories.OrderRepository().GetTotals(dto)
}

func (s *orderService) FinishOrderWithStatus(ord *models.Order, status string, moment uint) error {

	if !models.IsStatusValid(status) {
		return fmt.Errorf("invalid status on finishing the order #%d", ord.ID)
	}

	ord.Status = status
	ord.DatePaid = &moment
	if err := repositories.OrderRepository().Save(ord); err != nil {
		return err
	}

	if status == models.StatusCompleted {
		crd := ord.Card
		req, err := card2.NewChangeBalanceRequest(crd.ID, ord.Amount, true)
		if err == nil {
			err = CardService().ChangeCardBalance(req)
			if err != nil {
				log.Printf("error while changing card balance after completed order")
			}
		} else {
			log.Println(err)
		}
	}

	defer EventService().OrderCompleted(ord)
	return nil
}

func (s *orderService) GetPaymentInfo(orderNumber string) (order.IOrderPaymentInfoDto, error) {
	ord, err := repositories.OrderRepository().FindByNumber(orderNumber)
	if err != nil {
		return nil, fmt.Errorf("order not found")
	}

	return order.NewPaymentInfo(ord)
}
