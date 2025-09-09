-- Initialize the spectro_lab database
-- This script runs when the PostgreSQL container starts for the first time

-- Create the database if it doesn't exist (this is handled by POSTGRES_DB env var)
-- But we can add any additional setup here

-- Create extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Set timezone
SET timezone = 'UTC';

-- Create a read-only user for monitoring (optional)
-- CREATE USER spectro_lab_readonly WITH PASSWORD 'readonly_password';
-- GRANT CONNECT ON DATABASE spectro_lab TO spectro_lab_readonly;
-- GRANT USAGE ON SCHEMA public TO spectro_lab_readonly;
-- GRANT SELECT ON ALL TABLES IN SCHEMA public TO spectro_lab_readonly;
-- ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO spectro_lab_readonly;
