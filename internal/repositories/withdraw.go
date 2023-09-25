package repositories

import (
	"fmt"
	"gorm.io/gorm"
	"payment-go/internal/database"
	"payment-go/internal/models"
	"payment-go/internal/repositories/include"
	"payment-go/internal/transport/model/withdraw"
	"strings"
	"sync"
)

type IWithdrawRepository interface {
	FindById(id uint) (*models.Withdraw, error)
	FindByNumber(number string) (*models.Withdraw, error)
	Find(dto *withdraw.FindWithdrawDto, page, size uint) (*include.PagedResultsList[models.Withdraw], error)
	Save(entity *models.Withdraw) error
}
type withdrawRepository struct {
	db *gorm.DB
}

var wdIns *withdrawRepository
var wdOnce = sync.Once{}

func WithdrawRepository() IWithdrawRepository {
	wdOnce.Do(func() {
		wdIns = &withdrawRepository{
			db: database.GetConnection(),
		}
	})
	return wdIns
}

func (repo *withdrawRepository) FindById(id uint) (*models.Withdraw, error) {
	var wd = &models.Withdraw{}
	err := repo.preload().First(wd, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return wd, nil
}

func (repo *withdrawRepository) FindByNumber(number string) (*models.Withdraw, error) {
	if len(number) == 0 {
		return nil, fmt.Errorf("withdraw number can not be empty")
	}

	var wd = &models.Withdraw{}
	err := repo.preload().First(wd, "number = ?", number).Error
	if err != nil {
		return nil, err
	}
	return wd, nil
}

func (repo *withdrawRepository) Find(dto *withdraw.FindWithdrawDto, page, size uint) (*include.PagedResultsList[models.Withdraw], error) {
	var res []*models.Withdraw
	query := repo.preload()

	if len(dto.ShopID) != 0 {
		query.Where("shop_id IN (?)", dto.ShopID)
	}

	if len(dto.OwnerID) != 0 {
		owners := repo.db.Model(&models.Shop{}).Select("id").Where("owner_id IN (?)", dto.OwnerID)
		query.Where("shop_id IN (?)", owners)
	}

	if len(dto.ID) != 0 {
		query.Where("id IN (?)", dto.ID)
	}

	if len(dto.Number) != 0 {
		query.Where("number = ?", dto.Number)
	}

	if len(dto.Type) != 0 {
		query.Where("type IN (?)", dto.Type)
	}

	if len(dto.CardNumber) != 0 {
		query.Where("card_number = ?", dto.CardNumber)
	}

	if len(dto.CardExpirationDate) != 0 {
		query.Where("card_expiration_date = ?", dto.CardExpirationDate)
	}

	if len(dto.Phone) != 0 {
		query.Where("phone = ?", dto.Phone)
	}

	if len(dto.Status) != 0 {
		query.Where("status IN (?)", dto.Status)
	}

	if dto.Amount != nil {
		query.Where("amount >= ? AND amount <= ?", dto.Amount.Min, dto.Amount.Max)
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
		query.Where("id LIKE ? OR number LIKE ? OR card_number LIKE ? OR phone LIKE ?", search, search, search, search)
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

	return &include.PagedResultsList[models.Withdraw]{
		Items: res,
		Total: uint(total),
	}, nil
}

func (repo *withdrawRepository) preload() *gorm.DB {
	res := repo.db.Model(&models.Withdraw{})
	res.Preload("Shop")
	return res
}

func (repo *withdrawRepository) Save(entity *models.Withdraw) error {
	return repo.db.Save(entity).Error
}
