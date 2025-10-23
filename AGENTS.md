# Agent Guidelines for door-control

## Build/Run Commands
- **Build**: `go build -o door-control`
- **Run**: `./door-control` or `go run main.go`
- **Test**: No test suite currently exists. To add tests, use `go test ./...` for all packages or `go test ./handlers` for specific packages.
- **Test single package**: `go test -v ./handlers -run TestFunctionName`

## Security Features
- **Rate Limiting**: 5 requests per second per IP on authentication endpoints (`/register/begin`, `/register/finish`, `/login/begin`, `/login/finish`)
- **Session Management**: Sessions stored with HttpOnly, Secure, SameSite=Lax cookies
- **Session Secret**: Use `SESSION_SECRET` environment variable in production (warns if using default)

## Code Style

### Imports
- Standard library first, blank line, then third-party packages, blank line, then local packages (e.g., `door-control/db`)
- Use blank identifier for side-effect imports: `_ "github.com/mattn/go-sqlite3"`

### Structure
- Handlers use struct-based pattern with dependencies injected (DB, WebAuthn, Store, Templates)
- Database operations in `db/` package, models in `models/` package, HTTP handlers in `handlers/`

### Error Handling & Logging
- Log errors with `log.Printf("Context: %v", err)` before returning HTTP errors
- Return generic error messages to clients, detailed logs server-side
- Use `http.Error()` for HTTP error responses with appropriate status codes
- Check `sql.ErrNoRows` explicitly when querying for non-existent records
- **Logging Standards**:
  - Include user ID and IP address in security-relevant logs
  - Log all authentication attempts (success and failure)
  - Log all door unlock attempts with location data
  - Log booking creations and conflicts
  - Use descriptive log messages with context

### Naming & Types
- Exported types use PascalCase, unexported use camelCase
- Use explicit types for IDs: `int64` for database IDs, `[]byte` for credentials/keys
- Handler methods: `(h *HandlerName) MethodName(w http.ResponseWriter, r *http.Request)`
