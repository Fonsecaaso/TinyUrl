-- Add unique constraint on original_url column to enable UPSERT optimization
-- This prevents duplicate URLs and allows INSERT ... ON CONFLICT to work
ALTER TABLE urls ADD CONSTRAINT urls_original_url_unique UNIQUE (original_url);
