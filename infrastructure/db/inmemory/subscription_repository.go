package inmemory

import (
	"errors"
	"sync"
	"time"

	"github.com/4rt3mio/cryptoCore/domain/model"
	"github.com/4rt3mio/cryptoCore/domain/repository"
)

type InMemorySubscriptionRepo struct {
	data   map[int]model.Subscription
	mu     sync.RWMutex
	nextID int
}

func NewInMemorySubscriptionRepo() repository.SubscriptionRepository {
    return &InMemorySubscriptionRepo{
        data:   make(map[int]model.Subscription),
        nextID: 1,
    }
}

func (r *InMemorySubscriptionRepo) Add(sub model.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	sub.ID = r.nextID
	r.nextID++
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()
	r.data[sub.ID] = sub
	return nil
}

func (r *InMemorySubscriptionRepo) Remove(userID string, subID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	sub, exists := r.data[subID]
	if !exists || sub.UserID != userID {
		return errors.New("подписка не найдена")
	}
	delete(r.data, subID)
	return nil
}

func (r *InMemorySubscriptionRepo) Update(userID string, subID int, newThreshold float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	sub, exists := r.data[subID]
	if !exists || sub.UserID != userID {
		return errors.New("подписка не найдена")
	}
	sub.Token.Threshold = newThreshold
	sub.UpdatedAt = time.Now()
	r.data[subID] = sub
	return nil
}

func (r *InMemorySubscriptionRepo) List(userID string) ([]model.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []model.Subscription
	for _, sub := range r.data {
		if sub.UserID == userID {
			result = append(result, sub)
		}
	}
	return result, nil
}

func (r *InMemorySubscriptionRepo) ListAll() ([]model.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []model.Subscription
	for _, sub := range r.data {
		result = append(result, sub)
	}
	return result, nil
}