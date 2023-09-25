package card_manager

import (
	"fmt"
	"payment-go/internal/config"
	"payment-go/internal/models"
	"sync"
	"time"
)

type ICardLocker interface {
	GetLocked(cardId uint) ISafeCard
	LockCard(crd *models.Card) (ISafeCard, error)
	UnlockCard(cardId uint)
	UnlockCardByOrderId(orderId uint)
	IsLocked(cardId uint) bool
}
type cardLocker struct {
	mu    sync.RWMutex
	cards []*lockedCard
}

var lockIns *cardLocker
var lockOnce = sync.Once{}

func CardLocker() ICardLocker {
	lockOnce.Do(func() {
		lockIns = &cardLocker{
			mu:    sync.RWMutex{},
			cards: make([]*lockedCard, 0),
		}

		go lockIns.gc()
	})
	return lockIns
}

func (cl *cardLocker) gc() {
	time.Sleep(config.CardLockerGCInterval)

	cl.mu.Lock()
	defer cl.mu.Unlock()

	newList := make([]*lockedCard, 0)
	for _, lock := range cl.cards {
		if time.Now().Unix() <= lock.unlockTime.Unix() {
			newList = append(newList, lock)
		}
	}
	cl.cards = newList

	go cl.gc()
}

func (cl *cardLocker) LockCard(crd *models.Card) (ISafeCard, error) {
	if crd == nil {
		return nil, fmt.Errorf("unable to lock nil card")
	}

	if !crd.CanBeLocked() {
		return &defaultCard{card: crd}, nil
	}

	lock := &lockedCard{
		id:         crd.ID,
		card:       crd,
		orderId:    0,
		unlockTime: time.Now().Add(config.CardLockingTimeout),
	}

	cl.mu.Lock()
	defer cl.mu.Unlock()

	alreadyLocked := false
	for _, lock := range cl.cards {
		if lock.id == crd.ID {
			alreadyLocked = true
			break
		}
	}
	if alreadyLocked {
		return nil, fmt.Errorf("card is already locked")
	}
	cl.cards = append(cl.cards, lock)

	return lock, nil
}

func (cl *cardLocker) UnlockCard(cardId uint) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	for i, lock := range cl.cards {
		if lock.id == cardId {
			cl.cards = append(cl.cards[:i], cl.cards[i+1:]...)
			break
		}
	}
}

func (cl *cardLocker) UnlockCardByOrderId(orderId uint) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	for i, lock := range cl.cards {
		if lock.orderId == orderId {
			cl.cards = append(cl.cards[:i], cl.cards[i+1:]...)
			break
		}
	}
}

func (cl *cardLocker) IsLocked(cardId uint) bool {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	for _, lock := range cl.cards {
		if lock.id == cardId {
			return true
		}
	}
	return false
}

func (cl *cardLocker) GetLocked(cardId uint) ISafeCard {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	for _, lock := range cl.cards {
		if lock.id == cardId {
			return lock
		}
	}
	return nil
}

type ISafeCard interface {
	GetCard() *models.Card
	GetOrderId() uint
	SetOrderId(orderId uint)
	Unlock()
}
type lockedCard struct {
	id         uint
	card       *models.Card
	orderId    uint
	unlockTime time.Time
}

func (lc *lockedCard) GetCard() *models.Card {
	return lc.card
}
func (lc *lockedCard) SetOrderId(orderId uint) {
	lc.orderId = orderId
}
func (lc *lockedCard) GetOrderId() uint {
	return lc.orderId
}
func (lc *lockedCard) Unlock() {
	CardLocker().UnlockCard(lc.id)
}

type defaultCard struct {
	card *models.Card
}

func (lc *defaultCard) GetCard() *models.Card {
	return lc.card
}
func (lc *defaultCard) SetOrderId(uint) {}
func (lc *defaultCard) GetOrderId() uint {
	return 0
}
func (lc *defaultCard) Unlock() {}
