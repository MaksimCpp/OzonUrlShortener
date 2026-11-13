package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/MaksimCpp/OzonUrlShortener/internal/domain"
)

type repositoryStub struct {
	t          *testing.T
	create     func(context.Context, *domain.URL) (*domain.URL, error)
	byOriginal func(context.Context, string) (*domain.URL, error)
	byCode     func(context.Context, string) (*domain.URL, error)
}

func (r repositoryStub) Create(ctx context.Context, u *domain.URL) (*domain.URL, error) {
	r.t.Helper()

	if r.create == nil {
		r.t.Fatal("unexpected Create call")
	}

	return r.create(ctx, u)
}

func (r repositoryStub) GetByOriginalURL(ctx context.Context, u string) (*domain.URL, error) {
	r.t.Helper()

	if r.byOriginal == nil {
		r.t.Fatal("unexpected GetByOriginalURL call")
	}

	return r.byOriginal(ctx, u)
}

func (r repositoryStub) GetByShortCode(ctx context.Context, code string) (*domain.URL, error) {
	r.t.Helper()

	if r.byCode == nil {
		r.t.Fatal("unexpected GetByShortCode call")
	}

	return r.byCode(ctx, code)
}

func TestCreateShortURLSavesNewURL(t *testing.T) {
	const original = "https://ozon.ru/path?a=1&b=2#section"
	var saved domain.URL
	calls := 0
	repo := repositoryStub{
		t: t,
		create: func(_ context.Context, value *domain.URL) (*domain.URL, error) {
			calls++
			saved = *value
			return value, nil
		},
	}

	result, err := NewURLService(repo, "https://short.ozon/prefix/").CreateShortURL(context.Background(), original)

	if err != nil {
		t.Fatal(err)
	}

	if saved.OriginalURL != original || saved.ShortCode == "" || calls != 1 {
		t.Fatalf("saved = %+v, calls = %d", saved, calls)
	}

	if result != "https://short.ozon/prefix/"+saved.ShortCode {
		t.Fatalf("short URL = %q", result)
	}
}

func TestCreateShortURLRetriesCollisions(t *testing.T) {
	calls := 0
	var savedCode string
	repo := repositoryStub{
		t: t,
		create: func(_ context.Context, value *domain.URL) (*domain.URL, error) {
			calls++
			if calls < 5 {
				return nil, fmt.Errorf("insert: %w", domain.ErrShortCodeAlreadyExists)
			}
			savedCode = value.ShortCode
			return value, nil
		},
	}

	result, err := NewURLService(repo, "https://short.ozon").CreateShortURL(context.Background(), "https://ozon.ru")

	if err != nil {
		t.Fatal(err)
	}

	if calls != 5 {
		t.Fatalf("attempts = %d; want 5", calls)
	}

	if result != "https://short.ozon/"+savedCode {
		t.Fatalf("short URL = %q", result)
	}
}

func TestCreateShortURLStopsAfterFiveCollisions(t *testing.T) {
	calls := 0
	repo := repositoryStub{
		t: t,
		create: func(context.Context, *domain.URL) (*domain.URL, error) {
			calls++
			return nil, domain.ErrShortCodeAlreadyExists
		},
	}

	result, err := NewURLService(repo, "https://short.ozon").CreateShortURL(context.Background(), "https://ozon.ru")

	if result != "" || !errors.Is(err, domain.ErrShortCodeAlreadyExists) {
		t.Fatalf("result = %q, error = %v", result, err)
	}

	if calls != 5 {
		t.Fatalf("attempts = %d; want 5", calls)
	}
}

func TestCreateShortURLReturnsExistingURL(t *testing.T) {
	const original = "https://ozon.ru/path?a=1&b=2#section"
	calls, lookups := 0, 0
	repo := repositoryStub{
		t: t,
		create: func(_ context.Context, value *domain.URL) (*domain.URL, error) {
			calls++
			if value.OriginalURL != original {
				t.Fatal("original URL changed")
			}
			return nil, fmt.Errorf("insert: %w", domain.ErrURLAlreadyExist)
		},
		byOriginal: func(_ context.Context, value string) (*domain.URL, error) {
			lookups++
			if value != original {
				t.Fatalf("lookup URL = %q", value)
			}
			return &domain.URL{ShortCode: "Existing_1"}, nil
		},
	}

	result, err := NewURLService(repo, "https://short.ozon").CreateShortURL(context.Background(), original)

	if err != nil {
		t.Fatal(err)
	}
	if result != "https://short.ozon/Existing_1" {
		t.Fatalf("short URL = %q", result)
	}
	if calls != 1 || lookups != 1 {
		t.Fatalf("attempts = %d, lookups = %d", calls, lookups)
	}
}

func TestCreateShortURLReturnsStorageError(t *testing.T) {
	failure := errors.New("storage unavailable")
	repo := repositoryStub{
		t: t,
		create: func(context.Context, *domain.URL) (*domain.URL, error) {
			return nil, failure
		},
	}

	result, err := NewURLService(repo, "https://short.ozon").CreateShortURL(context.Background(), "https://ozon.ru")

	if result != "" || !errors.Is(err, failure) {
		t.Fatalf("result = %q, error = %v", result, err)
	}
}

