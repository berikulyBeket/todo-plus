package usecase

import (
	"context"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/internal/repository"

	"github.com/berikulyBeket/todo-plus/pkg/hash"
	"github.com/berikulyBeket/todo-plus/pkg/token"
)

// AuthUseCase handles the authentication logic
type AuthUseCase struct {
	repo       repository.Auth
	hasher     hash.Hasher
	tokenMaker token.TokenMaker
}

// NewAuthUseCase creates a new instance of AuthUseCase
func NewAuthUseCase(r repository.Auth, hasher hash.Hasher, tokenMaker token.TokenMaker) *AuthUseCase {
	return &AuthUseCase{
		repo:       r,
		hasher:     hasher,
		tokenMaker: tokenMaker,
	}
}

// CreateUser hashes the user's password and creates a new user in the repository
func (uc *AuthUseCase) CreateUser(ctx context.Context, user entity.User) (int, error) {
	user.Password = uc.hasher.HashPassword(user.Password)
	return uc.repo.CreateUser(ctx, user)
}

// AuthenticateUser checks the provided credentials by hashing the password and querying the repository
func (uc *AuthUseCase) AuthenticateUser(ctx context.Context, username, password string) (entity.User, error) {
	return uc.repo.GetUserByCredentials(ctx, username, uc.hasher.HashPassword(password))
}

// GenerateToken generates a token for the provided user Id
func (uc *AuthUseCase) GenerateToken(ctx context.Context, userId int) (string, error) {
	return uc.tokenMaker.CreateToken(userId)
}

// ParseToken verifies the provided access token and returns the associated user Id
func (uc *AuthUseCase) ParseToken(ctx context.Context, accessToken string) (int, error) {
	return uc.tokenMaker.VerifyToken(accessToken)
}
