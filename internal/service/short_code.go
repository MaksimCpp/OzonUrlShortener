package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	shortCodeAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	shortCodeLength   = 10
)

func generateShortCode() (string, error) {
	code := make([]byte, shortCodeLength)
	limit := big.NewInt(int64(len(shortCodeAlphabet)))

	for i := range code {
		index, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", fmt.Errorf("Generate short code: %w", err)
		}
		code[i] = shortCodeAlphabet[index.Int64()]
	}

	return string(code), nil
}
