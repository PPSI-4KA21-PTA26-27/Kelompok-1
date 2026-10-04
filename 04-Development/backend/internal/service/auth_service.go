package service

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"kel1/backend/internal/dto"
	"kel1/backend/internal/models"
	"kel1/backend/internal/repository"
	"kel1/backend/internal/utils"
)

var (
	ErrEmailTaken         = errors.New("email sudah terdaftar")
	ErrInvalidCredentials = errors.New("email atau password salah")
	ErrUserNotFound       = errors.New("user tidak ditemukan")
)

type AuthService interface {
	Register(req dto.RegisterRequest) (*models.User, error)
	Login(req dto.LoginRequest) (string, *models.User, error)
}

type authService struct {
	users     repository.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewAuthService(users repository.UserRepository, jwtSecret string, jwtExpiry time.Duration) AuthService {
	return &authService{users: users, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

func (s *authService) Register(req dto.RegisterRequest) (*models.User, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if _, err := s.users.FindByEmail(email); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:     strings.TrimSpace(req.Name),
		Email:    email,
		Password: hash,
		Role:     models.RoleUser,
	}
	if err := s.users.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) Login(req dto.LoginRequest) (string, *models.User, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.users.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}

	if !utils.CheckPassword(user.Password, req.Password) {
		return "", nil, ErrInvalidCredentials
	}

	token, err := utils.GenerateToken(s.jwtSecret, s.jwtExpiry, user.ID, user.Role)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}
