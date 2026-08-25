-- Runs once on first container start (docker-entrypoint-initdb.d).
-- Creates a separate database for integration tests (`make test-integration`),
-- kept apart from the dev database `eduhub` so tests never touch dev data.
CREATE DATABASE eduhub_test;
