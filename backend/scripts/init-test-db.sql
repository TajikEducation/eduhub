-- Выполняется один раз при первом старте контейнера (docker-entrypoint-initdb.d).
-- Создаёт отдельную БД для интеграционных тестов (`make test-integration`),
-- отдельно от dev-базы `eduhub`, чтобы тесты никогда не трогали dev-данные.
CREATE DATABASE eduhub_test;
