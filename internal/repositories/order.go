package repositories

import (
	"fmt"
	"gorm.io/gorm"
	"payment-go/internal/database"
	"payment-go/internal/models"
	"payment-go/internal/repositories/include"
	"payment-go/internal/transport/analytics/card"
	"payment-go/internal/transport/model/order"
	"strings"
	"sync"
)

type IOrderRepository interface {
	CountAll() (uint, error)
	Delete(id uint) error
	Find(dto *order.ReadOrderDto, page, size uint) (*include.PagedResultsList[models.Order], error)
	FindById(id uint) (*models.Order, error)
	FindByNumber(number string) (*models.Order, error)
	GetPaged(page uint, size uint, order string, shopId uint, ownerId uint) (*include.PagedResultsList[models.Order], error)
	GetTotals(dto *card.GetTotalsDto) ([]*TotalsResultDto, error)
	Save(entity *models.Order) error
}
type orderRepository struct {
	db *gorm.DB
}

var orderIns IOrderRepository
var orderOnce = sync.Once{}

func OrderRepository() IOrderRepository {
	orderOnce.Do(func() {
		orderIns = &orderRepository{
			db: database.GetConnection(),
		}
	})
	return orderIns
}

func (repo *orderRepository) Save(entity *models.Order) error {
	return repo.db.Save(entity).Error
}

func (repo *orderRepository) Delete(id uint) error {
	return repo.db.Delete(&models.Order{}, id).Error
}

func (repo *orderRepository) CountAll() (uint, error) {
	var result int64
	err := repo.db.Table("orders").Where("deleted_at IS NULL").Count(&result).Error
	if err != nil {
		return 0, err
	}
	return uint(result), nil
}

type PagedOrders struct {
	Orders []*models.Order
	Total  uint
}

func (repo *orderRepository) GetPaged(page uint, size uint, order string, shopId uint, ownerId uint) (*include.PagedResultsList[models.Order], error) {
	var res []*models.Order
	query := repo.preload()

	if shopId != 0 {
		query.Where("shop_id = ?", shopId)
	} else if ownerId != 0 {
		query.Where("shop_id IN (?)", repo.db.Model(&models.Shop{}).Select("id").Where("owner_id = ?", ownerId))
	}

	if strings.ToUpper(order) == "DESC" {
		query.Order("id DESC")
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

	return &include.PagedResultsList[models.Order]{
		Items: res,
		Total: uint(total),
	}, nil
}

func (repo *orderRepository) FindById(id uint) (*models.Order, error) {
	var ord = &models.Order{}
	err := repo.preload().First(ord, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return ord, nil
}

func (repo *orderRepository) FindByNumber(number string) (*models.Order, error) {
	if len(number) == 0 {
		return nil, fmt.Errorf("order number can not be empty")
	}

	var ord = &models.Order{}
	err := repo.preload().First(ord, "number = ?", number).Error
	if err != nil {
		return nil, err
	}
	return ord, nil
}

func (repo *orderRepository) preload() *gorm.DB {
	res := repo.db.Model(&models.Order{})
	res.Preload("Shop")
	res.Preload("Card")
	return res
}

func (repo *orderRepository) Find(dto *order.ReadOrderDto, page, size uint) (*include.PagedResultsList[models.Order], error) {
	var res []*models.Order
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

	if len(dto.CardID) != 0 {
		query.Where("card_id IN (?)", dto.CardID)
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
		query.Where("id LIKE ? OR number LIKE ?", search, search)
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

	return &include.PagedResultsList[models.Order]{
		Items: res,
		Total: uint(total),
	}, nil
}

type TotalsResultDto struct {
	DateGroup uint    `json:"date_group"`
	Card      uint    `json:"card"`
	Status    string  `json:"status"`
	Total     float64 `json:"total"`
}

func (repo *orderRepository) GetTotals(dto *card.GetTotalsDto) ([]*TotalsResultDto, error) {
	query := repo.preload().Where("date_paid IS NOT NULL") // получаем только завершённые заказы

	if dto.CardID != nil && len(dto.CardID) != 0 {
		query.Where("card_id IN (?)", dto.CardID)
	}

	if dto.DateFrom != nil {
		query.Where("date_paid >= ?", *dto.DateFrom)
	}
	if dto.DateTo != nil {
		query.Where("date_paid < ?", *dto.DateTo)
	}

	if dto.OrderStatuses != nil && len(dto.OrderStatuses) != 0 {
		query.Where("status IN (?)", dto.OrderStatuses)
	}

	pattern := "(date_paid - (date_paid MOD %d)) AS date_group, SUM(amount) AS total, card_id AS card, status"
	//if dto.IsGroupByHour() {
	//	interval := 3600
	//	query.Select(fmt.Sprintf(pattern, interval))
	//} else if dto.IsGroupByDay() {
	//	interval := 86400
	//	query.Select(fmt.Sprintf(pattern, interval))
	//} else if dto.IsGroupByWeek() {
	//	interval := 604800
	//	query.Select(fmt.Sprintf(pattern, interval))
	//} else if dto.IsGroupByMonth() {
	//	format := "%Y-%m"
	//	query.Select("(DATE_FORMAT(date_paid, ?)) AS date_group, SUM(amount) AS total, card_id AS card", format)
	//}
	var interval int
	if dto.IsGroupByHour() {
		interval = 3600
	} else {
		interval = 86400
	}

	query.Select(fmt.Sprintf(pattern, interval))

	query.Group("date_group, card, status")
	query.Order("date_group ASC")

	var result []*TotalsResultDto
	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}

	var filtered = make([]*TotalsResultDto, 0)
	for _, r := range result {
		if r.DateGroup != 0 {
			filtered = append(filtered, r)
		}
	}

	return filtered, nil
}
