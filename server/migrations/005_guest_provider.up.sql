ALTER TABLE users DROP CONSTRAINT IF EXISTS users_provider_provider_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS users_provider_unique ON users(provider, provider_id)
    WHERE provider != 'guest';
