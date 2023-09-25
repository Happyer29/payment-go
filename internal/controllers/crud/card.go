package crud

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"payment-go/internal/repositories"
	"payment-go/internal/services"
	"payment-go/internal/transport/model/card"
	"strconv"
	"strings"
	"sync"
)

type ICardCrudController interface {
	ICrudController
	CheckActivity(ctx *fiber.Ctx) error
	ChangeBalance(ctx *fiber.Ctx) error
}
type cardCrudController struct {
}

var cardIns *cardCrudController
var cardOnce = sync.Once{}

func CardCrudController() ICardCrudController {
	cardOnce.Do(func() {
		cardIns = &cardCrudController{}
	})
	return cardIns
}

func (crud *cardCrudController) GetActions() CrudActions {
	return AllCrudActions()
}

func (crud *cardCrudController) Create(ctx *fiber.Ctx) error {
	dto, err := ParseJSON(card.CreateCardDto{}, ctx)
	if err != nil {
		return InvalidJSON(ctx)
	}

	c, err := services.CardService().CreateFromDto(dto)
	if err != nil {
		return ErrorJSON(ctx, "Error while saving")
	}

	info, err := services.CardService().GetCardInfo(c.ID)
	if err != nil {
		info = nil
	}

	return SuccessAdvancedJSON(ctx, fiber.Map{
		"card": card.FromCard(c, info),
	})
}

func (crud *cardCrudController) Update(ctx *fiber.Ctx) error {
	dto, err := ParseJSON(card.UpdateCardDto{}, ctx)
	if err != nil {
		return InvalidJSON(ctx)
	}

	if err = dto.Validate(); err != nil {
		return ErrorJSON(ctx, err.Error())
	}

	crd, err := services.CardService().UpdateFromDto(dto)
	if crd == nil && err != nil {
		return ErrorJSON(ctx, "Card not found")
	} else if err != nil {
		return ErrorJSON(ctx, "Error while saving")
	}

	info, err := services.CardService().GetCardInfo(crd.ID)
	if err != nil {
		info = nil
	}

	return SuccessAdvancedJSON(ctx, fiber.Map{
		"card": card.FromCard(crd, info),
	})
}

func (crud *cardCrudController) Read(ctx *fiber.Ctx) error {
	strId := ctx.Query("id")
	id, err := strconv.ParseUint(strId, 10, 32)
	if err != nil || id == 0 {
		return ErrorJSON(ctx, "Invalid card id passed")
	}

	crd, err := repositories.CardRepository().FindById(uint(id))
	if err != nil {
		return ErrorJSON(ctx, "Card not found")
	}

	info, err := services.CardService().GetCardInfo(crd.ID)
	if err != nil {
		info = nil
	}

	return SuccessAdvancedJSON(ctx, fiber.Map{
		"card": card.FromCard(crd, info),
	})
}

func (crud *cardCrudController) Delete(ctx *fiber.Ctx) error {
	strId := ctx.Query("id")
	id, err := strconv.ParseUint(strId, 10, 32)
	if err != nil || id == 0 {
		return ErrorJSON(ctx, "Invalid card id passed")
	}

	crd, err := repositories.CardRepository().FindById(uint(id))
	if err != nil {
		return ErrorJSON(ctx, "Card not found")
	}

	if crd.IsActive() {
		return ErrorJSON(ctx, "Deactivate card before deleting.")
	}

	err = repositories.CardRepository().Delete(uint(id))
	if err != nil {
		return ErrorJSON(ctx, "Error while deleting")
	}

	return SuccessJSON(ctx, fmt.Sprintf("Card #%d was successfully deleted", id))
}

func (crud *cardCrudController) List(ctx *fiber.Ctx) error {
	p, err := NewPaginator(ctx)
	if err != nil {
		return ErrorJSON(ctx, err.Error())
	}

	cards, err := repositories.CardRepository().GetPaged(p.GetArgs())
	if err != nil {
		return ErrorJSON(ctx, "Unable to get cards.")
	}

	totalCount, err := repositories.CardRepository().CountAll()
	if err != nil {
		return ErrorJSON(ctx, "Database error.")
	}

	var cardsInfo []*card.CardResponseDto
	for _, crd := range cards {
		info, err := services.CardService().GetCardInfo(crd.ID)
		if err != nil {
			info = nil
		}
		cardsInfo = append(cardsInfo, card.FromCard(crd, info))
	}
	return SuccessAdvancedJSON(ctx, fiber.Map{
		"total": totalCount,
		"cards": cardsInfo,
	})
}

func (crud *cardCrudController) CheckActivity(ctx *fiber.Ctx) error {
	ids := strings.Split(ctx.Query("ids"), ",")

	res := fiber.Map{}
	for _, strId := range ids {
		id, err := strconv.ParseUint(strId, 10, 32)
		if err != nil {
			continue
		}

		crd, err := repositories.CardRepository().FindById(uint(id))
		if err != nil {
			continue
		}

		status, chance := services.CardService().GetCardUseStatus(crd)
		res[strconv.FormatUint(id, 10)] = fiber.Map{
			"status": status,
			"chance": chance,
		}
	}

	return ctx.JSON(res)
}

func (crud *cardCrudController) ChangeBalance(ctx *fiber.Ctx) error {
	var dto *card.ChangeBalanceDto
	if err := json.Unmarshal(ctx.Request().Body(), &dto); err != nil {
		return ctx.Status(502).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	req, err := dto.GetRequest()
	if err != nil {
		return ctx.Status(502).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	err = services.CardService().ChangeCardBalance(req)
	if err != nil {
		return ctx.Status(502).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
