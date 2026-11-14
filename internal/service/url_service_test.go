package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/MaksimCpp/OzonUrlShortener/internal/domain"
)

type mockURLRepository struct {
	t                *testing.T
	create           func(ctx context.Context, url *domain.URL) (*domain.URL, error)
	getByShortCode   func(ctx context.Context, shortCode string) (*domain.URL, error)
	getByOriginalURL func(ctx context.Context, originalURL string) (*domain.URL, error)
}

func (r mockURLRepository) Create(ctx context.Context, url *domain.URL) (*domain.URL, error) {
	r.t.Helper()
	return r.create(ctx, url)
}

func (r mockURLRepository) GetByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	r.t.Helper()
	return r.getByShortCode(ctx, shortCode)
}

func (r mockURLRepository) GetByOriginalURL(ctx context.Context, originalURL string) (*domain.URL, error) {
	r.t.Helper()
	return r.getByOriginalURL(ctx, originalURL)
}

func TestCreateShortURL_TooLongURL(t *testing.T) {
	baseURL := "http://localhost:8080/"
	originalURL := "https://ozon.ru/" + strings.Repeat("a", 2001 - len("https://ozon.ru/"))

	result, err := NewURLService(
		nil, 
		baseURL,
	).CreateShortURL(context.Background(), originalURL)

	if !errors.Is(err, domain.ErrInvalidURLLength) {
		t.Fatalf("expected domain.ErrInvalidURLLength, got %v", err)
	}

	if result != "" {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestCreateShortURL_EmptyURL(t *testing.T) {
	baseURL := "http://localhost:8080/"

	result, err := NewURLService(
		nil, 
		baseURL,
	).CreateShortURL(context.Background(), "")

	if !errors.Is(err, domain.ErrInvalidURL) {
		t.Fatalf("expected domain.ErrInvalidURL, got %v", err)
	}

	if result != "" {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestCreateShortURL_CanceledCtx(t *testing.T) {
	baseURL := "http://localhost:8080/"
	originalURL := "https://ozon.ru/"

	ctx, cancel := context.WithCancel(context.Background())

	cancel()

	result, err := NewURLService(
		nil, 
		baseURL,
	).CreateShortURL(ctx, originalURL)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if result != "" {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestCreateShortURL_Success(t *testing.T) {
	baseURL := "http://localhost:8080/"
	originalURL := "https://ozon.ru/"
	shortCode := "abcABC123_"

	repo := mockURLRepository{
		t: t,
		create: func(_ context.Context, url *domain.URL) (*domain.URL, error) {
			if url.OriginalURL != originalURL {
				t.Errorf("expected original URL %q, got %q", originalURL, url.OriginalURL)
			}

			if len(url.ShortCode) != 10 {
				t.Errorf("expected code length 10, got %d", len(url.ShortCode))
			}

			return &domain.URL{
				OriginalURL: originalURL,
				ShortCode: shortCode,
			}, nil
		},
	}

	result, err := NewURLService(
		repo, 
		baseURL,
	).CreateShortURL(context.Background(), originalURL)


	wantResult := baseURL + shortCode

	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	if result != wantResult {
		t.Fatalf("expected result = %v, got %v", wantResult, result)
	}
}

func TestCreateShortURL_RepoError(t *testing.T) {
	baseURL := "http://localhost:8080/"
	originalURL := "https://ozon.ru/"
	repoErr := errors.New("Repo error")

	repo := mockURLRepository{
		t: t,
		create: func(_ context.Context, url *domain.URL) (*domain.URL, error) {
			return nil, repoErr
		},
	}

	result, err := NewURLService(
		repo, 
		baseURL,
	).CreateShortURL(context.Background(), originalURL)

	if !errors.Is(err, repoErr) {
		t.Fatalf("expected err %v, got %v", repoErr, err)
	}

	if result != "" {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestCreateShortURL_URLAlreadyExist(t *testing.T) {
	baseURL := "http://localhost:8080/"
	originalURL := "https://ozon.ru/"
	shortCode := "abcABC123_"

	repo := mockURLRepository{
		t: t,
		create: func(_ context.Context, url *domain.URL) (*domain.URL, error) {
			return nil, domain.ErrURLAlreadyExist
		},

		getByOriginalURL: func(_ context.Context, gotURL string) (*domain.URL, error) {
			if gotURL != originalURL {
				t.Errorf("expected original URL %q, got %q", originalURL, gotURL)
			}

			return &domain.URL{
				OriginalURL: originalURL,
				ShortCode: shortCode,
			}, nil
		},
	}

	result, err := NewURLService(
		repo, 
		baseURL,
	).CreateShortURL(context.Background(), originalURL)

	wantResult := baseURL + shortCode

	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	if result != wantResult {
		t.Fatalf("excepted result = %v, got %v", wantResult, result)
	}
}

func TestCreateShortURL_ErrorGetURL(t *testing.T) {
	baseURL := "http://localhost:8080/"
	originalURL := "https://ozon.ru/"

	repo := mockURLRepository{
		t: t,
		create: func(_ context.Context, url *domain.URL) (*domain.URL, error) {
			return nil, domain.ErrURLAlreadyExist
		},

		getByOriginalURL: func(_ context.Context, gotURL string) (*domain.URL, error) {
			return nil, domain.ErrURLNotFound
		},
	}

	result, err := NewURLService(
		repo, 
		baseURL,
	).CreateShortURL(context.Background(), originalURL)

	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Fatalf("expected domain.ErrURLNotFound, got %v", err)
	}

	if result != "" {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestCreateShortURL_Collision(t *testing.T) {
	baseURL := "http://localhost:8080/"
	originalURL := "https://ozon.ru/"
	shortCode := "abcABC123_"
	calls := 0

	repo := mockURLRepository{
		t: t,
		create: func(_ context.Context, url *domain.URL) (*domain.URL, error) {
			calls++

			if calls < 2 {
				return nil, domain.ErrShortCodeAlreadyExists
			}

			return &domain.URL{
				OriginalURL: originalURL,
				ShortCode: shortCode,
			}, nil
		},
	}

	result, err := NewURLService(
		repo, 
		baseURL,
	).CreateShortURL(context.Background(), originalURL)

	if calls != 2 {
		t.Errorf("expected 2 Create calls, got %d", calls)
	}

	wantResult := baseURL + shortCode

	if err != nil {
		t.Fatalf("expected err nil, got %v", err)
	}

	if result != wantResult {
		t.Fatalf("expected result = %v, got %v", wantResult, result)
	}
}

func TestCreateShortURL_CollisionAttemptsExhausted(t *testing.T) {
	baseURL := "http://localhost:8080/"
	originalURL := "https://ozon.ru/"
	calls := 0

	repo := mockURLRepository{
		t: t,
		create: func(_ context.Context, url *domain.URL) (*domain.URL, error) {
			calls++
			return nil, domain.ErrShortCodeAlreadyExists
		},
	}

	result, err := NewURLService(
		repo, 
		baseURL,
	).CreateShortURL(context.Background(), originalURL)

	if calls != 5 {
		t.Errorf("expected 5 Create calls, got %d", calls)
	}

	if !errors.Is(err, domain.ErrShortCodeAlreadyExists) {
		t.Fatalf("expected err domain.ErrShortCodeAlreadyExists, got %v", err)
	}

	if result != "" {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestCreateShortURL_InvalidURL(t *testing.T) {
	baseURL := "http://localhost:8080/"
	originalURL := "hello, ozon"

	result, err := NewURLService(
		nil, 
		baseURL,
	).CreateShortURL(context.Background(), originalURL)

	if !errors.Is(err, domain.ErrInvalidURL) {
		t.Fatalf("expected domain.ErrInvalidURL, got %v", err)
	}

	if result != "" {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestGetOriginalURL_Success(t *testing.T) {
	baseURL := "http://localhost:8080/"
	originalURL := "https://ozon.ru/"
	shortCode := "abcABC123_"

	repo := mockURLRepository{
		t: t,
		getByShortCode: func(_ context.Context, short string) (*domain.URL, error) {
			if short != shortCode {
				t.Errorf("expected shortCode %v, got %v", shortCode, short)
			}

			return &domain.URL{
				OriginalURL: originalURL,
				ShortCode: shortCode,
			}, nil
		},
	}

	shortURL := baseURL + shortCode

	result, err := NewURLService(
		repo,
		baseURL,
	).GetOriginalURL(context.Background(), shortURL)

	if err != nil {
		t.Fatalf("expected err nil, got %v", err)
	}

	if result != originalURL {
		t.Fatalf("excepted result = %v, got %v", originalURL, result)
	}
}

func TestGetOriginalURL_InvalidURL(t *testing.T) {
	baseURL := "http://localhost:8080/"
	shortURL := "hello, ozon"

	result, err := NewURLService(
		nil,
		baseURL,
	).GetOriginalURL(context.Background(), shortURL)

	if !errors.Is(err, domain.ErrInvalidURL) {
		t.Fatalf("expected domain.ErrInvalidURL, got %v", err)
	}

	if result != "" {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestGetOriginalURL_InvalidShortCodeLength(t *testing.T) {
	baseURL := "http://localhost:8080/"

	tests := []struct {
		name      string
		shortCode string
	}{
		{name: "empty code", shortCode: ""},
		{name: "too short code", shortCode: "abcABC123"},
		{name: "too long code", shortCode: "abcABC123__"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			shortURL := baseURL + test.shortCode

			result, err := NewURLService(
				nil,
				baseURL,
			).GetOriginalURL(context.Background(), shortURL)

			if !errors.Is(err, domain.ErrInvalidShortCode) {
				t.Fatalf("expected domain.ErrInvalidShortCode, got %v", err)
			}

			if result != "" {
				t.Fatalf("expected empty result, got %v", result)
			}
		})
	}
}

func TestGetOriginalURL_InvalidShortCodeCharacter(t *testing.T) {
	baseURL := "http://localhost:8080/"
	shortCode := "abcABC123-"
	shortURL := baseURL + shortCode

	result, err := NewURLService(
		nil,
		baseURL,
	).GetOriginalURL(context.Background(), shortURL)

	if !errors.Is(err, domain.ErrInvalidShortCode) {
		t.Fatalf("expected domain.ErrInvalidShortCode, got %v", err)
	}

	if result != "" {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestGetOriginalURL_RepoError(t *testing.T) {
	baseURL := "http://localhost:8080/"
	shortCode := "abcABC123_"

	tests := []struct {
		name    string
		repoErr error
	}{
		{name: "URL not found", repoErr: domain.ErrURLNotFound},
		{name: "repo error", repoErr: errors.New("Repo error")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := mockURLRepository{
				t: t,
				getByShortCode: func(_ context.Context, short string) (*domain.URL, error) {
					if short != shortCode {
						t.Errorf("expected shortCode %v, got %v", shortCode, short)
					}

					return nil, test.repoErr
				},
			}

			shortURL := baseURL + shortCode

			result, err := NewURLService(
				repo,
				baseURL,
			).GetOriginalURL(context.Background(), shortURL)

			if !errors.Is(err, test.repoErr) {
				t.Fatalf("expected err %v, got %v", test.repoErr, err)
			}

			if result != "" {
				t.Fatalf("expected empty result, got %v", result)
			}
		})
	}
}
