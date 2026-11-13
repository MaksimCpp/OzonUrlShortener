package inmemory

import (
	"context"
	"sync"

	"github.com/MaksimCpp/OzonUrlShortener/internal/domain"
)

type URLRepository struct {
	mu            sync.Mutex
	lastID        int64
	byOriginalURL map[string]domain.URL
	byShortCode   map[string]domain.URL
}

func NewURLRepository() *URLRepository {
	return &URLRepository{
		byOriginalURL: make(map[string]domain.URL),
		byShortCode:   make(map[string]domain.URL),
	}
}

func (repo *URLRepository) Create(
	ctx context.Context, value *domain.URL,
) (*domain.URL, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(value.OriginalURL) == 0 || len(value.OriginalURL) > 2000 {
		return nil, domain.ErrInvalidURLLength
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if _, exists := repo.byOriginalURL[value.OriginalURL]; exists {
		return nil, domain.ErrURLAlreadyExist
	}

	if _, exists := repo.byShortCode[value.ShortCode]; exists {
		return nil, domain.ErrShortCodeAlreadyExists
	}

	repo.lastID++

	result := domain.URL{
		ID:          repo.lastID,
		OriginalURL: value.OriginalURL,
		ShortCode:   value.ShortCode,
	}

	repo.byOriginalURL[result.OriginalURL] = result
	repo.byShortCode[result.ShortCode] = result

	return &result, nil
}

func (repo *URLRepository) GetByShortCode(
	ctx context.Context, shortCode string,
) (*domain.URL, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result, exists := repo.byShortCode[shortCode]

	if !exists {
		return nil, domain.ErrURLNotFound
	}

	return &result, nil
}

func (repo *URLRepository) GetByOriginalURL(
	ctx context.Context, originalURL string,
) (*domain.URL, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result, exists := repo.byOriginalURL[originalURL]

	if !exists {
		return nil, domain.ErrURLNotFound
	}

	return &result, nil
}
