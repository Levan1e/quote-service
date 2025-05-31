package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"quote-service/internal/domain"
	"quote-service/internal/models"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
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

// Тесты
func TestHandler_CreateQuote(t *testing.T) {
	mockQuerier := new(MockQuerier)
	handler := NewHandler(mockQuerier, zap.NewNop())

	quote := models.Quote{
		Author: "Confucius",
		Quote:  "Life is simple",
	}
	_, err := json.Marshal(quote)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("invalid input", func(t *testing.T) {
		invalidQuote := models.Quote{Author: "", Quote: ""}
		invalidJSON, err := json.Marshal(invalidQuote)
		if err != nil {
			t.Fatal(err)
		}

		mockQuerier.On("Exists", mock.Anything, "", "").Return(false, domain.ErrInvalidInput).Once()

		req := httptest.NewRequest(http.MethodPost, "/quotes", bytes.NewReader(invalidJSON))
		w := httptest.NewRecorder()

		handler.createQuote(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/quotes", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		handler.createQuote(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_GetAllQuotes(t *testing.T) {
	mockQuerier := new(MockQuerier)
	handler := NewHandler(mockQuerier, zap.NewNop())

	t.Run("successful get all", func(t *testing.T) {
		quotes := []models.Quote{
			{Author: "Confucius", Quote: "Life is simple"},
			{Author: "Socrates", Quote: "Know thyself"},
		}
		mockQuerier.On("GetAll", mock.Anything).Return(quotes, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/quotes", nil)
		w := httptest.NewRecorder()

		handler.getAllQuotes(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}
		data, ok := result["data"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, len(quotes), len(data))
		if len(data) > 0 {
			quote, ok := data[0].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, "Confucius", quote["author"])
		}
	})

	t.Run("empty result", func(t *testing.T) {
		mockQuerier.On("GetAll", mock.Anything).Return([]models.Quote{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/quotes", nil)
		w := httptest.NewRecorder()

		handler.getAllQuotes(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}
		data, ok := result["data"].([]interface{})
		assert.True(t, ok)
		assert.Empty(t, data)
	})

	t.Run("successful get by author", func(t *testing.T) {
		author := "Confucius"
		quotes := []models.Quote{{Author: author, Quote: "Life is simple"}}
		mockQuerier.On("GetByAuthor", mock.Anything, author).Return(quotes, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/quotes?author="+author, nil)
		w := httptest.NewRecorder()

		handler.getAllQuotes(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}
		data, ok := result["data"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, len(quotes), len(data))
		if len(data) > 0 {
			quote, ok := data[0].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, author, quote["author"])
		}
	})

	t.Run("invalid author", func(t *testing.T) {
		mockQuerier.On("GetAll", mock.Anything).Return([]models.Quote{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/quotes?author=", nil)
		w := httptest.NewRecorder()

		handler.getAllQuotes(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}
		data, ok := result["data"].([]interface{})
		assert.True(t, ok)
		assert.Empty(t, data)
	})

	t.Run("error from GetAll", func(t *testing.T) {
		mockQuerier.On("GetAll", mock.Anything).Return([]models.Quote{}, domain.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/quotes", nil)
		w := httptest.NewRecorder()

		handler.getAllQuotes(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("error from GetByAuthor", func(t *testing.T) {
		author := "Confucius"
		mockQuerier.On("GetByAuthor", mock.Anything, author).Return([]models.Quote{}, domain.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/quotes?author="+author, nil)
		w := httptest.NewRecorder()

		handler.getAllQuotes(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_GetRandomQuote(t *testing.T) {
	mockQuerier := new(MockQuerier)
	handler := NewHandler(mockQuerier, zap.NewNop())

	t.Run("successful get random", func(t *testing.T) {
		quote := &models.Quote{Author: "Confucius", Quote: "Life is simple"}
		mockQuerier.On("GetRandom", mock.Anything).Return(quote, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/quotes/random", nil)
		w := httptest.NewRecorder()

		handler.getRandomQuote(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}
		data, ok := result["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "Confucius", data["author"])
		assert.Equal(t, "Life is simple", data["quote"])
	})

	t.Run("no quotes available", func(t *testing.T) {
		mockQuerier.On("GetRandom", mock.Anything).Return((*models.Quote)(nil), domain.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/quotes/random", nil)
		w := httptest.NewRecorder()

		handler.getRandomQuote(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandler_DeleteQuote(t *testing.T) {
	mockQuerier := new(MockQuerier)
	handler := NewHandler(mockQuerier, zap.NewNop())

	t.Run("successful delete", func(t *testing.T) {
		id := 1
		mockQuerier.On("Delete", mock.Anything, id).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/quotes/1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.deleteQuote(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/quotes/invalid", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.deleteQuote(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("quote not found", func(t *testing.T) {
		id := 999
		mockQuerier.On("Delete", mock.Anything, id).Return(domain.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodDelete, "/quotes/999", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.deleteQuote(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
