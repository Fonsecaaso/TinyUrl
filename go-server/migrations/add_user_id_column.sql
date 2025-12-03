ALTER TABLE urls
ADD COLUMN user_id UUID references users(id);