package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/MaksimCpp/OzonUrlShortener/internal/domain"
)

type mockURLService struct {
	t              *testing.T
	createShortURL func(ctx context.Context, originalURL string) (string, error)
	getOriginalURL func(ctx context.Context, shortURL string) (string, error)
}

func (s mockURLService) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
	s.t.Helper()
	return s.createShortURL(ctx, originalURL)
}

func (s mockURLService) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	s.t.Helper()
	return s.getOriginalURL(ctx, shortURL)
}

func TestCreateShortURL_InvalidJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
	}{
		{name: "opening brace only", jsonData: "{"},
		{name: "closing brace only", jsonData: "}"},
		{name: "plain text", jsonData: "hello, ozon"},
		{name: "unquoted key and value", jsonData: "{url: https://ozon.ru/"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/api/v1/urls", strings.NewReader(test.jsonData))

			NewURLHandler(nil).CreateShortURL(w, r)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected code StatusBadRequest, got %d", w.Code)
			}
		})

	}
}

func TestCreateShortURL_InvalidURL(t *testing.T) {
	tests := []struct {
		name string
		body map[string]string
	}{
		{
			name: "ozon",
			body: map[string]string{"url": "hello, ozon"},
		},
		{
			name: "go dev",
			body: map[string]string{"url": "Go-разработка"},
		},
		{
			name: "github file",
			body: map[string]string{"url": "htps://github.com/MaksimCpp/OzonUrlShortener/blob/main/internal/repository/postgresql/url_repository.go"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			body, err := json.Marshal(test.body)

			if err != nil {
				t.Fatalf("expected nil err, got %v", err)
			}

			r := httptest.NewRequest("POST", "/api/v1/urls", strings.NewReader(string(body)))

			srv := mockURLService{
				t: t,
				createShortURL: func(ctx context.Context, originalURL string) (string, error) {
					if originalURL != test.body["url"] {
						t.Errorf("expected URL %q, got %q", test.body["url"], originalURL)
					}

					return "", domain.ErrInvalidURL
				},
			}

			NewURLHandler(srv).CreateShortURL(w, r)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected code StatusBadRequest, got %d", w.Code)
			}
		})
	}
}

func TestCreateShortURL_InternalError(t *testing.T) {
	fail := errors.New("unexpected error")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/urls", strings.NewReader(`{"url": "https://ozon.ru/"}`))

	srv := mockURLService{
		t: t,
		createShortURL: func(ctx context.Context, originalURL string) (string, error) {
			return "", fail
		},
	}

	NewURLHandler(srv).CreateShortURL(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected code StatusInternalServerError, got %d", w.Code)
	}
}

func TestCreateShortURL_Success(t *testing.T) {
	tests := []struct {
		name     string
		body     map[string]string
		response string
	}{
		{
			name:     "ozon",
			body:     map[string]string{"url": "https://ozon.ru"},
			response: "http://localhost:8080/abcABC123_",
		},
		{
			name:     "ozon fintech",
			body:     map[string]string{"url": "https://career.ozon.ru/fintech/vacancy/138892012"},
			response: "http://localhost:8080/abcAlkC123_",
		},
		{
			name:     "github file",
			body:     map[string]string{"url": "https://github.com/MaksimCpp/OzonUrlShortener/blob/main/internal/repository/postgresql/url_repository.go"},
			response: "http://localhost:8080/abcABC427_",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			body, err := json.Marshal(test.body)

			if err != nil {
				t.Fatalf("expected nil err, got %v", err)
			}

			r := httptest.NewRequest("POST", "/api/v1/urls", strings.NewReader(string(body)))

			srv := mockURLService{
				t: t,
				createShortURL: func(ctx context.Context, originalURL string) (string, error) {
					if originalURL != test.body["url"] {
						t.Errorf("expected URL %q, got %q", test.body["url"], originalURL)
					}

					return test.response, nil
				},
			}

			NewURLHandler(srv).CreateShortURL(w, r)

			if w.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("expected application/json, got %q", w.Header().Get("Content-Type"))
			}

			if w.Code != http.StatusCreated {
				t.Fatalf("expected code StatusCreated, got %d", w.Code)
			}

			var res struct {
				ShortURL string `json:"short_url"`
			}

			err = json.NewDecoder(w.Body).Decode(&res)

			if err != nil {
				t.Fatalf("expected nil err, got %v", err)
			}

			if res.ShortURL != test.response {
				t.Fatalf("expected short URL %v, got %v", test.response, res.ShortURL)
			}
		})
	}
}

func TestCreateShortURL_InvalidLengthURL(t *testing.T) {
	target := "/api/v1/urls"
	addr := "http://ozon.ru/" + strings.Repeat("a", 2001-len("http://ozon.ru/"))

	body, err := json.Marshal(map[string]string{"url": addr})

	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		"POST",
		target,
		strings.NewReader(string(body)),
	)
	srv := mockURLService{
		t: t,
		createShortURL: func(ctx context.Context, originalURL string) (string, error) {
			if originalURL != addr {
				t.Fatalf("expected URL %v, got %v", addr, originalURL)
			}

			return "", domain.ErrInvalidURLLength
		},
	}

	NewURLHandler(srv).CreateShortURL(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected code StatusBadRequest, got %d", w.Code)
	}
}

