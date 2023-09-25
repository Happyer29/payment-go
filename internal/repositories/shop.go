package repositories

import (
	"gorm.io/gorm"
	"payment-go/internal/database"
	"payment-go/internal/models"
	"payment-go/internal/repositories/include"
	"payment-go/internal/transport/model/shop"
	"strings"
	"sync"
)

type IShopRepository interface {
	CountAll() (uint, error)
	CountAllUserShops(ownerId uint) (uint, error)
	Delete(id uint) error
	Find(dto *shop.FindShopDto, page, size uint) (*include.PagedResultsList[models.Shop], error)
	FindById(id uint) (*models.Shop, error)
	FindByShopKeys(keys models.ShopKeys) (*models.Shop, error)
	FindByPublicKey(publicKey string) (*models.Shop, error)
	GetUserShops(ownerId uint) ([]*models.Shop, error)
	GetPaged(page uint, size uint, order string, ownerId uint) (*include.PagedResultsList[models.Shop], error)
	Save(entity *models.Shop) error
}
type shopRepository struct {
	db *gorm.DB
}

var shopIns *shopRepository
var shopOnce = sync.Once{}

func ShopRepository() IShopRepository {
	shopOnce.Do(func() {
		shopIns = &shopRepository{db: database.GetConnection()}
	})
	return shopIns
}

func (repo *shopRepository) CountAll() (uint, error) {
	var result int64
	err := repo.db.Table("shops").Where("deleted_at IS NULL").Count(&result).Error
	if err != nil {
		return 0, err
	}
	return uint(result), nil
}

// CountAllUserShops Если ownerId=0, посчитает для всех пользователей
func (repo *shopRepository) CountAllUserShops(ownerId uint) (uint, error) {
	var result int64

	var query = repo.db.Table("shops").Where("deleted_at IS NULL")
	if ownerId != 0 {
		query.Where("owner_id = ?", ownerId)
	}

	if err := query.Count(&result).Error; err != nil {
		return 0, err
	}
	return uint(result), nil
}
func (repo *shopRepository) Delete(id uint) error {
	return repo.db.Delete(&models.Shop{}, id).Error
}
func (repo *shopRepository) FindById(id uint) (*models.Shop, error) {
	var sh = &models.Shop{}
	err := repo.db.First(sh, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return sh, nil
}

func (repo *shopRepository) FindByShopKeys(keys models.ShopKeys) (*models.Shop, error) {
	var sh = &models.Shop{}
	err := repo.db.First(
		sh,
		"public_key = ? AND private_key = ?",
		keys.PublicKey, keys.PrivateKey,
	).Error
	if err != nil {
		return nil, err
	}
	return sh, nil
}

func (repo *shopRepository) FindByPublicKey(publicKey string) (*models.Shop, error) {
	var sh = &models.Shop{}
	err := repo.db.First(
		sh, "public_key = ?", publicKey,
	).Error
	if err != nil {
		return nil, err
	}
	return sh, nil
}

func (repo *shopRepository) GetUserShops(ownerId uint) ([]*models.Shop, error) {
	var shops = make([]*models.Shop, 0)
	err := repo.db.Where("owner_id = ?", ownerId).Find(&shops).Error
	if err != nil {
		return nil, err
	}
	return shops, nil
}

// GetPaged Если ownerId=0, выдаст для всех пользователей
func (repo *shopRepository) GetPaged(page uint, size uint, order string, ownerId uint) (*include.PagedResultsList[models.Shop], error) {
	var res []*models.Shop
	var query = repo.preload().Limit(int(size)).Offset(int(page * size))
	if strings.ToUpper(order) == "DESC" {
		query.Order("id DESC")
	}
	if ownerId != 0 {
		query.Where("owner_id = ?", ownerId)
	}
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	query.Limit(int(size)).Offset(int(page * size))

	err = query.Find(&res).Error
	if err != nil {
		return nil, err
	}

	return &include.PagedResultsList[models.Shop]{
		Items: res,
		Total: uint(total),
	}, nil
}
func (repo *shopRepository) Save(entity *models.Shop) error {
	return repo.db.Save(entity).Error
}

func (repo *shopRepository) Find(dto *shop.FindShopDto, page, size uint) (*include.PagedResultsList[models.Shop], error) {
	var res []*models.Shop
	query := repo.preload()

	if len(dto.ID) != 0 {
		query.Where("id IN (?)", dto.ID)
	}

	if len(dto.OwnerID) != 0 {
		query.Where("owner_id IN (?)", dto.OwnerID)
	}

	if dto.Active != nil {
		query.Where("active = ?", *dto.Active == true)
	}

	if dto.Moderated != nil {
		query.Where("moderated = ?", *dto.Moderated == true)
	}

	if dto.HostValidated != nil {
		query.Where("host_validated = ?", *dto.HostValidated == true)
	}

	if dto.Sort != nil {
		var direction = "ASC"
		if strings.ToUpper(dto.Sort.Direction) != "ASC" {
			direction = "DESC"
		}
		query.Order(dto.Sort.Field + " " + direction)
	}

	if len(dto.Search) != 0 {
		search := "%" + dto.Search + "%"
		query.Where("id LIKE ? OR name LIKE ? OR host LIKE ? OR public_key LIKE ?", search, search, search, search)
	}

	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	query.Limit(int(size)).Offset(int(page * size))

	err = query.Find(&res).Error
	if err != nil {
		return nil, err
	}

	return &include.PagedResultsList[models.Shop]{
		Items: res,
		Total: uint(total),
	}, nil
}

func (repo *shopRepository) preload() *gorm.DB {
	return repo.db.Model(&models.Shop{})
}
