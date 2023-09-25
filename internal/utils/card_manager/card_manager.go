package card_manager

import (
	"errors"
	"fmt"
	"math/rand"
	"payment-go/internal/models"
	"sort"
	"sync"
)

type ICardManager interface {
	GetNextCard(ord *models.Order) (ISafeCard, error)
	GetCardChance(id uint) float64
	CardsCount() int
	IsCardUsed(id uint) bool
}
type cardManager struct {
	cardsCount int
	cards      []priorityCard
	limit      int
}

type priorityCard struct {
	card     models.Card
	mu       *sync.RWMutex
	isLocked bool
	limit    int
	delta    float64
}

var ErrNoCardsAvailable = errors.New("unable to choose card for selected payment method")

func NewCardManager(cards []*models.Card) ICardManager {
	ins := &cardManager{
		cardsCount: len(cards),
	}
	pCards := ins.sortCards(cards)
	if pCards != nil {
		ins.cards = pCards
	} else {
		ins.cardsCount = 0
	}
	return ins
}

func (cm *cardManager) sortCards(cards []*models.Card) []priorityCard {
	sums := cm.getSums(cards)

	// создадим список карт с указанием их breakpoint для дальнейшего использования в выдаче карт
	var limit int
	var pCards = make([]priorityCard, cm.cardsCount)
	for _, card := range cards {
		delta, ok := sums[card.Stats.TotalPaymentSum]
		if !ok {
			return nil
		}
		limit += delta
		pCards = append(pCards, priorityCard{
			card:     *card,
			mu:       &sync.RWMutex{},
			isLocked: false,
			limit:    limit,
			delta:    float64(delta),
		})
	}
	cm.limit = limit

	return pCards
}

// getSums возвращает мапу из сумм и их порядковых номеров (по убыванию)
func (cm *cardManager) getSums(cards []*models.Card) map[float64]int {
	// создадим множество сумм
	var sumSet = make([]float64, 0)
	for _, card := range cards {
		sum := card.Stats.TotalPaymentSum
		for _, s := range sumSet {
			if s == sum {
				continue
			}
		}
		sumSet = append(sumSet, sum)
	}

	// отсортируем множество сумм в обратном порядке
	sort.Slice(sumSet, func(i, j int) bool {
		return sumSet[i] > sumSet[j]
	})

	// присвоим каждой сумме K равный её порядковому номеру (K >= 1)
	var sumsMap = make(map[float64]int)
	for i, sum := range sumSet {
		sumsMap[sum] = i + 1
	}
	return sumsMap
}

func (cm *cardManager) tryGetCard(card priorityCard, ord *models.Order) ISafeCard {
	if !card.card.SupportsPaymentMethod(ord.PaymentMethod) {
		return nil
	}

	if lock, err := CardLocker().LockCard(&card.card); err == nil {
		return lock
	} else {
		return nil
	}
}

func (cm *cardManager) GetNextCard(ord *models.Order) (ISafeCard, error) {
	if cm.cardsCount <= 0 {
		return nil, fmt.Errorf("no cards")
	}
	x := rand.Intn(cm.limit)
	for _, card := range cm.cards {
		if x < card.limit {
			if crd := cm.tryGetCard(card, ord); crd != nil {
				return crd, nil
			}
		}
	}

	// если вдруг карта не была найдена, вернём любую подходящую
	for _, card := range cm.cards {
		if crd := cm.tryGetCard(card, ord); crd != nil {
			return crd, nil
		}
	}

	return nil, ErrNoCardsAvailable
}

func (cm *cardManager) CardsCount() int {
	return cm.cardsCount
}

func (cm *cardManager) IsCardUsed(id uint) bool {
	for _, card := range cm.cards {
		if card.card.ID == id {
			return true
		}
	}
	return false
}

func (cm *cardManager) GetCardChance(id uint) float64 {
	for _, card := range cm.cards {
		if card.card.ID == id {
			return card.delta / float64(cm.limit)
		}
	}
	return 0
}
