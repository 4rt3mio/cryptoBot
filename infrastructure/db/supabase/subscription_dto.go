package supabase

import (
	"time"

	"github.com/4rt3mio/cryptoCore/domain/model"
)

type SubscriptionDTO struct {
	ID          int       `db:"id"`
	UserID      string    `db:"user_id"`
	TokenName   string    `db:"token_name"`
	TokenSymbol string    `db:"token_symbol"`
	Threshold   float64   `db:"threshold"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (dto *SubscriptionDTO) ToDomain() model.Subscription {
	return model.Subscription{
		ID:     dto.ID,
		UserID: dto.UserID,
		Token: model.Token{
			Name:      dto.TokenName,
			Symbol:    dto.TokenSymbol,
			Threshold: dto.Threshold,
		},
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}

func FromDomain(sub model.Subscription) *SubscriptionDTO {
	return &SubscriptionDTO{
		ID:          sub.ID,
		UserID:      sub.UserID,
		TokenName:   sub.Token.Name,
		TokenSymbol: sub.Token.Symbol,
		Threshold:   sub.Token.Threshold,
		CreatedAt:   sub.CreatedAt,
		UpdatedAt:   sub.UpdatedAt,
	}
}
