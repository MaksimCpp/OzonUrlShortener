package service

import (
	"context"
	"net/url"

	"github.com/MaksimCpp/OzonUrlShortener/internal/config"
	"github.com/MaksimCpp/OzonUrlShortener/internal/domain"
	"github.com/MaksimCpp/OzonUrlShortener/internal/utils"
)

const (
	maxLengthOriginalURL = 2000
)

type URLService struct {
	repo domain.URLRepository
	cfg  *config.Config
}

func NewURLService(
	repo domain.URLRepository,
	cfg  *config.Config) *URLService {
	return &URLService{
		repo: repo,
		cfg: cfg,
	}
}

func (s *URLService) CreateShortURL(
	ctx context.Context, originalURL string,
) (string, error) {
	if len(originalURL) > maxLengthOriginalURL {
		return "", domain.ErrInvalidURLLength
	}

	err := utils.CheckURL(originalURL)

	if err != nil {
		return "", err
	}

	shortCode, err := generateShortCode()

	if err != nil {
		return "", err
	}

	in := domain.URL{
		OriginalURL: originalURL,
		ShortCode: shortCode,
	}

	result, err := s.repo.Create(ctx, &in)

	if err != nil {
		return "", err
	}

	shortURL, err := url.JoinPath(s.cfg.BaseURL, result.ShortCode)

	if err != nil {
		return "", err
	}

	return shortURL, nil
}