func TestCreateShortURLReturnsExistingLookupError(t *testing.T) {
	failure := errors.New("lookup unavailable")
	repo := repositoryStub{
		t: t,
		create: func(context.Context, *domain.URL) (*domain.URL, error) {
			return nil, domain.ErrURLAlreadyExist
		},
		byOriginal: func(context.Context, string) (*domain.URL, error) {
			return nil, failure
		},
	}

	result, err := NewURLService(repo, "https://short.ozon").CreateShortURL(context.Background(), "https://ozon.ru")

	if result != "" || !errors.Is(err, failure) {
		t.Fatalf("result = %q, error = %v", result, err)
	}
}

func TestGetOriginalURLExtractsCode(t *testing.T) {
	const original = "https://ozon.ru?a=1&b=2#part"
	calls := 0
	repo := repositoryStub{
		t: t,
		byCode: func(_ context.Context, code string) (*domain.URL, error) {
			calls++
			if code != "Ab3_dE7fG9" {
				t.Fatalf("code = %q", code)
			}
			return &domain.URL{OriginalURL: original}, nil
		},
	}

	result, err := NewURLService(repo, "https://short.ozon").GetOriginalURL(context.Background(), "https://short.ozon/Ab3_dE7fG9")

	if err != nil {
		t.Fatal(err)
	}
	if result != original || calls != 1 {
		t.Fatalf("result = %q, lookups = %d", result, calls)
	}
}

func TestGetOriginalURLRejectsInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		wantErr     error
	}{
		{"invalid URL", "hello", domain.ErrInvalidURL},
		{"empty path", "https://short.ozon", domain.ErrInvalidShortCode},
		{"nine characters", "https://short.ozon/Ab3_dE7fG", domain.ErrInvalidShortCode},
		{"eleven characters", "https://short.ozon/Ab3_dE7fG90", domain.ErrInvalidShortCode},
		{"multibyte ten bytes", "https://short.ozon/" + strings.Repeat("я", 5), domain.ErrInvalidShortCode},
		{"invalid alphabet", "https://short.ozon/Ab3-dE7fG9", domain.ErrInvalidShortCode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := NewURLService(repositoryStub{t: t}, "https://short.ozon").GetOriginalURL(context.Background(), tc.input)

			if result != "" || !errors.Is(err, tc.wantErr) {
				t.Fatalf("result = %q, error = %v; want %v", result, err, tc.wantErr)
			}
		})
	}
}

func TestGetOriginalURLReturnsStorageError(t *testing.T) {
	for _, failure := range []error{domain.ErrURLNotFound, errors.New("storage unavailable")} {
		t.Run(failure.Error(), func(t *testing.T) {
			repo := repositoryStub{
				t: t,
				byCode: func(context.Context, string) (*domain.URL, error) {
					return nil, failure
				},
			}

			result, err := NewURLService(repo, "https://short.ozon").GetOriginalURL(context.Background(), "https://short.ozon/Ab3_dE7fG9")

			if result != "" || !errors.Is(err, failure) {
				t.Fatalf("result = %q, error = %v", result, err)
			}
		})
	}
}

func TestCreateShortURLValidation(t *testing.T) {
	const prefix = "https://ozon.ru/"
	for _, tc := range []struct {
		name, value string
		wantErr     error
		wantCalls   int
	}{
		{"empty", "", domain.ErrInvalidURL, 0},
		{"relative", "hello", domain.ErrInvalidURL, 0},
		{"2000 bytes", prefix + strings.Repeat("a", 2000-len(prefix)), nil, 1},
		{"2001 bytes", prefix + strings.Repeat("a", 2001-len(prefix)), domain.ErrInvalidURLLength, 0},
		{"unicode 2000 bytes", prefix + strings.Repeat("я", (2000-len(prefix))/len("я")), nil, 1},
		{"unicode 2001 bytes", prefix + strings.Repeat("я", (2000-len(prefix))/len("я")) + "a", domain.ErrInvalidURLLength, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			repo := repositoryStub{
				t: t,
				create: func(_ context.Context, value *domain.URL) (*domain.URL, error) {
					calls++
					if value.OriginalURL != tc.value {
						t.Fatal("original URL changed")
					}
					return value, nil
				},
			}

			result, err := NewURLService(repo, "https://short.ozon").CreateShortURL(context.Background(), tc.value)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v; want %v", err, tc.wantErr)
			}
			if calls != tc.wantCalls {
				t.Fatalf("insertions = %d; want %d", calls, tc.wantCalls)
			}
			if tc.wantErr != nil && result != "" {
				t.Fatalf("result on error = %q", result)
			}
			if tc.wantErr == nil && result == "" {
				t.Fatal("valid URL returned an empty result")
			}
		})
	}
}

func TestCancellationStopsCollisionRetries(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	repo := repositoryStub{t: t, create: func(got context.Context, _ *domain.URL) (*domain.URL, error) {
		calls++
		if got != ctx {
			t.Fatal("context changed")
		}
		cancel()
		return nil, domain.ErrShortCodeAlreadyExists
	}}
	result, err := NewURLService(repo, "https://short.ozon").CreateShortURL(ctx, "https://ozon.ru")
	if result != "" || !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("result = %q, error = %v, calls = %d", result, err, calls)
	}
}

func TestGenerateShortCodeFormat(t *testing.T) {
	for i := 0; i < 100; i++ {
		code, err := generateShortCode()
		if err != nil {
			t.Fatal(err)
		}
		if len(code) != 10 {
			t.Fatalf("length = %d", len(code))
		}
		for _, c := range code {
			if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_", c) {
				t.Fatalf("invalid code %q", code)
			}
		}
	}
}
