package utils

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/MaksimCpp/OzonUrlShortener/internal/domain"
)

func CheckURL(urlString string) error {
	if urlString == "" {
		return domain.ErrInvalidURL
	}

	parsed, err := url.ParseRequestURI(urlString)
	if err != nil {
		return domain.ErrInvalidURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return domain.ErrInvalidURL
	}

	if parsed.Hostname() == "" {
		return domain.ErrInvalidURL
	}

	if strings.ContainsAny(parsed.Hostname(), " \t\n") {
		return domain.ErrInvalidURL
	}

	if parsed.Port() != "" {
		port, err := strconv.Atoi(parsed.Port())

		if err != nil || port < 1 || port > 65535 {
			return domain.ErrInvalidURL
		}
	}

	return nil
}
