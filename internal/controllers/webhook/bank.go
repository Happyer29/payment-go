package webhook

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"payment-go/internal/services"
	"payment-go/internal/transport/bank/webhook"
	"sync"
)

type IBankWebhookController interface {
	PaymentInfo(ctx *fiber.Ctx) error
}
type bankWebhookController struct {
}

var bankIns *bankWebhookController
var bankOnce = sync.Once{}

func BankWebhookController() IBankWebhookController {
	bankOnce.Do(func() {
		bankIns = &bankWebhookController{}
	})
	return bankIns
}

func (con *bankWebhookController) PaymentInfo(ctx *fiber.Ctx) error {
	fmt.Println("PaymentInfo")
	fmt.Println(string(ctx.Request().Body()))

	request, err := webhook.FromRequestBody(ctx.Request().Body())
	if err != nil {
		request, err = webhook.FromPostFields(ctx.FormValue)
		if err != nil {
			return ctx.Status(502).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}
		//return ctx.Status(502).JSON(fiber.Map{
		//	"success": false,
		//	"error":   err.Error(),
		//})
	}

	dto, err := request.ToDto()
	if err != nil {
		return ctx.Status(502).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	if err := services.CardService().ReceivePayment(dto); err != nil {
		return ctx.Status(502).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return ctx.Status(200).JSON(fiber.Map{
		"success": true,
		"message": "Payment is successfully processed",
	})
}
