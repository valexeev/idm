GitHub Copilot Instructions

## Project Structure

- Go project, main module in `idm/inner`, entry: `idm/cmd/main.go`.
- Domains: `employee`, `role`, `info`, `web`.
- Migrations: `idm/migrations`.
- Tests: `idm/tests`.
- В каждом домене: controller, service, repository, dto, request, mocks (для role).
- Все DTO и запросы определяются в `dto.go` и `request.go` соответствующего домена.
- Все взаимодействия с БД только через интерфейс `Repo` (repository.go).
- Сервисы реализуют бизнес-логику, используют Repo и Validator.
- Контроллеры отвечают за HTTP-логику, используют сервисы, возвращают ответы через common.Response.

## Libraries

- `sqlx` for database (use parameterized queries).
- `fiber` for HTTP server.
- `validator/v10` for validation.
- `zap` for logging.
- `testify/assert` for testing.
- Для моков используйте testify/mock (и mocks/* для role).

## Code Guidelines

- Use idiomatic Go and `gofmt`.
- Always use `context.Context` for DB and external calls.
- Handle and log errors (используйте zap).
- Use DTOs with explicit JSON tags.
- Separate controllers, services, repositories.
- Parameterize all SQL queries.
- Comment complex code briefly.
- Все структуры и методы должны быть покрыты тестами (unit/integration).
- Все внешние вызовы и операции с БД используют context.Context.
- Ошибки возвращаются в стандартизированном формате через common.Response.
- Для логики авторизации используйте middleware из web/middleware.go.

## Security, Authentication, and Validation

- SSL certificates (ssl.cert, ssl.key) are stored in the `certs/` directory. Use them for HTTPS server configuration if required.
- Authentication and authorization are handled via middleware in `web/middleware.go` (e.g., JWT/OAuth2). Always protect sensitive endpoints with proper middleware.
- Use `validator/v10` for all input validation. Define validation tags in DTO/request structs. Validate all incoming data in controllers and services.
- Never store secrets or certificates in source code. Use environment variables or config files for sensitive data.

## Testing Guidelines

- Use `t.Run` for grouping test cases.
- Use `assert` for checks.
- Test HTTP with `httptest.NewRequest` and `app.Test`.
- Parse JSON with `json.NewDecoder`.
- Clean and seed DB in tests using fixtures (tests/fixture.go).
- Cover positive, negative, and edge cases.
- Ensure migrations or table creation are automatic via fixture or migration tool.
- Use the same DB connection for fixtures and app in integration tests.
- Для моков сервисов и репозиториев используйте testify/mock.
- Интеграционные тесты используют реальные миграции и фикстуры.
- Для каждого слоя (controller, service, repository) есть отдельные unit-тесты.

## Testing Guidelines (Extended)

- Write both unit and integration tests for all layers (controller, service, repository).
- Use `testify/assert` for assertions and `testify/mock` for mocks/stubs.
- For service/repository tests, mock dependencies (e.g., Repo, Validator) using testify/mock. Place custom mocks in `mocks/` if needed.
- For controller tests, use Fiber's `httptest.NewRequest` and `app.Test` to simulate HTTP requests. Always check status codes, response structure, and error messages.
- Use `t.Run` to group related test cases and table-driven tests for coverage.
- For integration tests, use real database and fixtures from `tests/fixture.go`. Ensure DB is cleaned and seeded before each test.
- Always cover positive, negative, and edge cases. Test error handling and validation logic.
- Use the same DB connection for app and fixtures in integration tests to avoid transaction issues.
- Place integration tests in `*_integration_test.go` and unit tests in `*_test.go`.
- Mock/stub only external dependencies; do not mock internal business logic.
- Use clear, descriptive test names and comments for complex scenarios.

## Архитектурные соглашения

- Каждый домен реализует паттерн controller-service-repository.
- DTO (Data Transfer Objects) о��ределяются в dto.go и request.go.
- Все SQL-запросы параметризованы.
- Для логирования используйте zap, для валидации — validator/v10.
- Код форматируется с помощью gofmt.
- Все ошибки логируются и возвращаются в стандартизированном виде.
- Для интеграционных тестов таблицы создаются автоматически через миграции или фикстуры.
