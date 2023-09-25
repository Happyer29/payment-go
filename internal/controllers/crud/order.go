package crud

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"payment-go/internal/repositories"
	"payment-go/internal/services"
	"payment-go/internal/transport/model/order"
	"strconv"
	"sync"
)

type IOrderCrudController interface {
	ICrudController
	Find(ctx *fiber.Ctx) error
}
type orderCrudController struct {
}

var orderIns *orderCrudController
var orderOnce = sync.Once{}

func OrderCrudController() IOrderCrudController {
	orderOnce.Do(func() {
		orderIns = &orderCrudController{}
	})
	return orderIns
}

func (crud *orderCrudController) Create(ctx *fiber.Ctx) error {
	dto, err := ParseJSON(order.CreateOrderDto{}, ctx)
	if err != nil {
		return InvalidJSON(ctx)
	}

	ord, err := services.OrderService().Create(dto)
	if err != nil {
		return ErrorJSON(ctx, "Error while saving")
	}

	return SuccessAdvancedJSON(ctx, fiber.Map{
		"order": order.FromOrder(ord),
	})
}

func (crud *orderCrudController) Update(ctx *fiber.Ctx) error {
	dto, err := ParseJSON(order.UpdateOrderDto{}, ctx)
	if err != nil {
		return InvalidJSON(ctx)
	}

	if err = dto.Validate(); err != nil {
		return ErrorJSON(ctx, err.Error())
	}

	ord, err := services.OrderService().UpdateFromDto(dto)
	if ord == nil && err != nil {
		return ErrorJSON(ctx, "Order not found")
	} else if err != nil {
		return ErrorJSON(ctx, "Error while saving")
	}

	return SuccessAdvancedJSON(ctx, fiber.Map{
		"order": order.FromOrder(ord),
	})
}

func (crud *orderCrudController) Read(ctx *fiber.Ctx) error {
	strId := ctx.Query("id")
	id, err := strconv.ParseUint(strId, 10, 32)
	if err != nil || id == 0 {
		return ErrorJSON(ctx, "Invalid order id passed")
	}

	ord, err := repositories.OrderRepository().FindById(uint(id))
	if err != nil {
		return ErrorJSON(ctx, "Order not found")
	}

	return SuccessAdvancedJSON(ctx, fiber.Map{
		"order": order.FromOrder(ord),
	})
}

func (crud *orderCrudController) Delete(ctx *fiber.Ctx) error {
	strId := ctx.Query("id")
	id, err := strconv.ParseUint(strId, 10, 32)
	if err != nil || id == 0 {
		return ErrorJSON(ctx, "Invalid order id passed")
	}

	err = repositories.OrderRepository().Delete(uint(id))
	if err != nil {
		return ErrorJSON(ctx, "Error while deleting")
	}

	return SuccessJSON(ctx, fmt.Sprintf("Order #%d was successfully deleted", id))
}

func (crud *orderCrudController) List(ctx *fiber.Ctx) error {
	p, err := NewPaginator(ctx)
	if err != nil {
		return ErrorJSON(ctx, err.Error())
	}
	page, size, sort := p.GetArgs()

	strShop := ctx.Query("shop_id")
	shopId, err := strconv.ParseUint(strShop, 10, 32)

	strOwner := ctx.Query("owner_id")
	ownerId, err := strconv.ParseUint(strOwner, 10, 32)

	data, err := repositories.OrderRepository().GetPaged(page, size, sort, uint(shopId), uint(ownerId))
	if err != nil {
		return ErrorJSON(ctx, "Unable to get orders.")
	}

	return SuccessAdvancedJSON(ctx, fiber.Map{
		"total":  data.Total,
		"orders": order.FromOrders(data.Items),
	})
}

func (crud *orderCrudController) Find(ctx *fiber.Ctx) error {
	p, err := NewPaginator(ctx)
	if err != nil {
		return ErrorJSON(ctx, err.Error())
	}
	page, size, _ := p.GetArgs()

	body := ctx.Body()
	dto, err := order.BuildFindDto(body)
	if err != nil {
		return ErrorJSON(ctx, "Invalid request.")
	}

	orders, err := repositories.OrderRepository().Find(dto, page, size)
	if err != nil {
		return ErrorJSON(ctx, "Database error.")
	}

	return SuccessAdvancedJSON(ctx, fiber.Map{
		"total":  orders.Total,
		"orders": order.FromOrders(orders.Items),
	})
}

func (crud *orderCrudController) GetActions() CrudActions {
	return AllCrudActions()
}
