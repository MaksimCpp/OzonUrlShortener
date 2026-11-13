package httpdelivery_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	httpdelivery "github.com/MaksimCpp/OzonUrlShortener/internal/delivery/http"
	"github.com/MaksimCpp/OzonUrlShortener/internal/domain"
	"github.com/MaksimCpp/OzonUrlShortener/internal/repository/inmemory"
	"github.com/MaksimCpp/OzonUrlShortener/internal/service"
)

func newMux(repo domain.URLRepository) *http.ServeMux {
	mux := http.NewServeMux()
	httpdelivery.NewURLHandler(service.NewURLService(repo, "https://short.ozon")).Route(mux)
	return mux
}

func request(mux *http.ServeMux, method, target, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(method, target, strings.NewReader(body)))
	return w
}

func TestURLRoundTrip(t *testing.T) {
	mux := newMux(inmemory.NewURLRepository())
	const original = "https://ozon.ru/path?a=1&b=two+words#part"

	short := createShortURL(t, mux, original)
	repeated := createShortURL(t, mux, original)

	if repeated != short {
		t.Fatalf("duplicate URL changed: %q != %q", repeated, short)
	}
	response := request(mux, http.MethodGet, getTarget(short), "")
	if response.Code != http.StatusOK {
		t.Fatalf("GET status = %d: %s", response.Code, response.Body)
	}
	var result map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result["url"] != original {
		t.Fatalf("GET response = %v", result)
	}
}

func createShortURL(t *testing.T, mux *http.ServeMux, original string) string {
	t.Helper()
	body, err := json.Marshal(map[string]string{"url": original})
	if err != nil {
		t.Fatal(err)
	}
	response := request(mux, http.MethodPost, "/api/v1/urls", string(body))
	if response.Code != http.StatusCreated {
		t.Fatalf("POST status = %d: %s", response.Code, response.Body)
	}
	if response.Header().Get("Content-Type") != "application/json" {
		t.Fatal("missing JSON content type")
	}
	var result map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || !strings.HasPrefix(result["short_url"], "https://short.ozon/") {
		t.Fatalf("POST response = %v", result)
	}
	return result["short_url"]
}

func getTarget(shortURL string) string {
	return "/api/v1/urls?" + url.Values{"short_url": {shortURL}}.Encode()
}

func TestCreateShortURLRejectsInvalidBody(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"malformed JSON", "{"},
		{"missing original", "{}"},
		{"invalid field type", `{"url":42}`},
		{"null body", "null"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := request(newMux(inmemory.NewURLRepository()), http.MethodPost, "/api/v1/urls", tc.body)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d: %s", response.Code, response.Body)
			}
		})
	}
}

func TestGetOriginalURLRejectsInvalidQuery(t *testing.T) {
	for _, tc := range []struct{ name, shortURL string }{
		{"invalid URL", "hello"},
		{"invalid code", "https://short.ozon/bad"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := request(newMux(inmemory.NewURLRepository()), http.MethodGet, getTarget(tc.shortURL), "")

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d: %s", response.Code, response.Body)
			}
		})
	}
}

func TestGetOriginalURLReturnsNotFound(t *testing.T) {
	mux := newMux(inmemory.NewURLRepository())

	response := request(mux, http.MethodGet, getTarget("https://short.ozon/AAAAAAAAAA"), "")

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d: %s", response.Code, response.Body)
	}
}

func TestHTTPRouting(t *testing.T) {
	mux := newMux(inmemory.NewURLRepository())
	for _, tc := range []struct {
		name, method, target string
		status               int
	}{
		{"wrong method", http.MethodPut, "/api/v1/urls", http.StatusMethodNotAllowed},
		{"unknown route", http.MethodGet, "/not-a-route", http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := request(mux, tc.method, tc.target, "")

			if response.Code != tc.status {
				t.Fatalf("status = %d; want %d", response.Code, tc.status)
			}
		})
	}
}

type failingRepository struct{ t *testing.T }

func (r failingRepository) checkContext(ctx context.Context) {
	r.t.Helper()
	deadline, ok := ctx.Deadline()
	if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > 5*time.Second {
		r.t.Error("repository did not receive the request deadline")
	}
}

func (r failingRepository) Create(ctx context.Context, _ *domain.URL) (*domain.URL, error) {
	r.checkContext(ctx)
	return nil, errors.New("private database details")
}

func (r failingRepository) GetByShortCode(ctx context.Context, _ string) (*domain.URL, error) {
	r.checkContext(ctx)
	return nil, errors.New("private database details")
}

func (r failingRepository) GetByOriginalURL(context.Context, string) (*domain.URL, error) {
	r.t.Fatal("unexpected GetByOriginalURL")
	return nil, nil
}

func TestInternalErrorsAreNotExposed(t *testing.T) {
	for _, tc := range []struct{ method, target, body string }{
		{http.MethodPost, "/api/v1/urls", `{"url":"https://ozon.ru"}`},
		{http.MethodGet, getTarget("https://short.ozon/AAAAAAAAAA"), ""},
	} {
		t.Run(tc.method, func(t *testing.T) {
			mux := newMux(failingRepository{t: t})

			response := request(mux, tc.method, tc.target, tc.body)

			if response.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d", response.Code)
			}
			if strings.Contains(response.Body.String(), "private database details") {
				t.Fatal("response exposed internal details")
			}
		})
	}
}
