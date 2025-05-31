package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"quote-service/internal/domain"
	"quote-service/internal/models"
	"quote-service/internal/service"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	logger  *zap.Logger
	service *service.QuoteService
}

func NewHandler(db service.Querier, logger *zap.Logger) *Handler {
	return &Handler{
		logger:  logger,
		service: service.NewQuoteService(db),
	}
}

func (h *Handler) Routes() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", h.createQuote)         // POST /quotes
	r.Get("/", h.getAllQuotes)         // GET /quotes или GET /quotes?author={author}
	r.Get("/random", h.getRandomQuote) // GET /quotes/random
	r.Delete("/{id}", h.deleteQuote)   // DELETE /quotes/{id}
	return r
}

func (h *Handler) createQuote(w http.ResponseWriter, r *http.Request) {
	var quote models.Quote
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Ошибка чтения тела запроса", zap.Error(err))
		sendErrorResponse(w, domain.ErrInvalidInput.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	h.logger.Info("Request body", zap.String("body", string(body)))

	if err := json.Unmarshal(body, &quote); err != nil {
		h.logger.Error("Ошибка декодирования запроса", zap.Error(err))
		sendErrorResponse(w, domain.ErrInvalidInput.Error(), http.StatusBadRequest)
		return
	}

	exists, err := h.service.Exists(r.Context(), quote.Author, quote.Quote)
	if err != nil {
		if err == domain.ErrInvalidInput {
			h.logger.Error("Неверный ввод для проверки существования цитаты", zap.Error(err))
			sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		} else {
			h.logger.Error("Ошибка проверки существования цитаты", zap.Error(err))
			sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if exists {
		h.logger.Info("Цитата уже существует")
		sendErrorResponse(w, "Quote already exists", http.StatusBadRequest)
		return
	}

	if err := h.service.Create(r.Context(), &quote); err != nil {
		h.logger.Error("Ошибка создания цитаты", zap.Error(err))
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]interface{}{
		"data": quote,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Ошибка кодирования ответа", zap.Error(err))
	}
}

func (h *Handler) getAllQuotes(w http.ResponseWriter, r *http.Request) {
	author := r.URL.Query().Get("author")
	if author == "" {
		quotes, err := h.service.GetAll(r.Context())
		if err != nil {
			h.logger.Error("Ошибка получения цитат", zap.Error(err))
			sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"data": quotes,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			h.logger.Error("Ошибка кодирования ответа", zap.Error(err))
		}
		return
	}

	quotes, err := h.service.GetByAuthor(r.Context(), author)
	if err != nil {
		if err == domain.ErrInvalidInput {
			h.logger.Error("Неверный ввод для получения цитат по автору", zap.Error(err))
			sendErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.logger.Error("Ошибка получения цитат по автору", zap.Error(err))
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(quotes) == 0 {
		sendErrorResponse(w, "No quotes found for the specified author", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"data": quotes,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Ошибка кодирования ответа", zap.Error(err))
	}
}

func (h *Handler) getRandomQuote(w http.ResponseWriter, r *http.Request) {
	quote, err := h.service.GetRandom(r.Context())
	if err != nil {
		if err == domain.ErrNotFound {
			h.logger.Error("Цитаты не найдены", zap.Error(err))
			sendErrorResponse(w, "No quotes found", http.StatusNotFound)
			return
		}
		h.logger.Error("Ошибка получения случайной цитаты", zap.Error(err))
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"data": quote,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Ошибка кодирования ответа", zap.Error(err))
	}
}

func (h *Handler) deleteQuote(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Неверный формат ID", zap.Error(err))
		sendErrorResponse(w, domain.ErrInvalidInput.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if err == domain.ErrNotFound {
			h.logger.Error("Цитата не найдена", zap.Error(err))
			sendErrorResponse(w, "Quote not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Ошибка удаления цитаты", zap.Error(err))
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message": fmt.Sprintf("Quote with ID %d deleted successfully", id),
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Ошибка кодирования ответа", zap.Error(err))
	}
}

func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{
		"error": message,
	}
	json.NewEncoder(w).Encode(response)
}
