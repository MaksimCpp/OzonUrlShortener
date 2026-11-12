CREATE TABLE urls(
    id BIGSERIAL PRIMARY KEY,
    url TEXT UNIQUE NOT NULL,
    short_url VARCHAR(10) UNIQUE NOT NULL CHECK (short_url ~ '^[A-Za-z0-9_]{10}$'),

    CONSTRAINT urls_url_length_check
        CHECK (octet_length(url) BETWEEN 1 AND 2000)
);