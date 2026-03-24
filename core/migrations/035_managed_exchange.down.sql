DROP TRIGGER IF EXISTS exchange_threads_updated_at ON exchange_threads;
DROP FUNCTION IF EXISTS update_exchange_threads_updated_at();

DROP TABLE IF EXISTS exchange_items;
DROP TABLE IF EXISTS exchange_threads;
DROP TABLE IF EXISTS exchange_channels;
DROP TABLE IF EXISTS exchange_schema_registry;
DROP TABLE IF EXISTS exchange_field_registry;
