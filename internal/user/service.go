package user

import (
	"context"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/matheusantiquera/minhas-rifas/domain"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrEmailAlreadyExists = errors.New("email já está em uso")
	ErrUserNotFound       = errors.New("usuário não encontrado")
)

type Service interface {
	Create(ctx context.Context, input CreateInput) (domain.User, error)
	Update(ctx context.Context, id int, input UpdateInput) (domain.User, error)
	Delete(ctx context.Context, id int) error
}

type service struct {
	validate   *validator.Validate
	repository Repository
}

func NewService(validate *validator.Validate, repository Repository) Service {
	return &service{
		validate:   validate,
		repository: repository,
	}
}

func (s *service) Create(ctx context.Context, input CreateInput) (domain.User, error) {
	if err := s.validate.Struct(input); err != nil {
		return domain.User{}, err
	}

	existing, err := s.repository.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return domain.User{}, err
	}
	if existing != nil {
		return domain.User{}, ErrEmailAlreadyExists
	}

	now := time.Now()
	user := domain.User{
		Name:      input.Name,
		Email:     input.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.repository.Create(ctx, user)
}

func (s *service) Update(ctx context.Context, id int, input UpdateInput) (domain.User, error) {
	if err := s.validate.Struct(input); err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		Name:      input.Name,
		Email:     input.Email,
		UpdatedAt: time.Now(),
	}

	updated, err := s.repository.Update(ctx, id, user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, err
	}

	return updated, nil
}

func (s *service) Delete(ctx context.Context, id int) error {
	err := s.repository.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}
