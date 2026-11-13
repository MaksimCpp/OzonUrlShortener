package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/MaksimCpp/OzonUrlShortener/internal/domain"
	"github.com/MaksimCpp/OzonUrlShortener/internal/utils"
)

const (
	maxLengthOriginalURL = 2000
	maxCreateAttempts    = 5
)

type URLService struct {
	repo    domain.URLRepository
	baseURL string
}

func NewURLService(
	repo domain.URLRepository,
	baseURL string,
) *URLService {
	return &URLService{
		repo:    repo,
		baseURL: baseURL,
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

	for attempt := 0; attempt < maxCreateAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
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

		switch {
		case errors.Is(err, domain.ErrShortCodeAlreadyExists):
			continue

		case errors.Is(err, domain.ErrURLAlreadyExist):
			result, err = s.repo.GetByOriginalURL(ctx, originalURL)
		}

		if err != nil {
			return "", err
		}
		
		return url.JoinPath(s.baseURL, result.ShortCode)
	}

	return "", fmt.Errorf("create short URL after %d attempts: %w", maxCreateAttempts, domain.ErrShortCodeAlreadyExists)
}

func (s *URLService) GetOriginalURL(
	ctx context.Context, shortURL string,
) (string, error) {
	err := utils.CheckURL(shortURL)

	if err != nil {
		return "", err
	}

	parsed, err := url.Parse(shortURL)

	if err != nil {
		return "", err
	}

	index := strings.LastIndexByte(parsed.Path, '/')
	if index < 0 {
		return "", domain.ErrInvalidShortCode
	}

	shortCode := parsed.Path[index + 1:]
	if len(shortCode) != shortCodeLength {
		return "", domain.ErrInvalidShortCode
	}

	for i := range shortCode {
		if !strings.ContainsRune(shortCodeAlphabet, rune(shortCode[i])) {
			return "", domain.ErrInvalidShortCode
		}
	}

	result, err := s.repo.GetByShortCode(ctx, shortCode)

	if err != nil {
		return "", err
	}

	return result.OriginalURL, nil
}
