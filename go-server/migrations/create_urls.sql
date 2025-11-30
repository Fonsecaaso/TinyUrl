CREATE TABLE urls (
    original_url VARCHAR NOT NULL,
    id VARCHAR UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);