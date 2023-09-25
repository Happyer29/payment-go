package services

import (
	"errors"
	"fmt"
	"payment-go/internal/config"
	"payment-go/internal/models"
	"payment-go/internal/repositories"
	"payment-go/internal/transport/bank/webhook"
	"payment-go/internal/transport/model/card"
	"payment-go/internal/utils/card_manager"
	"sync"
	"time"
)

type ICardService interface {
	ChooseCard(ord *models.Order) (card_manager.ISafeCard, error)
	CreateFromDto(dto *card.CreateCardDto) (*models.Card, error)
	UpdateFromDto(dto *card.UpdateCardDto) (*models.Card, error)
	GetCardInfo(cardID uint) (*models.CardInfo, error)
	GetCardUseStatus(crd *models.Card) (string, float64)
	ChangeCardBalance(req card.IChangeBalanceRequest) error
	ReceivePayment(dto *webhook.BankPaymentInfoDto) error
}
type cardService struct {
	cardManager card_manager.ICardManager
}

var cardIns *cardService
var cardOnce = sync.Once{}

var ErrDifferentAmount = errors.New("got different amount")

func CardService() ICardService {
	cardOnce.Do(func() {
		cardIns = &cardService{}
		cardIns.init()
	})
	return cardIns
}

// resetCardManager пересоздаёт менеджер карт с актуальным списком активных карт
func (s *cardService) resetCardManager() {
	cards, err := repositories.CardRepository().GetAllActiveCards()
	if err != nil {
		cards = []*models.Card{}
	}
	for _, crd := range cards {
		_ = repositories.CardRepository().AssertInfoExists(crd.ID)
	}
	s.cardManager = card_manager.NewCardManager(cards)
	fmt.Println(fmt.Sprintf("CardManager resetting. %d cards in use", s.cardManager.CardsCount()))

	if s.cardManager.CardsCount() == 0 {
		// запустим событие
		go EventService().NoCardsAvailable(nil)
	}
}

// ChooseCard получает и возвращает подходящую карту из менеджера карт
func (s *cardService) ChooseCard(ord *models.Order) (card_manager.ISafeCard, error) {
	// получаем оптимальную карту
	c, err := s.cardManager.GetNextCard(ord)
	if err != nil {
		if err == card_manager.ErrNoCardsAvailable {
			// запустим событие
			go EventService().NoCardsAvailable(&ord.PaymentMethod)
		}
		return nil, err
	}

	return c, nil
}

func (s *cardService) init() {
	// запустим автоматическую актуализацию менеджера карт
	var recursiveManagerResetting func()
	recursiveManagerResetting = func() {
		s.resetCardManager()
		go func() {
			time.Sleep(config.TaskReloadCardsInterval)
			go recursiveManagerResetting() // рекурсия
		}()
	}
	recursiveManagerResetting()
}

func (s *cardService) CreateFromDto(dto *card.CreateCardDto) (*models.Card, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	crd, err := dto.ToCard()
	if err != nil {
		return nil, err
	}
	if err := repositories.CardRepository().Save(crd); err != nil {
		return nil, err
	}
	return crd, nil
}

func (s *cardService) UpdateFromDto(dto *card.UpdateCardDto) (*models.Card, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	crd, err := repositories.CardRepository().FindById(dto.ID)
	if err != nil {
		return nil, err
	}

	crd.PhonePrefix = dto.PhonePrefix
	crd.PhoneNumber = dto.PhoneNumber
	crd.CardNumber = dto.CardNumber
	crd.Status = dto.Status

	err = repositories.CardRepository().Save(crd)
	if err != nil {
		return crd, err
	}
	return crd, nil
}

func (s *cardService) GetCardUseStatus(crd *models.Card) (string, float64) {
	var chance float64 = 0
	isUsed := s.cardManager.IsCardUsed(crd.ID)
	if isUsed {
		chance = s.cardManager.GetCardChance(crd.ID)
	}

	if isUsed && crd.IsActive() {
		return "used", chance
	} else if !isUsed && !crd.IsActive() {
		return "not_used", chance
	} else if isUsed {
		return "removing", chance
	} else {
		return "adding", chance
	}
}

func (s *cardService) ChangeCardBalance(req card.IChangeBalanceRequest) error {
	if req == nil {
		return fmt.Errorf("invalid request")
	}
	if req.Amount() <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	crd, err := repositories.CardRepository().FindById(req.CardID())
	if err != nil {
		return err
	}

	if req.IsIncrease() {
		if err := repositories.CardRepository().IncreaseBalance(crd.ID, req.Amount()); err != nil {
			return err
		}
		// запустим событие
		go EventService().CardBalanceIncreased(crd)
	} else {
		return repositories.CardRepository().DecreaseBalance(crd.ID, req.Amount())
	}

	return nil
}

func (s *cardService) GetCardInfo(cardID uint) (*models.CardInfo, error) {
	info, err := repositories.CardRepository().GetCardInfo(cardID)
	if err != nil {
		return nil, fmt.Errorf("card information not found")
	}

	return info, nil
}

func (s *cardService) ReceivePayment(dto *webhook.BankPaymentInfoDto) error {
	msg := dto.ToBankMessage()
	ordId, err := s.receivePayment(dto)
	if ord, err1 := repositories.OrderRepository().FindById(ordId); err1 == nil {
		msg.ShopID = ord.ShopID
	}
	msg.OrderID = ordId

	if err != nil {
		msg.Error = err.Error()
		if err == ErrDifferentAmount {
			msg.SetStatus(models.BankMessageStatusPendingApproval)
		} else {
			msg.SetStatus(models.BankMessageStatusError)
		}
	} else {
		msg.SetStatus(models.BankMessageStatusSuccess)
	}

	// попробуем сохранить инфу в бд, ошибка не критична
	_ = repositories.BankMessageRepository().Save(msg)
	return err
}

// receivePayment ordId uint, error
func (s *cardService) receivePayment(dto *webhook.BankPaymentInfoDto) (uint, error) {
	crd, err := repositories.CardRepository().FindByCardNumber(dto.CardNumber)
	if err != nil {
		return 0, fmt.Errorf("card not found")
	}

	lock := card_manager.CardLocker().GetLocked(crd.ID)
	if lock == nil {
		return 0, fmt.Errorf("card with this number is not waiting for payment")
	}
	ordId := lock.GetOrderId()
	lock.Unlock()

	if ordId != 0 {
		ord, err := repositories.OrderRepository().FindById(ordId)
		if err != nil {
			return ordId, fmt.Errorf("order not found")
		}

		// если сумма не совпадает с ожидаемой
		if ord.Amount != dto.Amount {
			return ordId, ErrDifferentAmount
		}

		now := uint(time.Now().Unix())
		err = OrderService().FinishOrderWithStatus(ord, models.StatusCompleted, now)
		if err != nil {
			return ordId, fmt.Errorf("unable to finish order")
		}
	}
	return ordId, nil
}
