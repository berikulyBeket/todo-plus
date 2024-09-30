package usecase

import (
	"context"

	"github.com/berikulyBeket/todo-plus/internal/repository"
)

// UserUseCase handles the business logic related to users
type UserUseCase struct {
	repo repository.User
}

// NewAuthUseCase creates a new instance of AuthUseCase
func NewUserUseCase(r repository.User) *UserUseCase {
	return &UserUseCase{
		repo: r,
	}
}

// DeleteOneByAdmin deletes a user by their ID, typically used by an admin
func (uc *UserUseCase) DeleteOneByAdmin(ctx context.Context, userId int) error {
	return uc.repo.DeleteOneById(ctx, userId)
}
