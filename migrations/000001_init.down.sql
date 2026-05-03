DROP TABLE IF EXISTS refresh_tokens;
DROP TRIGGER IF EXISTS users_set_updated_date ON users;
DROP TABLE IF EXISTS users;
DROP FUNCTION IF EXISTS set_updated_date();
