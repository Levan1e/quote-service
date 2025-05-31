package service

import (
	"context"
	"quote-service/internal/domain"
	"quote-service/internal/models"
)

type Querier interface {
	Create(ctx context.Context, quote *models.Quote) error
	GetAll(ctx context.Context) ([]models.Quote, error)
	GetRandom(ctx context.Context) (*models.Quote, error)
	GetByAuthor(ctx context.Context, author string) ([]models.Quote, error)
	Delete(ctx context.Context, id int) error
	Exists(ctx context.Context, author, quote string) (bool, error)
}

type QuoteService struct {
	repo Querier
}

func NewQuoteService(repo Querier) *QuoteService {
	return &QuoteService{repo: repo}
}

func (s *QuoteService) Create(ctx context.Context, quote *models.Quote) error {
	if quote.Author == "" || quote.Quote == "" {
		return domain.ErrInvalidInput
	}
	return s.repo.Create(ctx, quote)
}

func (s *QuoteService) GetAll(ctx context.Context) ([]models.Quote, error) {
	return s.repo.GetAll(ctx)
}

func (s *QuoteService) GetRandom(ctx context.Context) (*models.Quote, error) {
	return s.repo.GetRandom(ctx)
}

func (s *QuoteService) GetByAuthor(ctx context.Context, author string) ([]models.Quote, error) {
	if author == "" {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.GetByAuthor(ctx, author)
}

func (s *QuoteService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return domain.ErrInvalidInput
	}
	return s.repo.Delete(ctx, id)
}

func (s *QuoteService) Exists(ctx context.Context, author, quote string) (bool, error) {
	if author == "" || quote == "" {
		return false, domain.ErrInvalidInput
	}
	return s.repo.Exists(ctx, author, quote)
}
