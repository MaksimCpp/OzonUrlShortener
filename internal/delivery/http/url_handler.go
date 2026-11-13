package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/MaksimCpp/OzonUrlShortener/internal/domain"
	"github.com/MaksimCpp/OzonUrlShortener/internal/service"
)

type URLHandler struct {
	srv *service.URLService
}

func NewURLHandler(srv *service.URLService) *URLHandler {
	return &URLHandler{
		srv: srv,
	}
}

type OriginalURLRequest struct {
	OriginalURL string `json:"url"`
}

type OriginalURLResponse struct {
	OriginalURL string `json:"url"`
}

type ShortURLResponse struct {
	ShortURL string `json:"short_url"`
}

func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5 * time.Second)
	defer cancel()

	var req OriginalURLRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.srv.CreateShortURL(ctx, req.OriginalURL)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidShortCode) ||
			errors.Is(err, domain.ErrInvalidURLLength) ||
			errors.Is(err, domain.ErrInvalidURL):
			http.Error(w, err.Error(), http.StatusBadRequest)
			return

		default:
			log.Printf("create short URL: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(
		ShortURLResponse{
			ShortURL: result,
		},
	)

	if err != nil {
		log.Printf("write create short URL response: %v", err)
		return
	}
}

func (h *URLHandler) GetOriginalURL(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5 * time.Second)
	defer cancel()

	shortURL := r.URL.Query().Get("short_url")

	if shortURL == "" {
		http.Error(w, "short_url is required", http.StatusNotFound)
		return
	}

	result, err := h.srv.GetOriginalURL(ctx, shortURL)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrURLNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
			return

		case errors.Is(err, domain.ErrInvalidShortCode) ||
			errors.Is(err, domain.ErrInvalidURL):
			http.Error(w, err.Error(), http.StatusBadRequest)
			return

		default:
			log.Printf("get original URL: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(
		OriginalURLResponse{
			OriginalURL: result,
		},
	)

	if err != nil {
		log.Printf("write get original URL response: %v", err)
		return
	}
}

func (h *URLHandler) Route(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/urls", h.CreateShortURL)
	mux.HandleFunc("GET /api/v1/urls", h.GetOriginalURL)
}
