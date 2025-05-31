package service

import (
	"context"
	"quote-service/internal/domain"
	"quote-service/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) Create(ctx context.Context, quote *models.Quote) error {
	args := m.Called(ctx, quote)
	return args.Error(0)
}

func (m *MockQuerier) GetAll(ctx context.Context) ([]models.Quote, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Quote), args.Error(1)
}

func (m *MockQuerier) GetRandom(ctx context.Context) (*models.Quote, error) {
	args := m.Called(ctx)
	return args.Get(0).(*models.Quote), args.Error(1)
}

func (m *MockQuerier) GetByAuthor(ctx context.Context, author string) ([]models.Quote, error) {
	args := m.Called(ctx, author)
	return args.Get(0).([]models.Quote), args.Error(1)
}

func (m *MockQuerier) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQuerier) Exists(ctx context.Context, author, quote string) (bool, error) {
	args := m.Called(ctx, author, quote)
	return args.Bool(0), args.Error(1)
}

func TestQuoteService_Create(t *testing.T) {
	mockRepo := new(MockQuerier)
	service := NewQuoteService(mockRepo)

	quote := &models.Quote{
		Author: "Confucius",
		Quote:  "Life is simple",
	}

	t.Run("successful create", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, quote).Return(nil).Once()

		err := service.Create(context.Background(), quote)
		assert.NoError(t, err)
	})

	t.Run("invalid input", func(t *testing.T) {
		invalidQuote := &models.Quote{Author: "", Quote: ""}
		err := service.Create(context.Background(), invalidQuote)
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})
}

func TestQuoteService_GetAll(t *testing.T) {
	mockRepo := new(MockQuerier)
	service := NewQuoteService(mockRepo)

	quotes := []models.Quote{
		{ID: 1, Author: "Confucius", Quote: "Life is simple", CreatedAt: time.Now()},
	}

	mockRepo.On("GetAll", mock.Anything).Return(quotes, nil).Once()

	result, err := service.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, quotes[0].Author, result[0].Author)
}

func TestQuoteService_GetRandom(t *testing.T) {
	mockRepo := new(MockQuerier)
	service := NewQuoteService(mockRepo)

	quote := &models.Quote{ID: 1, Author: "Confucius", Quote: "Life is simple", CreatedAt: time.Now()}
	mockRepo.On("GetRandom", mock.Anything).Return(quote, nil).Once()

	result, err := service.GetRandom(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, quote.Author, result.Author)
}

func TestQuoteService_GetByAuthor(t *testing.T) {
	mockRepo := new(MockQuerier)
	service := NewQuoteService(mockRepo)

	quotes := []models.Quote{
		{ID: 1, Author: "Confucius", Quote: "Life is simple", CreatedAt: time.Now()},
	}

	t.Run("valid author", func(t *testing.T) {
		mockRepo.On("GetByAuthor", mock.Anything, "Confucius").Return(quotes, nil).Once()

		result, err := service.GetByAuthor(context.Background(), "Confucius")
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("empty author", func(t *testing.T) {
		result, err := service.GetByAuthor(context.Background(), "")
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
		assert.Nil(t, result)
	})
}

func TestQuoteService_Delete(t *testing.T) {
	mockRepo := new(MockQuerier)
	service := NewQuoteService(mockRepo)

	t.Run("successful delete", func(t *testing.T) {
		mockRepo.On("Delete", mock.Anything, 1).Return(nil).Once()

		err := service.Delete(context.Background(), 1)
		assert.NoError(t, err)
	})

	t.Run("invalid id", func(t *testing.T) {
		err := service.Delete(context.Background(), 0)
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})
}

func TestQuoteService_Exists(t *testing.T) {
	mockRepo := new(MockQuerier)
	service := NewQuoteService(mockRepo)

	t.Run("quote exists", func(t *testing.T) {
		mockRepo.On("Exists", mock.Anything, "Confucius", "Life is simple").Return(true, nil).Once()

		exists, err := service.Exists(context.Background(), "Confucius", "Life is simple")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("invalid input", func(t *testing.T) {
		exists, err := service.Exists(context.Background(), "", "")
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
		assert.False(t, exists)
	})
}
