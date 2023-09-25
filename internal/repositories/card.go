package repositories

import (
	"fmt"
	"gorm.io/gorm"
	"payment-go/internal/database"
	"payment-go/internal/models"
	"strings"
	"sync"
)

type ICardRepository interface {
	AssertInfoExists(cardId uint) error
	CountAll() (uint, error)
	Delete(id uint) error
	FindById(id uint) (*models.Card, error)
	FindByCardNumber(cardNumber string) (*models.Card, error)
	GetAllActiveCards() ([]*models.Card, error)
	GetPaged(page uint, size uint, order string) ([]*models.Card, error)
	GetCardInfo(cardId uint) (*models.CardInfo, error)
	IncreaseBalance(cardId uint, amount float64) error
	DecreaseBalance(cardId uint, amount float64) error
	Save(entity *models.Card) error
}
type cardRepository struct {
	db *gorm.DB
}

var cardIns ICardRepository
var cardOnce = sync.Once{}

func CardRepository() ICardRepository {
	cardOnce.Do(func() {
		cardIns = &cardRepository{
			db: database.GetConnection(),
		}
	})
	return cardIns
}

func (repo *cardRepository) Save(entity *models.Card) error {
	err := repo.db.Save(entity).Error
	if err != nil {
		return err
	}

	return repo.AssertInfoExists(entity.ID)
}

func (repo *cardRepository) Delete(id uint) error {
	return repo.db.Delete(&models.Card{}, id).Error
}

func (repo *cardRepository) FindById(id uint) (*models.Card, error) {
	var card = &models.Card{}
	err := repo.db.First(card, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (repo *cardRepository) FindByCardNumber(cardNumber string) (*models.Card, error) {
	var card = &models.Card{}
	err := repo.db.First(card, "card_number = ?", cardNumber).Error
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (repo *cardRepository) GetAllActiveCards() ([]*models.Card, error) {
	var cards = make([]*models.Card, 0)
	err := repo.db.Where("status = ?", models.CardStatusEnabled).Find(&cards).Error
	if err != nil {
		return nil, err
	}
	return cards, nil
}

func (repo *cardRepository) CountAll() (uint, error) {
	var result int64
	err := repo.db.Table("cards").Where("deleted_at IS NULL").Count(&result).Error
	if err != nil {
		return 0, err
	}
	return uint(result), nil
}

func (repo *cardRepository) GetPaged(page uint, size uint, order string) ([]*models.Card, error) {
	var res []*models.Card
	query := repo.db.Limit(int(size)).Offset(int(page * size))
	if strings.ToUpper(order) == "DESC" {
		query.Order("id DESC")
	}
	err := query.Find(&res).Error
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (repo *cardRepository) GetCardInfo(cardId uint) (*models.CardInfo, error) {
	var res = &models.CardInfo{}
	err := repo.db.First(res, "card_id = ?", cardId).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (repo *cardRepository) addInfoForCard(cardId uint) error {
	var entity = &models.CardInfo{
		CardID:  cardId,
		Balance: 0,
	}
	return repo.db.Create(entity).Error
}

func (repo *cardRepository) AssertInfoExists(cardId uint) error {
	// создать card info если его нету
	_, err := repo.GetCardInfo(cardId)
	if err != nil {
		err = repo.addInfoForCard(cardId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repo *cardRepository) IncreaseBalance(cardId uint, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	err := repo.AssertInfoExists(cardId)
	if err != nil {
		return err
	}

	query := repo.db.Model(&models.CardInfo{})
	query.Where("card_id = ?", cardId)
	return query.Update("balance", gorm.Expr("balance + ?", amount)).Error
}

func (repo *cardRepository) DecreaseBalance(cardId uint, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	err := repo.AssertInfoExists(cardId)
	if err != nil {
		return err
	}

	query := repo.db.Model(&models.CardInfo{})
	query.Where("card_id = ?", cardId)
	return query.Update("balance", gorm.Expr("balance - ?", amount)).Error
}
