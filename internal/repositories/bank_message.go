package repositories

import (
	"gorm.io/gorm"
	"payment-go/internal/database"
	"payment-go/internal/models"
	"payment-go/internal/repositories/include"
	"payment-go/internal/transport/model/bank_message"
	"strings"
	"sync"
)

type IBankMessageRepository interface {
	Find(dto *bank_message.FindBankMessageDto, page, size uint) (*include.PagedResultsList[models.BankMessage], error)
	FindById(id uint) (*models.BankMessage, error)
	Save(msg *models.BankMessage) error
}
type bankMessageRepository struct {
	db *gorm.DB
}

var bmIns *bankMessageRepository
var bmOnce = sync.Once{}

func BankMessageRepository() IBankMessageRepository {
	bmOnce.Do(func() {
		bmIns = &bankMessageRepository{
			db: database.GetConnection(),
		}
	})
	return bmIns
}

func (repo *bankMessageRepository) Find(dto *bank_message.FindBankMessageDto, page, size uint) (*include.PagedResultsList[models.BankMessage], error) {
	var res []*models.BankMessage
	query := repo.preload()

	if len(dto.ID) != 0 {
		query.Where("id IN (?)", dto.ID)
	}
	if len(dto.OrderID) != 0 {
		query.Where("order_id IN (?)", dto.OrderID)
	}
	if len(dto.ShopID) != 0 {
		query.Where("shop_id IN (?)", dto.ShopID)
	}

	if len(dto.CardNumber) != 0 {
		query.Where("card_number IN (?)", dto.CardNumber)
	}

	if dto.Amount != nil {
		query.Where("amount >= ? AND amount <= ?", dto.Amount.Min, dto.Amount.Max)
	}

	if len(dto.Status) != 0 {
		query.Where("status IN (?)", dto.Status)
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
		query.Where(
			"id LIKE ? OR sender_phone LIKE ? OR receiver_phone LIKE ? OR raw_message LIKE ? OR card_number LIKE ? OR error LIKE ?",
			search, search, search, search, search, search,
		)
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

	return &include.PagedResultsList[models.BankMessage]{
		Items: res,
		Total: uint(total),
	}, nil
}

func (repo *bankMessageRepository) FindById(id uint) (*models.BankMessage, error) {
	var msg = &models.BankMessage{}
	err := repo.db.First(msg, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (repo *bankMessageRepository) Save(msg *models.BankMessage) error {
	return repo.db.Save(msg).Error
}

func (repo *bankMessageRepository) preload() *gorm.DB {
	return repo.db.Model(&models.BankMessage{})
}
