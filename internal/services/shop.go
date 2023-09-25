package services

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"payment-go/internal/models"
	"payment-go/internal/repositories"
	"payment-go/internal/transport/model/shop"
	"payment-go/internal/utils/shop/validation"
	"strconv"
	"sync"
)

type IShopService interface {
	AuthorizeDto(dto IAuthorizable) (*models.Shop, error)
	CreateFromDto(dto *shop.CreateShopDto) (*models.Shop, error)
	GetShopByPublicKey(publicKey string) (*models.Shop, error)
	GetShopHostValidator(sh *models.Shop) (validation.IShopValidator, error)
	GetShopHostConfirmCode(sh *models.Shop) string
	UpdateFromDto(dto *shop.UpdateShopDto) (*models.Shop, error)
	ValidateHost(sh *models.Shop) (bool, error)
	RegeneratePrivateKey(sh *models.Shop) error
}
type shopService struct {
}

var shopIns *shopService
var shopOnce = sync.Once{}

func ShopService() IShopService {
	shopOnce.Do(func() {
		shopIns = &shopService{}
	})
	return shopIns
}

func (s *shopService) CreateFromDto(dto *shop.CreateShopDto) (*models.Shop, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	entity, err := models.NewShop(dto.Name, dto.Host, dto.OwnerId)
	if err != nil {
		return nil, err
	}

	if err := repositories.ShopRepository().Save(entity); err != nil {
		return nil, fmt.Errorf("error while creating shop")
	}

	return entity, nil
}

func (s *shopService) UpdateFromDto(dto *shop.UpdateShopDto) (*models.Shop, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	sh, err := repositories.ShopRepository().FindById(dto.ID)
	if err != nil {
		return nil, err
	}

	sh.Name = dto.Name
	sh.Active = dto.Active
	sh.Moderated = dto.Moderated
	sh.Webhooks = dto.Webhooks.Resolve()

	err = repositories.ShopRepository().Save(sh)
	if err != nil {
		return sh, err
	}
	return sh, nil
}

func (s *shopService) ValidateHost(sh *models.Shop) (bool, error) {
	sv, err := s.GetShopHostValidator(sh)
	if err != nil {
		return false, err
	}
	ok := sv.Validate()
	sh.HostValidated = ok
	err = repositories.ShopRepository().Save(sh)
	return ok, err
}

func (s *shopService) GetShopHostValidator(sh *models.Shop) (validation.IShopValidator, error) {
	return validation.NewValidator(validation.ShopValidationOptions{
		ShopUrl:     sh.Host,
		ConfirmCode: s.GetShopHostConfirmCode(sh),
	})
}

func (s *shopService) GetShopHostConfirmCode(sh *models.Shop) string {
	data := sh.Keys.PublicKey.String() + strconv.Itoa(int(sh.ID)) + strconv.Itoa(int(sh.OwnerId))
	h := md5.New()
	h.Write([]byte(data))
	return "payment_confirm_" + hex.EncodeToString(h.Sum(nil))
}

func (s *shopService) RegeneratePrivateKey(sh *models.Shop) error {
	sh.RegeneratePrivateKey()
	return repositories.ShopRepository().Save(sh)
}

type IAuthorizable interface {
	GetCredentials() models.ShopKeys
}

func (s *shopService) AuthorizeDto(dto IAuthorizable) (*models.Shop, error) {
	sh, err := repositories.ShopRepository().FindByShopKeys(dto.GetCredentials())
	if err != nil {
		return nil, fmt.Errorf("authorizing by shop keys failed")
	}
	return sh, nil
}

func (s *shopService) GetShopByPublicKey(publicKey string) (*models.Shop, error) {
	return repositories.ShopRepository().FindByPublicKey(publicKey)
}
