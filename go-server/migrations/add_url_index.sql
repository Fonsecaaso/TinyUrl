-- Add index on url column for faster lookups
CREATE INDEX IF NOT EXISTS idx_urls_url ON urls(url);