func TestCreateShortURL_InvalidShortCode(t *testing.T) {
	target := "/api/v1/urls"
	addr := "http://ozon.ru/"

	body, err := json.Marshal(map[string]string{"url": addr})

	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		"POST",
		target,
		strings.NewReader(string(body)),
	)
	srv := mockURLService{
		t: t,
		createShortURL: func(ctx context.Context, originalURL string) (string, error) {
			if originalURL != addr {
				t.Fatalf("expected URL %v, got %v", addr, originalURL)
			}

			return "", domain.ErrInvalidShortCode
		},
	}

	NewURLHandler(srv).CreateShortURL(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected code StatusBadRequest, got %d", w.Code)
	}
}

func TestGetOriginalURL_MissingShortURL(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/urls", nil)

	NewURLHandler(nil).GetOriginalURL(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected StatusBadRequest, got %v", w.Code)
	}
}

func TestGetOriginalURL_Success(t *testing.T) {
	tests := []struct {
		name     string
		param    map[string]string
		response string
	}{
		{
			name:     "ozon",
			param:    map[string]string{"short_url": "http://localhost:8080/abcABC123_"},
			response: "https://ozon.ru",
		},
		{
			name:     "ozon fintech",
			param:    map[string]string{"short_url": "http://localhost:8080/abcAlkC123_"},
			response: "https://career.ozon.ru/fintech/vacancy/138892012_",
		},
		{
			name:     "github file",
			param:    map[string]string{"short_url": "http://localhost:8080/abcABC427_"},
			response: "https://github.com/MaksimCpp/OzonUrlShortener/blob/main/internal/repository/postgresql/url_repository.go",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r := httptest.NewRequest(
				"GET",
				"/api/v1/urls?short_url="+url.QueryEscape(test.param["short_url"]),
				nil,
			)

			srv := mockURLService{
				t: t,
				getOriginalURL: func(ctx context.Context, shortURL string) (string, error) {
					if shortURL != test.param["short_url"] {
						t.Errorf("expected short URL %q, got %q", test.param["short_url"], shortURL)
					}

					return test.response, nil
				},
			}

			NewURLHandler(srv).GetOriginalURL(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("expected code StatusOK, got %d", w.Code)
			}

			if w.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("expected application/json, got %q", w.Header().Get("Content-Type"))
			}

			var res struct {
				OriginalURL string `json:"url"`
			}

			err := json.NewDecoder(w.Body).Decode(&res)

			if err != nil {
				t.Fatalf("expected nil err, got %v", err)
			}

			if res.OriginalURL != test.response {
				t.Fatalf("expected original URL %v, got %v", test.response, res.OriginalURL)
			}
		})
	}
}

func TestGetOriginalURL_ErrorNotFound(t *testing.T) {
	short := "http://localhost:8080/abcABC123_"
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		"GET",
		"/api/v1/urls?short_url="+url.QueryEscape(short),
		nil,
	)

	srv := mockURLService{
		t: t,
		getOriginalURL: func(ctx context.Context, shortURL string) (string, error) {
			if shortURL != short {
				t.Fatalf("expected short URL %v, got %v", short, shortURL)
			}

			return "", domain.ErrURLNotFound
		},
	}

	NewURLHandler(srv).GetOriginalURL(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected StatusNotFound, got %v", w.Code)
	}
}

func TestGetOriginalURL_InvalidURL(t *testing.T) {
	tests := []struct {
		name   string
		param  map[string]string
		srvErr error
	}{
		{
			name:   "test 1",
			param:  map[string]string{"short_url": "http://localhost:8080/abcABC123_"},
			srvErr: domain.ErrInvalidURL,
		},
		{
			name:   "test 2",
			param:  map[string]string{"short_url": "http://localhost:8080/abcAlkC123_"},
			srvErr: domain.ErrInvalidURL,
		},
		{
			name:   "test 3",
			param:  map[string]string{"short_url": "http;http://localhost:8080/abcABC427_"},
			srvErr: domain.ErrInvalidShortCode,
		},
		{
			name:   "test 4",
			param:  map[string]string{"short_url": "hi:://ozon"},
			srvErr: domain.ErrInvalidShortCode,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r := httptest.NewRequest(
				"GET",
				"/api/v1/urls?short_url="+url.QueryEscape(test.param["short_url"]),
				nil,
			)

			srv := mockURLService{
				t: t,
				getOriginalURL: func(ctx context.Context, shortURL string) (string, error) {
					if shortURL != test.param["short_url"] {
						t.Errorf("expected short URL %q, got %q", test.param["short_url"], shortURL)
					}

					return "", test.srvErr
				},
			}

			NewURLHandler(srv).GetOriginalURL(w, r)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected code StatusBadRequest, got %d", w.Code)
			}
		})
	}
}

func TestGetOriginalURL_InternalError(t *testing.T) {
	fail := errors.New("unexpected error")
	short := "http://localhost:8080/abcABC123_"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		"GET",
		"/api/v1/urls?short_url="+url.QueryEscape(short),
		nil,
	)
	srv := mockURLService{
		t: t,
		getOriginalURL: func(ctx context.Context, shortURL string) (string, error) {
			return "", fail
		},
	}

	NewURLHandler(srv).GetOriginalURL(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected code StatusInternalServerError, got %d", w.Code)
	}
}
