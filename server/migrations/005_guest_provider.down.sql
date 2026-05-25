DROP INDEX IF EXISTS users_provider_unique;
ALTER TABLE users ADD CONSTRAINT users_provider_provider_id_key UNIQUE (provider, provider_id);
