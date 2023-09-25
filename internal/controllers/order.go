package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"payment-go/internal/config"
	"payment-go/internal/models"
	"payment-go/internal/repositories"
	"payment-go/internal/services"
	"payment-go/internal/transport/model/order"
	"payment-go/internal/utils/card_manager"
	"sync"
	"time"
)

type IOrderController interface {
	Create(ctx *fiber.Ctx) error
	CreateLinkSSE(w http.ResponseWriter, r *http.Request)
	//CreateAndWaitForLink(ctx *fiber.Ctx) error
	CheckStatus(ctx *fiber.Ctx) error
	GetPaymentInfo(ctx *fiber.Ctx) error
	//StartChecking(ctx *fiber.Ctx) error
}
type orderController struct {
}

var orderIns IOrderController
var orderOnce = sync.Once{}

func OrderController() IOrderController {
	orderOnce.Do(func() {
		orderIns = &orderController{}
	})
	return orderIns
}

func (c *orderController) Create(ctx *fiber.Ctx) error {
	var dto *order.CreateOrderDto
	err := json.Unmarshal(ctx.Request().Body(), &dto)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"success": false,
			"error":   "Invalid json",
		})
	}

	// validate dto
	if err = dto.Validate(); err != nil {
		return ctx.JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	ord, err := services.OrderService().Create(dto)
	if err != nil {
		errText := err.Error()
		if err == card_manager.ErrNoCardsAvailable {
			errText = "There are no available cards. Try again later."
		}
		return ctx.JSON(fiber.Map{
			"success": false,
			"error":   errText,
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Order successfully created.",
		//"order_id":     ord.ID,
		"order_number": ord.Number.String(),
		"link":         fmt.Sprintf("%s/api/order/%s/process_to_payment", config.GetConfig().AppHost, ord.Number.String()),
		//"wait_for_link":   ord.HaveLink(),
		"card":            ord.GetCardIfPublic(),
		"order_timestamp": ord.CreatedAt.Unix(),
	})
}

func (c *orderController) GetPaymentInfo(ctx *fiber.Ctx) error {
	orderNumber := ctx.Query("order_number")
	if len(orderNumber) == 0 {
		return ctx.JSON(fiber.Map{
			"success": false,
			"error":   "No order_number specified",
		})
	}

	info, err := services.OrderService().GetPaymentInfo(orderNumber)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success":      true,
		"order_number": orderNumber,
		"info":         info.Map(),
	})
}

//func (c *orderController) StartChecking(ctx *fiber.Ctx) error {
//	orderNumber := ctx.Query("order_number")
//	ord, err := repositories.OrderRepository().FindByNumber(orderNumber)
//	if err != nil {
//		return ctx.SendString("Order not found")
//	}
//
//	link, err := repositories.PaymentLinkRepository().FindById(ord.ID)
//	if err != nil {
//		return ctx.SendString("get link err")
//	}
//
//	services.BankApiService().TaskCheckLink(link, func(link *models.PaymentLink) {})
//	return ctx.SendString("successfully added task")
//}

//func (c *orderController) CreateAndWaitForLink(ctx *fiber.Ctx) error {
//	var dto *order.CreateOrderDto
//	err := json.Unmarshal(ctx.Request().Body(), &dto)
//	if err != nil {
//		return ctx.SendStatus(502)
//	}
//	if err = dto.Validate(); err != nil {
//		return ctx.Status(502).JSON(fiber.Map{
//			"success": false,
//			"error":   err.Error(),
//		})
//	}
//
//	// authorize dto
//	_, err = services.ShopService().AuthorizeDto(dto)
//	if err != nil {
//		return ctx.Status(403).JSON(fiber.Map{
//			"success": false,
//			"error":   "Invalid credentials",
//		})
//	}
//
//	ord, err := services.OrderService().Create(dto)
//	if err != nil {
//		return ctx.Status(502).JSON(fiber.Map{
//			"success": false,
//			"error":   "Error while saving",
//		})
//	}
//
//	// дождёмся генерации ссылки
//	var link *models.PaymentLink
//	var linkChan = make(chan *models.PaymentLink)
//	services.BankApiService().TaskCreateLink(ord, func(link *models.PaymentLink) {
//		linkChan <- link
//	})
//
//	select {
//	case link = <-linkChan:
//	case <-time.After(config.CreateLinkTimeout):
//		return ctx.JSON(fiber.Map{
//			"success": false,
//			"error":   "Timeout while creating a payment link.",
//		})
//	}
//
//	// если ссылка создалась, запускаем таск на проверку ссылки
//	services.BankApiService().TaskCheckLink(link, func(link *models.PaymentLink) {
//		// если указан вебхук, отправляем инфу о платеже
//		services.WebhookService().SendOrderCompleted(&link.Order, &dto.PaymentCallbackUrl)
//	})
//
//	if link.Status == models.StatusFailed {
//		return ctx.JSON(fiber.Map{
//			"success": false,
//			"error":   "Error while creating link.",
//			//"order_id": ord.ID,
//			"order_number": ord.Number.String(),
//		})
//	}
//
//	return ctx.JSON(fiber.Map{
//		"success": true,
//		"message": "Order successfully created.",
//		//"order_id":     ord.ID,
//		"order_number": ord.Number.String(),
//		"payment_link": link.URL,
//	})
//}

type SSEOrderClient struct {
	name   string
	events chan string
}

func (c *orderController) CreateLinkSSE(w http.ResponseWriter, r *http.Request) {

	orderNumber := r.URL.Query().Get("order_number")
	ord, err := repositories.OrderRepository().FindByNumber(orderNumber)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	orderId := ord.ID

	//w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// если ссылка готова
	link, err := repositories.PaymentLinkRepository().FindByOrderId(orderId)
	if err == nil && link.Status == models.StatusPending {
		fmt.Fprintf(w, "data: {\"link\":\"%v\"}\n\n", link.URL)
		w.WriteHeader(200)
		return
	}

	f, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(502)
		return
	}

	linkChan := make(chan string, 1)
	defer func() {
		close(linkChan)
		linkChan = nil
	}()

	cb := func(entity any) {
		//fmt.Printf("callback %#v\n", entity)
		lnk, ok := entity.(*models.PaymentLink)
		if ok {
			linkChan <- lnk.URL
		}
	}

	err = services.SubscriptionService().Subscribe(cb, services.FByID(link.ID, link))
	if err != nil {
		fmt.Println("--unable to subscribe", orderId, err)
		w.WriteHeader(502)
		return
	}

	select {
	case lnk := <-linkChan:
		if lnk == "" {
			f.Flush()
			return
		}

		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.Encode(map[string]any{
			"link": lnk,
		})
		fmt.Fprintf(w, "data: %v\n\n", buf.String())
	case <-time.After(config.CreateLinkTimeout):
		w.WriteHeader(504)
		return
		//fmt.Fprintf(w, ": nothing to sent\n\n")
	case <-r.Context().Done():
		return
	}
	f.Flush()
}

func (c *orderController) CheckStatus(ctx *fiber.Ctx) error {
	orderNumber := ctx.Params("order_number")

	status, err := services.OrderService().GetOrderStatus(orderNumber)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Order with number '%s' not found.", orderNumber),
		})
	}

	return ctx.JSON(fiber.Map{
		"success":      true,
		"order_status": status,
	})
}
