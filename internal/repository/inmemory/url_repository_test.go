package inmemory

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/MaksimCpp/OzonUrlShortener/internal/domain"
)

func TestCreate_Success(t *testing.T) {
	repo := NewURLRepository()
	value := domain.URL{
		OriginalURL: "https://ozon.ru/",
		ShortCode:   "abcABC123_",
	}

	result, err := repo.Create(context.Background(), &value)

	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	wantResult := domain.URL{
		ID:          1,
		OriginalURL: value.OriginalURL,
		ShortCode:   value.ShortCode,
	}

	if result == nil {
		t.Fatal("expected URL, got nil")
	}

	if *result != wantResult {
		t.Fatalf("expected result %v, got %v", wantResult, *result)
	}
}

func TestCreate_InvalidURLLength(t *testing.T) {
	tests := []struct {
		name        string
		originalURL string
	}{
		{name: "empty URL", originalURL: ""},
		{name: "too long URL", originalURL: "https://ozon.ru/" + strings.Repeat("a", 2001-len("https://ozon.ru/"))},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := NewURLRepository()
			value := domain.URL{
				OriginalURL: test.originalURL,
				ShortCode:   "abcABC123_",
			}

			result, err := repo.Create(context.Background(), &value)

			if !errors.Is(err, domain.ErrInvalidURLLength) {
				t.Fatalf("expected domain.ErrInvalidURLLength, got %v", err)
			}

			if result != nil {
				t.Fatalf("expected nil result, got %v", result)
			}
		})
	}
}

func TestCreate_AlreadyExists(t *testing.T) {
	tests := []struct {
		name    string
		value   domain.URL
		repoErr error
	}{
		{
			name:    "original URL already exists",
			value:   domain.URL{OriginalURL: "https://ozon.ru/", ShortCode: "xyzXYZ456_"},
			repoErr: domain.ErrURLAlreadyExist,
		},
		{
			name:    "short code already exists",
			value:   domain.URL{OriginalURL: "https://github.com/", ShortCode: "abcABC123_"},
			repoErr: domain.ErrShortCodeAlreadyExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := NewURLRepository()
			value := domain.URL{
				OriginalURL: "https://ozon.ru/",
				ShortCode:   "abcABC123_",
			}

			_, err := repo.Create(context.Background(), &value)
			if err != nil {
				t.Fatalf("expected nil err, got %v", err)
			}

			result, err := repo.Create(context.Background(), &test.value)

			if !errors.Is(err, test.repoErr) {
				t.Fatalf("expected err %v, got %v", test.repoErr, err)
			}

			if result != nil {
				t.Fatalf("expected nil result, got %v", result)
			}
		})
	}
}

func TestGetByShortCode_Success(t *testing.T) {
	repo := NewURLRepository()
	value := domain.URL{
		OriginalURL: "https://ozon.ru/",
		ShortCode:   "abcABC123_",
	}

	created, err := repo.Create(context.Background(), &value)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	result, err := repo.GetByShortCode(context.Background(), value.ShortCode)

	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	if result == nil {
		t.Fatal("expected URL, got nil")
	}

	if *result != *created {
		t.Fatalf("expected result %v, got %v", *created, *result)
	}
}

func TestGetByShortCode_NotFound(t *testing.T) {
	repo := NewURLRepository()

	result, err := repo.GetByShortCode(context.Background(), "abcABC123_")

	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Fatalf("expected domain.ErrURLNotFound, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestGetByOriginalURL_Success(t *testing.T) {
	repo := NewURLRepository()
	value := domain.URL{
		OriginalURL: "https://ozon.ru/",
		ShortCode:   "abcABC123_",
	}

	created, err := repo.Create(context.Background(), &value)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	result, err := repo.GetByOriginalURL(context.Background(), value.OriginalURL)

	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	if result == nil {
		t.Fatal("expected URL, got nil")
	}

	if *result != *created {
		t.Fatalf("expected result %v, got %v", *created, *result)
	}
}

func TestGetByOriginalURL_NotFound(t *testing.T) {
	repo := NewURLRepository()

	result, err := repo.GetByOriginalURL(context.Background(), "https://ozon.ru/")

	if !errors.Is(err, domain.ErrURLNotFound) {
		t.Fatalf("expected domain.ErrURLNotFound, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}
