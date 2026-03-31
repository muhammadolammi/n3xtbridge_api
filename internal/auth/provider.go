package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/muhammadolammi/goauth"
	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
)

// Provider implements auth.IdentityProvider
type Provider struct {
	q *database.Queries
}

func NewProvider(q *database.Queries) *Provider {
	return &Provider{
		q: q,
	}
}
func (p Provider) GetByEmail(ctx context.Context, email string) (*goauth.User, error) {
	user, err := p.q.GetUserByEmail(ctx, email)
	if err != nil {
		return &goauth.User{}, err
	}
	return &goauth.User{
		ID:           user.ID,
		PasswordHash: user.PasswordHash,
	}, nil
}
func (p Provider) GetByID(ctx context.Context, id uuid.UUID) (*goauth.User, error) {
	user, err := p.q.GetUserByID(ctx, id)
	if err != nil {
		return &goauth.User{}, err
	}
	return &goauth.User{
		ID:           user.ID,
		PasswordHash: user.PasswordHash,
	}, nil
}
func (p Provider) CreateRefreshToken(ctx context.Context, arg *goauth.CreateRefreshTokenParams) (*goauth.RefreshToken, error) {
	args := database.CreateRefreshTokenParams{
		UserID:    arg.UserID,
		ExpiresAt: arg.ExpiresAt,
		Token:     arg.Token,
	}
	dbRefreshToken, err := p.q.CreateRefreshToken(ctx, args)
	if err != nil {
		return &goauth.RefreshToken{}, err
	}
	return &goauth.RefreshToken{
		ID:         dbRefreshToken.ID,
		UserID:     dbRefreshToken.UserID,
		Token:      dbRefreshToken.Token,
		Revoked:    dbRefreshToken.Revoked,
		ReplacedBy: dbRefreshToken.ReplacedBy,
		CreatedAt:  dbRefreshToken.CreatedAt,
		ExpiresAt:  dbRefreshToken.ExpiresAt,
	}, nil
}
func (p Provider) GetRefreshToken(ctx context.Context, token string) (*goauth.RefreshToken, error) {
	dbRefreshToken, err := p.q.GetRefreshToken(ctx, token)
	if err != nil {
		return &goauth.RefreshToken{}, err
	}
	return &goauth.RefreshToken{
		ID:         dbRefreshToken.ID,
		UserID:     dbRefreshToken.UserID,
		Token:      dbRefreshToken.Token,
		Revoked:    dbRefreshToken.Revoked,
		ReplacedBy: dbRefreshToken.ReplacedBy,
		CreatedAt:  dbRefreshToken.CreatedAt,
		ExpiresAt:  dbRefreshToken.ExpiresAt,
	}, nil
}

func (p Provider) UpdateRefreshToken(ctx context.Context, arg *goauth.UpdateRefreshTokenParams) error {
	err := p.q.UpdateRefreshToken(ctx, database.UpdateRefreshTokenParams{
		ID:         arg.ID,
		ReplacedBy: arg.ReplacedBy,
		Revoked:    arg.Revoked,
	})
	return err
}
func (p Provider) RevokeUserTokens(ctx context.Context, userID uuid.UUID) error {
	err := p.q.RevokeRefreshTokens(ctx, userID)
	return err
}
