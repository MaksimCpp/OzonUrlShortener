package domain

import "context"

type URLRepository interface {
	Create(ctx context.Context, url *URL) (*URL, error)
	GetByShortCode(ctx context.Context, shortCode string) (*URL, error)
}