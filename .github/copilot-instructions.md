GitHub Copilot Instructions

## Project Structure

- Go project, main module in `idm/inner`, entry: `idm/cmd/main.go`.
- Domains: `employee`, `role`, `info`, `web`.
- Migrations: `idm/migrations`.
- Tests: `idm/tests`.

## Libraries

- `sqlx` for database (use parameterized queries).
- `fiber` for HTTP server.
- `validator/v10` for validation.
- `zap` for logging.
- `testify/assert` for testing.

## Code Guidelines

- Use idiomatic Go and `gofmt`.
- Always use `context.Context` for DB and external calls.
- Handle and log errors.
- Use DTOs with explicit JSON tags.
- Separate controllers, services, repositories.
- Parameterize all SQL queries.
- Comment complex code briefly.

## Testing Guidelines

- Use `t.Run` for grouping test cases.
- Use `assert` for checks.
- Test HTTP with `httptest.NewRequest` and `app.Test`.
- Parse JSON with `json.NewDecoder`.
- Clean and seed DB in tests using fixtures.
- Cover positive, negative, and edge cases.
- Ensure migrations or table creation are automatic via fixture or migration tool.
- Use the same DB connection for fixtures and app in integration tests.