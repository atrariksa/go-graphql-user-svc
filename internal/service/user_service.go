package service

import (
	"context"
	"errors"
	"go-graphql-user-svc/config"
	"go-graphql-user-svc/internal/model"
	"go-graphql-user-svc/internal/repository"
	"go-graphql-user-svc/util"
)

type IUserService interface {
	Login(ctx context.Context, user model.User) (string, error)
	GetAllUser(ctx context.Context) *[]model.User
	CreateUser(ctx context.Context, user model.User) (*model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	UpdateUser(ctx context.Context, id string, user model.User) (*model.User, error)
	DeleteUser(ctx context.Context, id string) error
}

type UserService struct {
	Repo repository.IUserRepository
	Cfg  *config.Config
}

// NewUserService creates a new service instance for user-related operations
func NewUserService(repo repository.IUserRepository, cfg *config.Config) *UserService {
	return &UserService{
		Repo: repo,
		Cfg:  cfg,
	}
}

func (s *UserService) Login(ctx context.Context, user model.User) (string, error) {
	userData, err := s.Repo.FindByEmail(ctx, user.Email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}
	if userData.ID == "" || user.Password == "" {
		return "", errors.New("invalid email or password")
	}
	if util.ValidatePassword(user.Password, userData.Password) != nil {
		return "", errors.New("invalid email or password")
	}
	token, err := util.GenerateToken(
		util.Claims{ID: string(userData.ID), Role: userData.Role},
		s.Cfg.AuthTokenConfig.Duration,
		s.Cfg.AuthTokenConfig.SecretKey,
	)
	if err != nil {
		return "", errors.New("internal error")
	}
	return token, nil
}

// GetAllUser calls the repository to get all user
func (s *UserService) GetAllUser(ctx context.Context) *[]model.User {
	return s.Repo.Getall(ctx)
}

// CreateUser calls the repository to create a new user
func (s *UserService) CreateUser(ctx context.Context, user model.User) (*model.User, error) {
	hashedPassword, _ := util.HashPassword(user.Password)
	user.Password = hashedPassword
	return s.Repo.Create(ctx, user)
}

// GetUserByID calls the repository to get a user by its ID
func (s *UserService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	return s.Repo.FindByID(ctx, id)
}

// UpdateUser calls the repository to update a user's data
func (s *UserService) UpdateUser(ctx context.Context, id string, user model.User) (*model.User, error) {
	hashedPassword, _ := util.HashPassword(user.Password)
	user.Password = hashedPassword
	return s.Repo.Update(ctx, id, user)
}

// DeleteUser calls the repository to delete a user by its ID
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.Repo.Delete(ctx, id)
}
