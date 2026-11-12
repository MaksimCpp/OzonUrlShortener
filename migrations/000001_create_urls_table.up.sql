CREATE TABLE urls(
    id BIGSERIAL PRIMARY KEY,
    original_url TEXT UNIQUE NOT NULL,
    short_code VARCHAR(10) UNIQUE NOT NULL
        CONSTRAINT urls_short_code_check CHECK (short_code ~ '^[A-Za-z0-9_]{10}$'),

    CONSTRAINT urls_url_length_check
        CHECK (octet_length(original_url) BETWEEN 1 AND 2000)
);
