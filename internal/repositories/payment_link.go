package repositories

import (
	"gorm.io/gorm"
	"payment-go/internal/config"
	"payment-go/internal/database"
	"payment-go/internal/models"
	"payment-go/internal/repositories/include"
	"payment-go/internal/transport/model/payment_link"
	"strings"
	"sync"
	"time"
)

type IPaymentLinkRepository interface {
	CountAll() (uint, error)
	Find(dto *payment_link.FindPaymentLinkDto, page, size uint) (*include.PagedResultsList[models.PaymentLink], error)
	FindById(id uint) (*models.PaymentLink, error)
	FindByOrderId(orderId uint) (*models.PaymentLink, error)
	GetPending() ([]*models.PaymentLink, error)
	GetPaged(page uint, size uint, order string, shopId uint, ownerId uint) (*PagedPaymentLinks, error)
	Save(entity *models.PaymentLink) error
}
type paymentLinkRepository struct {
	db *gorm.DB
}

var plIns IPaymentLinkRepository
var plOnce = sync.Once{}

func PaymentLinkRepository() IPaymentLinkRepository {
	plOnce.Do(func() {
		plIns = &paymentLinkRepository{
			db: database.GetConnection(),
		}
	})
	return plIns
}

func (repo *paymentLinkRepository) Save(entity *models.PaymentLink) error {
	return repo.db.Save(entity).Error
}

func (repo *paymentLinkRepository) FindById(id uint) (*models.PaymentLink, error) {
	var link = &models.PaymentLink{}
	err := repo.preload().First(link, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return link, nil
}

func (repo *paymentLinkRepository) FindByOrderId(orderId uint) (*models.PaymentLink, error) {
	var link = &models.PaymentLink{}
	err := repo.preload().First(link, "order_id = ?", orderId).Error
	if err != nil {
		return nil, err
	}
	return link, nil
}

func (repo *paymentLinkRepository) CountAll() (uint, error) {
	var result int64
	err := repo.db.Table("payment_links").Where("deleted_at IS NULL").Count(&result).Error
	if err != nil {
		return 0, err
	}
	return uint(result), nil
}

type PagedPaymentLinks struct {
	Links []*models.PaymentLink
	Total uint
}

func (repo *paymentLinkRepository) GetPaged(page uint, size uint, order string, shopId uint, ownerId uint) (*PagedPaymentLinks, error) {
	var res []*models.PaymentLink
	var query = repo.preload()

	if shopId != 0 { // by shop orders
		query.Where(
			"order_id IN (?)",
			repo.db.Model(&models.Order{}).Select("id").Where("shop_id = ?", shopId),
		)
	} else if ownerId != 0 { // by owner shop orders
		var subQuery = repo.db.Model(&models.Shop{}).Select("id").Where("owner_id = ?", ownerId)
		query.Where(
			"order_id IN (?)",
			repo.db.Model(&models.Order{}).Select("id").Where("shop_id IN (?)", subQuery),
		)
	}

	// count all
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	// ordering/sorting
	if strings.ToUpper(order) == "DESC" {
		query.Order("id DESC")
	}

	// paging
	query.Limit(int(size)).Offset(int(page * size))

	if err := query.Find(&res).Error; err != nil {
		return nil, err
	}

	return &PagedPaymentLinks{
		Links: res,
		Total: uint(total),
	}, nil
}

func (repo *paymentLinkRepository) GetPending() ([]*models.PaymentLink, error) {
	timeActive := time.Now().Add(-config.CheckLinkTimeout)

	query := repo.preload().Where("updated_at >= ? and status = ?", timeActive, models.StatusPending)

	var res []*models.PaymentLink
	err := query.Find(&res).Error
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (repo *paymentLinkRepository) Find(dto *payment_link.FindPaymentLinkDto, page, size uint) (*include.PagedResultsList[models.PaymentLink], error) {
	var res []*models.PaymentLink
	query := repo.preload()

	if len(dto.OwnerID) != 0 {
		shops := repo.db.Model(&models.Shop{}).Select("id").Where("owner_id IN (?)", dto.OwnerID)
		orders := repo.db.Model(&models.Order{}).Select("id").Where("shop_id IN (?)", shops)
		query.Where("order_id IN (?)", orders)
	}

	if len(dto.ShopID) != 0 {
		orders := repo.db.Model(&models.Order{}).Select("id").Where("shop_id IN (?)", dto.ShopID)
		query.Where("order_id IN (?)", orders)
	}

	if len(dto.CardID) != 0 {
		orders := repo.db.Model(&models.Order{}).Select("id").Where("card_id IN (?)", dto.CardID)
		query.Where("order_id IN (?)", orders)
	}

	if len(dto.OrderID) != 0 {
		query.Where("order_id IN (?)", dto.OrderID)
	}

	if len(dto.ID) != 0 {
		query.Where("id IN (?)", dto.ID)
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
		query.Where("id LIKE ? OR transaction_id LIKE ?", search, search)
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

	return &include.PagedResultsList[models.PaymentLink]{
		Items: res,
		Total: uint(total),
	}, nil
}

func (repo *paymentLinkRepository) preload() *gorm.DB {
	res := repo.db.Model(&models.PaymentLink{})
	res.Preload("Order")
	res.Preload("Order.Shop")
	res.Preload("Order.Card")
	return res
}
