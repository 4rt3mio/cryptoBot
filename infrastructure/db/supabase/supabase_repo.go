package supabase

import (
	"context"
	"errors"
	"time"

	"CourseWork/infrastructure/logger"
	"github.com/4rt3mio/cryptoCore/domain/model"
	"github.com/4rt3mio/cryptoCore/domain/repository"

	"github.com/jmoiron/sqlx"
)

type SubscriptionRepository struct {
	log *logger.ZapLogger
	db  *sqlx.DB
}

func NewSubscriptionRepository(db *sqlx.DB, log *logger.ZapLogger) repository.SubscriptionRepository {
	return &SubscriptionRepository{log: log, db: db}
}

func (r *SubscriptionRepository) Add(sub model.Subscription) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	dto := FromDomain(sub)
	dto.CreatedAt = time.Now()
	dto.UpdatedAt = time.Now()

	query := r.db.Rebind(`
		INSERT INTO subscriptions (user_id, token_name, token_symbol, threshold, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id
	`)

	err := r.db.Unsafe().QueryRowContext(ctx, query,
		dto.UserID,
		dto.TokenName,
		dto.TokenSymbol,
		dto.Threshold,
		dto.CreatedAt,
		dto.UpdatedAt,
	).Scan(&dto.ID)

	if err != nil {
		r.log.Error("SubscriptionRepository.Add: http.Post failed", "err", err)
		return err
	}

	sub.ID = dto.ID
	return nil
}

func (r *SubscriptionRepository) Remove(userID string, subID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	exists, err := r.subscriptionExists(userID, subID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("subscription not found")
	}

	query := r.db.Rebind(`
		DELETE FROM subscriptions 
		WHERE id = ? AND user_id = ?
	`)

	result, err := r.db.Unsafe().ExecContext(ctx, query, subID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("subscription not found")
	}

	return nil
}

func (r *SubscriptionRepository) Update(userID string, subID int, newThreshold float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	exists, er := r.subscriptionExists(userID, subID)
	if er != nil {
		return er
	}
	if !exists {
		return errors.New("subscription not found")
	}

	query := r.db.Rebind(`
		UPDATE subscriptions 
		SET threshold = ?, updated_at = ? 
		WHERE id = ? AND user_id = ?
	`)

	_, err := r.db.Unsafe().ExecContext(ctx, query,
		newThreshold,
		time.Now(),
		subID,
		userID,
	)
	return err
}

func (r *SubscriptionRepository) List(userID string) ([]model.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dtos []SubscriptionDTO
	query := r.db.Rebind("SELECT * FROM subscriptions WHERE user_id = ?")

	err := r.db.Unsafe().SelectContext(ctx, &dtos, query, userID)
	if err != nil {
		return nil, err
	}

	subs := make([]model.Subscription, len(dtos))
	for i, dto := range dtos {
		subs[i] = dto.ToDomain()
	}
	return subs, nil
}

func (r *SubscriptionRepository) ListAll() ([]model.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dtos []SubscriptionDTO
	err := r.db.Unsafe().SelectContext(ctx, &dtos, "SELECT * FROM subscriptions")
	if err != nil {
		return nil, err
	}

	subs := make([]model.Subscription, len(dtos))
	for i, dto := range dtos {
		subs[i] = dto.ToDomain()
	}
	return subs, nil
}

func (r *SubscriptionRepository) subscriptionExists(userID string, subID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := r.db.Rebind(`
        SELECT EXISTS(
            SELECT 1 
            FROM subscriptions 
            WHERE id = ? AND user_id = ?
        )
    `)

	var exists bool
	err := r.db.Unsafe().GetContext(ctx, &exists, query, subID, userID)
	return exists, err
}
