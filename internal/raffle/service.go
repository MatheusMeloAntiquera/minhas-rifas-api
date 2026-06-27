package raffle

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/matheusantiquera/minhas-rifas/domain"
	"github.com/matheusantiquera/minhas-rifas/internal/user"
)

var (
	ErrUserNotFound = errors.New("usuário não encontrado")
)

type Service interface {
	Create(ctx context.Context, input CreateInput) (domain.Raffle, error)
}

type service struct {
	validate       *validator.Validate
	repository     Repository
	userRepository user.Repository
	logger         *slog.Logger
}

func NewService(validate *validator.Validate, repository Repository, userRepository user.Repository, logger *slog.Logger) Service {
	return &service{
		validate:       validate,
		repository:     repository,
		userRepository: userRepository,
		logger:         logger,
	}
}

func (s *service) Create(ctx context.Context, input CreateInput) (domain.Raffle, error) {
	if err := s.validate.Struct(input); err != nil {
		return domain.Raffle{}, err
	}

	_, err := s.userRepository.FindByID(ctx, input.UserID)
	if err != nil {
		s.logger.Error("usuário não encontrado para criação de rifa", "error", err, "user_id", input.UserID)
		return domain.Raffle{}, ErrUserNotFound
	}

	now := time.Now()
	raffle := domain.Raffle{
		Title:       input.Title,
		Description: input.Description,
		ValueTicket: input.ValueTicket,
		UserID:      input.UserID,
		DrawDate:    input.DrawDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return s.repository.Create(ctx, raffle)
}
