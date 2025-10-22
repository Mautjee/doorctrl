# Agent Guidelines for door-control

## Build/Run Commands
- **Build**: `go build -o door-control`
- **Run**: `./door-control` or `go run main.go`
- **Test**: No test suite currently exists. To add tests, use `go test ./...` for all packages or `go test ./handlers` for specific packages.
- **Test single package**: `go test -v ./handlers -run TestFunctionName`

## Code Style

### Imports
- Standard library first, blank line, then third-party packages, blank line, then local packages (e.g., `door-control/db`)
- Use blank identifier for side-effect imports: `_ "github.com/mattn/go-sqlite3"`

### Structure
- Handlers use struct-based pattern with dependencies injected (DB, WebAuthn, Store, Templates)
- Database operations in `db/` package, models in `models/` package, HTTP handlers in `handlers/`

### Error Handling
- Log errors with `log.Printf("Context: %v", err)` before returning HTTP errors
- Return generic error messages to clients, detailed logs server-side
- Use `http.Error()` for HTTP error responses with appropriate status codes
- Check `sql.ErrNoRows` explicitly when querying for non-existent records

### Naming & Types
- Exported types use PascalCase, unexported use camelCase
- Use explicit types for IDs: `int64` for database IDs, `[]byte` for credentials/keys
- Handler methods: `(h *HandlerName) MethodName(w http.ResponseWriter, r *http.Request)`
