package domain

import "errors"

var (
	ErrURLNotFound      = errors.New("URL not found")
	ErrURLAlreadyExist  = errors.New("URL already exist")
	ErrInvalidURLLength = errors.New("URL length must be between 1 and 2000 bytes")
	ErrInvalidShortCode = errors.New("Short code must contain exactly 10 ASCII letters, digits or underscores")
	ErrInvalidURL       = errors.New("Invalid URL")
)
