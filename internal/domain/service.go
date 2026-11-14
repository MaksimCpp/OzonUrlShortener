package domain

import "context"

type URLServiceInterface interface {
	CreateShortURL(ctx context.Context, originalURL string) (string, error)
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
}