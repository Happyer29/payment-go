package services

import (
	"fmt"
	"gorm.io/gorm"
	"payment-go/internal/config"
	"reflect"
	"sync"
	"time"
)

type ISubscriptionService interface {
	Subscribe(callback func(any), filters ...SubscriptionFilter) error
	RunSubscribers(entity any)
}
type subService struct {
	subs []*subscriber
	mu   *sync.RWMutex
}

type SubscriptionFilter func(entity any) bool

type subscriber struct {
	timestamp time.Time
	filters   []SubscriptionFilter
	callback  func(any)
}

var subIns *subService
var subOnce = sync.Once{}

func SubscriptionService() ISubscriptionService {
	subOnce.Do(func() {
		subIns = &subService{
			subs: make([]*subscriber, 0),
			mu:   &sync.RWMutex{},
		}

		// запустим garbage collector
		go subIns.gc()
	})
	return subIns
}

func (s *subService) gc() {
	s.mu.Lock()
	newList := make([]*subscriber, 0)
	for _, sub := range s.subs {
		if time.Since(sub.timestamp) < config.ModelSubscriptionLifetime {
			newList = append(newList, sub)
		}
	}
	s.subs = newList
	s.mu.Unlock()

	go func() {
		time.Sleep(config.SubscriptionGCInterval)
	}()
}

func (s *subService) Subscribe(callback func(any), filters ...SubscriptionFilter) error {
	if filters == nil || len(filters) == 0 || callback == nil {
		return fmt.Errorf("filter and callback are required")
	}

	sub := &subscriber{
		timestamp: time.Now(),
		filters:   filters,
		callback:  callback,
	}
	s.mu.Lock()
	s.subs = append(s.subs, sub)
	s.mu.Unlock()

	return nil
}

func (s *subService) RunSubscribers(entity any) {
	s.mu.Lock()
	newList := make([]*subscriber, 0)
	for _, sub := range s.subs {
		if s.entityMatchFilters(sub.filters, entity) {
			sub.callback(entity)
		} else {
			newList = append(newList, sub)
		}
	}
	s.subs = newList
	s.mu.Unlock()
}

func (s *subService) entityMatchFilters(filters []SubscriptionFilter, entity any) bool {
	for _, filter := range filters {
		if filter(entity) == false {
			return false
		}
	}
	return true
}

func FByID(id uint, model any) SubscriptionFilter {
	typ := reflect.TypeOf(model).Elem().String()
	if (len(typ)) < 5 {
		return func(entity any) bool {
			return false
		}
	}
	return func(model any) bool {
		if typ != reflect.TypeOf(model).Elem().String() {
			return false
		}
		v := reflect.ValueOf(model)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		givenID := v.FieldByName("ID").Uint()
		return givenID == 0 || givenID == uint64(id)
	}
}

// FSameEntity Сравнивает по указателям
func FSameEntity(model *gorm.Model) SubscriptionFilter {
	return func(model any) bool {
		entity, ok := model.(*gorm.Model)
		if !ok {
			return false
		}
		return model == entity
	}
}
