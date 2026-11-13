package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/MaksimCpp/OzonUrlShortener/internal/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgreSQLURLRepository struct {
	pool *pgxpool.Pool
}

func NewPostgreSQLURLRepository(pool *pgxpool.Pool) *PostgreSQLURLRepository {
	return &PostgreSQLURLRepository{
		pool: pool,
	}
}

func (repo *PostgreSQLURLRepository) Create(
	ctx context.Context, url *domain.URL,
) (*domain.URL, error) {
	query := `
		INSERT INTO urls (original_url, short_code)
		VALUES ($1, $2)
		RETURNING id, original_url, short_code;
	`

	result := domain.URL{}

	err := repo.pool.QueryRow(
		ctx,
		query,
		url.OriginalURL,
		url.ShortCode,
	).Scan(
		&result.ID,
		&result.OriginalURL,
		&result.ShortCode,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				switch pgErr.ConstraintName {
				case "urls_original_url_key":
					return nil, domain.ErrURLAlreadyExist
				case "urls_short_code_key":
					return nil, domain.ErrShortCodeAlreadyExists
				}

			case pgerrcode.StringDataRightTruncationDataException:
				return nil, domain.ErrInvalidShortCode

			case pgerrcode.CheckViolation:
				urlLenConConstraintName := "urls_url_length_check"
				urlShortCodeConstraintName := "urls_short_code_check"

				switch pgErr.ConstraintName {
				case urlLenConConstraintName:
					return nil, domain.ErrInvalidURLLength

				case urlShortCodeConstraintName:
					return nil, domain.ErrInvalidShortCode
				}
			}
		}

		return nil, fmt.Errorf("Create URL: %w", err)
	}

	return &result, nil
}

func (repo *PostgreSQLURLRepository) GetByShortCode(
	ctx context.Context, shortCode string,
) (*domain.URL, error) {
	query := `
		SELECT id, original_url, short_code
		FROM urls
		WHERE short_code = $1;
	`

	result := domain.URL{}

	err := repo.pool.QueryRow(
		ctx,
		query,
		shortCode,
	).Scan(
		&result.ID,
		&result.OriginalURL,
		&result.ShortCode,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrURLNotFound
		}

		return nil, fmt.Errorf("Get URL by short code: %w", err)
	}

	return &result, nil
}

func (repo *PostgreSQLURLRepository) GetByOriginalURL(
	ctx context.Context, originalURL string,
) (*domain.URL, error) {
	query := `
		SELECT id, original_url, short_code
		FROM urls
		WHERE original_url = $1;
	`

	result := domain.URL{}

	err := repo.pool.QueryRow(
		ctx,
		query,
		originalURL,
	).Scan(
		&result.ID,
		&result.OriginalURL,
		&result.ShortCode,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrURLNotFound
		}

		return nil, fmt.Errorf("Get URL by original URL: %w", err)
	}

	return &result, nil
}
