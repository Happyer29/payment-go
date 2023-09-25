package crud

import (
	"github.com/gofiber/fiber/v2"
	"payment-go/internal/repositories"
	"payment-go/internal/services"
	"payment-go/internal/transport/model/bank_message"
	"strconv"
	"sync"
)

type IBankMessageCrudController interface {
	ICrudController
	Find(ctx *fiber.Ctx) error
	Approve(ctx *fiber.Ctx) error
	Decline(ctx *fiber.Ctx) error
}
type bankMessageCrudController struct {
}

var bmIns *bankMessageCrudController
var bmOnce = sync.Once{}

func BankMessageCrudController() IBankMessageCrudController {
	bmOnce.Do(func() {
		bmIns = &bankMessageCrudController{}
	})
	return bmIns
}

func (crud *bankMessageCrudController) GetActions() CrudActions {
	return CrudActions{
		Create: false,
		Read:   true,
		Update: false,
		Delete: false,
		List:   false,
	}
}

func (crud *bankMessageCrudController) Create(ctx *fiber.Ctx) error {
	return ctx.SendStatus(404)
}

func (crud *bankMessageCrudController) Read(ctx *fiber.Ctx) error {
	strId := ctx.Query("id")
	id, err := strconv.ParseUint(strId, 10, 32)
	if err != nil || id == 0 {
		return ErrorJSON(ctx, "Invalid bank_message id passed")
	}

	msg, err := repositories.BankMessageRepository().FindById(uint(id))
	if err != nil {
		return ErrorJSON(ctx, "Bank message not found")
	}

	return SuccessAdvancedJSON(ctx, fiber.Map{
		"bank_message": bank_message.FromBankMessage(msg),
	})
}

func (crud *bankMessageCrudController) Update(ctx *fiber.Ctx) error {
	return ctx.SendStatus(404)
}

func (crud *bankMessageCrudController) Delete(ctx *fiber.Ctx) error {
	return ctx.SendStatus(404)
}

func (crud *bankMessageCrudController) List(ctx *fiber.Ctx) error {
	return ctx.SendStatus(404)
}

func (crud *bankMessageCrudController) Find(ctx *fiber.Ctx) error {
	p, err := NewPaginator(ctx)
	if err != nil {
		return ErrorJSON(ctx, err.Error())
	}
	page, size, _ := p.GetArgs()

	body := ctx.Body()
	dto, err := bank_message.BuildFindDto(body)
	if err != nil {
		return ErrorJSON(ctx, "Invalid request.")
	}

	messages, err := repositories.BankMessageRepository().Find(dto, page, size)
	if err != nil {
		return ErrorJSON(ctx, "Database error.")
	}

	return SuccessAdvancedJSON(ctx, fiber.Map{
		"bank_messages": bank_message.FromBankMessages(messages.Items),
		"total":         messages.Total,
	})
}

func (crud *bankMessageCrudController) Approve(ctx *fiber.Ctx) error {
	dto, err := bank_message.ApprovalDtoFromJSON(ctx.Request().Body())
	if err != nil {
		return ctx.JSON(fiber.Map{
			"success": false,
			"error":   "invalid json",
		})
	}

	if err = services.BankMessageService().ProcessApproval(dto, true); err != nil {
		return ctx.JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func (crud *bankMessageCrudController) Decline(ctx *fiber.Ctx) error {
	dto, err := bank_message.ApprovalDtoFromJSON(ctx.Request().Body())
	if err != nil {
		return ctx.JSON(fiber.Map{
			"success": false,
			"error":   "invalid json",
		})
	}

	if err = services.BankMessageService().ProcessApproval(dto, false); err != nil {
		return ctx.JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
