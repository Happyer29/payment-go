package crud

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"strings"
)

type ICrudController interface {
	Create(ctx *fiber.Ctx) error
	Read(ctx *fiber.Ctx) error
	Update(ctx *fiber.Ctx) error
	Delete(ctx *fiber.Ctx) error
	List(ctx *fiber.Ctx) error
	GetActions() CrudActions
}

type CrudActions struct {
	Create bool
	Read   bool
	Update bool
	Delete bool
	List   bool
}

func AllCrudActions() CrudActions {
	return CrudActions{
		Create: true,
		Read:   true,
		Update: true,
		Delete: true,
		List:   true,
	}
}

func ParseJSON[DTO interface{}](dto DTO, ctx *fiber.Ctx) (*DTO, error) {
	err := json.Unmarshal(ctx.Body(), &dto)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func InvalidJSON(ctx *fiber.Ctx) error {
	return ErrorJSON(ctx, "Got Invalid JSON")
}

func ErrorJSON(ctx *fiber.Ctx, message string) error {
	return ctx.JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}

func SuccessJSON(ctx *fiber.Ctx, message string) error {
	return ctx.JSON(fiber.Map{
		"success": true,
		"message": message,
	})
}

func SuccessAdvancedJSON(ctx *fiber.Ctx, data fiber.Map) error {
	jsonMap := fiber.Map{
		"success": true,
	}
	for k, v := range data {
		jsonMap[k] = v
	}
	return ctx.JSON(jsonMap)
}

type CrudPaginator struct {
	Page  uint
	Size  uint
	Order string
}

func NewPaginator(ctx *fiber.Ctx) (*CrudPaginator, error) {
	strPage := ctx.Query("page")
	page, err := strconv.ParseUint(strPage, 10, 32)
	if err != nil || page == 0 {
		return nil, fmt.Errorf("page parameter is invalid or not specified")
	}

	strSize := ctx.Query("size")
	size, err := strconv.ParseUint(strSize, 10, 32)
	if err != nil || size == 0 {
		return nil, fmt.Errorf("size parameter is invalid or not specified")
	}

	order := ctx.Query("order")
	if strings.ToUpper(order) == "DESC" {
		order = "DESC"
	} else {
		order = "ASC"
	}

	return &CrudPaginator{
		Page:  uint(page),
		Size:  uint(size),
		Order: order,
	}, nil
}

func (p *CrudPaginator) GetArgs() (uint, uint, string) {
	return p.Page - 1, p.Size, p.Order
}
