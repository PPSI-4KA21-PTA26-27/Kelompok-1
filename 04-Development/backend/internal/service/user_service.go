package service

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"kel1/backend/internal/dto"
	"kel1/backend/internal/models"
	"kel1/backend/internal/repository"
)

type UserService interface {
	GetByID(id uint) (*models.User, error)
	List(p dto.PaginationQuery) ([]models.User, int64, error)
	Update(id uint, req dto.UpdateUserRequest) (*models.User, error)
	Delete(id uint) error
}

type userService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) UserService {
	return &userService{users: users}
}

func (s *userService) GetByID(id uint) (*models.User, error) {
	user, err := s.users.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) List(p dto.PaginationQuery) ([]models.User, int64, error) {
	p.Normalize()
	return s.users.List(p.Offset(), p.Limit)
}

func (s *userService) Update(id uint, req dto.UpdateUserRequest) (*models.User, error) {
	user, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if name := strings.TrimSpace(req.Name); name != "" {
		user.Name = name
	}
	if req.Role != "" {
		user.Role = req.Role
	}

	if err := s.users.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Delete(id uint) error {
	if _, err := s.GetByID(id); err != nil {
		return err
	}
	return s.users.Delete(id)
}
