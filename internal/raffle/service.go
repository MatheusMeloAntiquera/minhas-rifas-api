package raffle

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/matheusantiquera/minhas-rifas/domain"
	"github.com/matheusantiquera/minhas-rifas/internal/user"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrUserNotFound   = errors.New("usuário não encontrado")
	ErrRaffleNotFound = errors.New("rifa não encontrada")
)

type TicketRepository interface {
	CountByRaffle(ctx context.Context, raffleID int) (int64, error)
}

type Service interface {
	Create(ctx context.Context, input CreateInput) (domain.Raffle, error)
	ListByUser(ctx context.Context, userID int) ([]domain.Raffle, error)
	Get(ctx context.Context, id int) (GetResponse, error)
}

type service struct {
	validate         *validator.Validate
	repository       Repository
	userRepository   user.Repository
	ticketRepository TicketRepository
	logger           *slog.Logger
}

func NewService(validate *validator.Validate, repository Repository, userRepository user.Repository, ticketRepository TicketRepository, logger *slog.Logger) Service {
	return &service{
		validate:         validate,
		repository:       repository,
		userRepository:   userRepository,
		ticketRepository: ticketRepository,
		logger:           logger,
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

func (s *service) ListByUser(ctx context.Context, userID int) ([]domain.Raffle, error) {
	_, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		s.logger.Error("usuário não encontrado para listagem de rifas", "error", err, "user_id", userID)
		return nil, ErrUserNotFound
	}

	raffles, err := s.repository.ListByUser(ctx, userID)
	if err != nil {
		s.logger.Error("falha ao listar rifas do usuário", "error", err, "user_id", userID)
		return nil, err
	}

	return raffles, nil
}

func (s *service) Get(ctx context.Context, id int) (GetResponse, error) {
	raffle, err := s.repository.FindByID(ctx, id)
	if err != nil {
		//TODO: mover essa validação para o repositorio
		if errors.Is(err, mongo.ErrNoDocuments) {
			return GetResponse{}, ErrRaffleNotFound
		}
		s.logger.Error("falha ao buscar rifa", "error", err, "id", id)
		return GetResponse{}, err
	}

	ticketsSold, err := s.ticketRepository.CountByRaffle(ctx, id)
	if err != nil {
		s.logger.Error("falha ao contar tickets da rifa", "error", err, "id", id)
		return GetResponse{}, err
	}

	return GetResponse{Raffle: *raffle, TicketsSold: ticketsSold}, nil
}
