package crud

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"payment-go/internal/repositories"
	"payment-go/internal/services"
	"payment-go/internal/transport/model/shop"
	"strconv"
	"sync"
)

type IShopCrudController interface {
	ICrudController
	Find(ctx *fiber.Ctx) error
	RegeneratePrivateKey(ctx *fiber.Ctx) error
	ValidateShopHost(ctx *fiber.Ctx) error
}
type shopCrudController struct {
}

var shopIns *shopCrudController
var shopOnce = sync.Once{}

func ShopCrudController() IShopCrudController {
	shopOnce.Do(func() {
		shopIns = &shopCrudController{}
	})
	return shopIns
}

func (crud *shopCrudController) GetActions() CrudActions {
	return AllCrudActions()
}

func (crud *shopCrudController) Create(ctx *fiber.Ctx) error {
	dto, err := ParseJSON(shop.CreateShopDto{}, ctx)
	if err != nil {
		return InvalidJSON(ctx)
	}

	sh, err := services.ShopService().CreateFromDto(dto)
	if err != nil {
		return ErrorJSON(ctx, err.Error())
	}

	return SuccessAdvancedJSON(ctx, fiber.Map{
		"shop": shop.FromShop(sh),
	})
}

func (crud *shopCrudController) Read(ctx *fiber.Ctx) error {
	strId := ctx.Query("id")
	id, err := strconv.ParseUint(strId, 10, 32)
	if err != nil || id == 0 {
		return ErrorJSON(ctx, "Invalid shop id passed")
	}

	sh, err := repositories.ShopRepository().FindById(uint(id))
	if err != nil {
		return ErrorJSON(ctx, "Shop not found")
	}

	var shopData *shop.ShopResponseDto
	if sh.HostValidated {
		shopData = shop.FromShop(sh)
	} else {
		code := services.ShopService().GetShopHostConfirmCode(sh)
		shopData = shop.FromShopWithConfirmCode(sh, code)
	}
	return SuccessAdvancedJSON(ctx, fiber.Map{
		"shop": shopData,
	})
}

func (crud *shopCrudController) Update(ctx *fiber.Ctx) error {
	dto, err := ParseJSON(shop.UpdateShopDto{}, ctx)
	if err != nil {
		return InvalidJSON(ctx)
	}

	if err = dto.Validate(); err != nil {
		return ErrorJSON(ctx, err.Error())
	}

	sh, err := services.ShopService().UpdateFromDto(dto)
	if sh == nil && err != nil {
		return ErrorJSON(ctx, "Shop not found")
	} else if err != nil {
		return ErrorJSON(ctx, "Error while saving")
	}

	return SuccessAdvancedJSON(ctx, fiber.Map{
		"shop": shop.FromShop(sh),
	})
}

func (crud *shopCrudController) Delete(ctx *fiber.Ctx) error {
	strId := ctx.Query("id")
	id, err := strconv.ParseUint(strId, 10, 32)
	if err != nil || id == 0 {
		return ErrorJSON(ctx, "Invalid shop id passed")
	}

	sh, err := repositories.ShopRepository().FindById(uint(id))
	if err != nil {
		return ErrorJSON(ctx, "Shop not found")
	}

	if sh.IsAvailable() {
		return ErrorJSON(ctx, "Deactivate shop before deleting.")
	}

	err = repositories.ShopRepository().Delete(uint(id))
	if err != nil {
		return ErrorJSON(ctx, "Error while deleting")
	}

	return SuccessJSON(ctx, fmt.Sprintf("Shop #%d was successfully deleted", id))
}

func (crud *shopCrudController) List(ctx *fiber.Ctx) error {
	p, err := NewPaginator(ctx)
	if err != nil {
		return ErrorJSON(ctx, err.Error())
	}
	page, size, order := p.GetArgs()

	strOwner := ctx.Query("owner_id")
	ownerId, err := strconv.ParseUint(strOwner, 10, 32)

	shops, err := repositories.ShopRepository().GetPaged(page, size, order, uint(ownerId))
	if err != nil {
		return ErrorJSON(ctx, "Unable to get shops.")
	}

	return SuccessAdvancedJSON(ctx, fiber.Map{
		"total": shops.Total,
		"shops": shop.FromShops(shops.Items),
	})
}

func (crud *shopCrudController) RegeneratePrivateKey(ctx *fiber.Ctx) error {
	strId := ctx.Query("shop_id")
	id, err := strconv.ParseUint(strId, 10, 32)
	if err != nil {
		return ErrorJSON(ctx, "shop_id is required")
	}

	sh, err := repositories.ShopRepository().FindById(uint(id))
	if err != nil {
		return ErrorJSON(ctx, "Shop is not found")
	}

	if err = services.ShopService().RegeneratePrivateKey(sh); err != nil {
		return ErrorJSON(ctx, "Error while changing private key")
	}

	return SuccessAdvancedJSON(ctx, fiber.Map{
		"message":     "Private key was changed!",
		"private_key": sh.Keys.PrivateKey,
	})
}

func (crud *shopCrudController) ValidateShopHost(ctx *fiber.Ctx) error {
	strId := ctx.Query("shop_id")
	id, err := strconv.ParseUint(strId, 10, 32)
	if err != nil {
		return ErrorJSON(ctx, "shop_id is required")
	}

	sh, err := repositories.ShopRepository().FindById(uint(id))
	if err != nil {
		return ErrorJSON(ctx, "Shop is not found")
	}

	if sh.HostValidated {
		return ErrorJSON(ctx, "Shop host is already validated")
	}

	validated, err := services.ShopService().ValidateHost(sh)
	if err != nil {
		return ErrorJSON(ctx, "Error while validating host.")
	}

	if !validated {
		return ErrorJSON(ctx, "Validation is unsuccessful.")
	}
	return SuccessJSON(ctx, "Validation completed successfully!")
}

func (crud *shopCrudController) Find(ctx *fiber.Ctx) error {
	p, err := NewPaginator(ctx)
	if err != nil {
		return ErrorJSON(ctx, err.Error())
	}
	page, size, _ := p.GetArgs()

	body := ctx.Body()
	dto, err := shop.BuildFindDto(body)
	if err != nil {
		return ErrorJSON(ctx, "Invalid request.")
	}

	shops, err := repositories.ShopRepository().Find(dto, page, size)
	if err != nil {
		return ErrorJSON(ctx, "Database error.")
	}

	return SuccessAdvancedJSON(ctx, fiber.Map{
		"total": shops.Total,
		"shops": shop.FromShops(shops.Items),
	})
}
