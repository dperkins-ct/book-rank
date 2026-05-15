# Project Instructions for AI Agents

This file provides instructions and context for AI coding agents working on this project.

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:7510c1e2 -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md for details and anti-patterns.

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->


## Build & Test

### Backend (Go)
```bash
# Build and run
make build           # Build the application
make run            # Build and run locally
make test           # Run all tests
make test-coverage  # Run tests with coverage report
make lint           # Run golangci-lint

# Development setup
make setup          # Install dev dependencies and start database
make deps           # Install required linters and migration tools

# Database
make db-migrate     # Run database migrations
make db-reset       # Reset database (drop and recreate)
```

### Frontend (React)
```bash
cd frontend
npm install         # Install dependencies
npm run dev         # Start development server
npm run build       # Build for production
npm run test        # Run tests
npm run lint        # Run ESLint
```

### Full Stack Development
```bash
make start          # Start entire stack (database + backend + frontend)
make stop-all       # Stop all services
make restart        # Full restart
make health         # Check application health
```

## Architecture Overview

BookRank is a full-stack book rating and recommendation system built with Go backend and React frontend.

### Backend Architecture (Go)
- **`cmd/server/main.go`** - Application entry point, calls `run()` function that returns error
- **`internal/`** - Private application code
  - `api/` - HTTP handlers and routing (grouped by domain: auth, book, comparison)
  - `service/` - Business logic layer (ELO rating system, comparisons, recommendations)
  - `repository/` - Data access layer with GORM
  - `models/` - Database models and domain entities
  - `auth/` - Authentication and authorization
- **`pkg/`** - Public/reusable packages (if any)

### Frontend Architecture (React)
- **Component-based React SPA** with React Router for navigation
- **Context API** for global state (Auth, Book contexts)
- **Services layer** (`services/api.js`) for API communication
- **Atomic Design** - components organized by common, layout, and domain-specific

### Data Flow
1. **Authentication** - JWT tokens for secure API access
2. **Book Comparisons** - ELO rating system for pairwise comparisons
3. **Rankings** - User-specific book rankings derived from comparison history
4. **Recommendations** - Personalized suggestions based on ranking patterns

### Key Systems
- **ELO Rating Engine** - Mathematical ranking system for book preferences
- **Comparison Engine** - Intelligent book pairing for optimal ranking discovery
- **Recommendation Engine** - Personalized book suggestions based on user preferences

## Conventions & Patterns

### Go Code Patterns

#### Project Structure
- **Standard layout**: `cmd/`, `internal/`, `pkg/` structure
- **Domain-driven organization**: Group by business domain (auth, book, comparison)
- **Layer separation**: Clear boundaries between API, service, and repository layers

#### Main Function Pattern
```go
// cmd/server/main.go
func main() {
    if err := run(); err != nil {
        log.Fatal(err)
    }
}

func run() error {
    // Application setup and startup logic
    // Returns error for proper error handling
}
```

#### Table-Driven Tests
```go
func TestServiceMethod(t *testing.T) {
    tests := map[string]struct {
        input    InputType
        expected ExpectedType
        wantErr  bool
    }{
        "valid input returns success": {
            input:    validInput,
            expected: expectedResult,
            wantErr:  false,
        },
        "invalid input returns error": {
            input:   invalidInput,
            wantErr: true,
        },
    }

    for name, tt := range tests {
        t.Run(name, func(t *testing.T) {
            result, err := service.Method(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Method() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && result != tt.expected {
                t.Errorf("Method() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

#### Service Layer Patterns
- **Repository dependency injection**: Services depend on repository interfaces
- **Error wrapping**: Use `fmt.Errorf("failed to X: %w", err)` for context
- **Validation**: Input validation at service layer boundaries
- **Transaction handling**: Repository layer manages database transactions

#### API Handler Patterns
- **Middleware chain**: Auth, CORS, logging, rate limiting
- **Consistent error responses**: Standard HTTP status codes and JSON error format
- **Request/Response DTOs**: Separate from internal models
- **Context propagation**: User context through middleware

### Server Construction Pattern

#### NewServer Constructor
The `NewServer` function is the backbone of any Go service, creating the main `http.Handler`:

```go
func NewServer(
    logger *Logger,
    config *Config,
    commentStore *commentStore,
    anotherStore *anotherStore,
) http.Handler {
    mux := http.NewServeMux()
    addRoutes(
        mux,
        logger,
        config,
        commentStore,
        anotherStore,
    )
    var handler http.Handler = mux
    handler = someMiddleware(handler)
    handler = someMiddleware2(handler)
    handler = someMiddleware3(handler)
    return handler
}
```

**Key principles:**
- Takes all dependencies as arguments (dependency injection)
- Returns `http.Handler` when possible
- Configures its own muxer and calls `addRoutes()`
- Applies middleware in reverse order of execution
- Use `nil` for unused dependencies in tests

#### Server Setup Pattern
```go
srv := NewServer(
    logger,
    config,
    tenantsStore,
    slackLinkStore,
    msteamsLinkStore,
    proxy,
)
httpServer := &http.Server{
    Addr:    net.JoinHostPort(config.Host, config.Port),
    Handler: srv,
}

go func() {
    log.Printf("listening on %s\n", httpServer.Addr)
    if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
    }
}()

var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    <-ctx.Done()
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := httpServer.Shutdown(shutdownCtx); err != nil {
        fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
    }
}()
wg.Wait()
```

### Long Argument Lists Pattern
Prefer long argument lists over config structs for better type safety:

```go
srv := NewServer(
    logger,
    config,
    tenantsStore,
    commentsStore,
    conversationService,
    chatGPTService,
)
```

**Benefits:**
- Function forces you to provide all required arguments
- Type checking catches missing or misordered dependencies
- No need to look up struct field names
- Clear dependency requirements

### Routes Mapping Pattern

#### Centralized `routes.go`
Map the entire API surface in one place:

```go
func addRoutes(
    mux                 *http.ServeMux,
    logger              *logging.Logger,
    config              Config,
    tenantsStore        *TenantsStore,
    commentsStore       *CommentsStore,
    conversationService *ConversationService,
    chatGPTService      *ChatGPTService,
    authProxy           *authProxy,
) {
    mux.Handle("/api/v1/", handleTenantsGet(logger, tenantsStore))
    mux.Handle("/oauth2/", handleOAuth2Proxy(logger, authProxy))
    mux.HandleFunc("/healthz", handleHealthzPlease(logger))
    mux.Handle("/", http.NotFoundHandler())
}
```

**Principles:**
- Single file containing all routes
- Same dependency list as `NewServer`
- Simple, flat structure with no error handling (errors handled earlier in `run`)
- Complete API surface visible at a glance

### Enhanced Main/Run Pattern

#### Ultra-Simple Main
```go
func main() {
    ctx := context.Background()
    if err := run(ctx, os.Stdout, os.Args); err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err)
        os.Exit(1)
    }
}
```

#### Testable Run Function
```go
func run(
    ctx    context.Context,
    args   []string,
    getenv func(string) string,
    stdin  io.Reader,
    stdout, stderr io.Writer,
) error {
    ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
    defer cancel()

    // Application logic here
    // Can return errors normally
    return nil
}
```

**Operating System Dependencies:**
| Value | Type | Description |
|-------|------|-------------|
| `os.Args` | `[]string` | Command line arguments and flags |
| `os.Stdin` | `io.Reader` | Input stream |
| `os.Stdout` | `io.Writer` | Output stream |
| `os.Stderr` | `io.Writer` | Error logging stream |
| `os.Getenv` | `func(string) string` | Environment variables |
| `os.Getwd` | `func() (string, error)` | Working directory |

**Benefits:**
- Testable: can call `run()` with different arguments and streams
- No global state: enables `t.Parallel()` in tests
- Proper error handling: `main` only handles final error display
- Signal handling: graceful shutdown on Ctrl+C
- Self-contained: multiple calls don't interfere

### Testing with Args and Environment Variables

#### Testing with Different Arguments
The `args` parameter enables testing different flag combinations:

```go
func TestMyApp(t *testing.T) {
    args := []string{
        "myapp",
        "--out", outFile,
        "--fmt", "markdown",
    }
    
    err := run(ctx, args, testGetenv, testStdin, testStdout, testStderr)
    // assertions...
}
```

**Key points:**
- Use `flag.NewFlagSet` inside `run`, not global `flag` package
- Pass different `args` slices to test various flag combinations
- Enables comprehensive CLI testing

#### Testing with Mock Environment Variables
The `getenv` function parameter allows controlled environment testing:

```go
func TestWithEnvironment(t *testing.T) {
    getenv := func(key string) string {
        switch key {
        case "MYAPP_FORMAT":
            return "markdown"
        case "MYAPP_TIMEOUT":
            return "5s"
        default:
            return ""
        }
    }
    
    err := run(ctx, args, getenv, stdin, stdout, stderr)
    // assertions...
}
```

**Advantages over `t.SetEnv`:**
- Maintains `t.Parallel()` compatibility (no global state modification)
- More explicit and controllable
- Easier to test multiple environment scenarios

#### Production Main Function
In production, pass real OS functions:

```go
func main() {
    ctx := context.Background()
    if err := run(ctx, os.Args, os.Getenv, os.Stdin, os.Stdout, os.Stderr); err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err)
        os.Exit(1)
    }
}
```

### Handler Constructor Pattern

#### Maker Functions Return Handlers
Handler functions return `http.Handler` rather than implementing the interface directly:

```go
// handleSomething handles one of those web requests
// that you hear so much about.
func handleSomething(logger *Logger) http.Handler {
    thing := prepareThing()
    return http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
            // use thing to handle request
            logger.Info(r.Context(), "msg", "handleSomething")
        }
    )
}
```

**Pattern Benefits:**
- **Closure environment**: Each handler gets its own initialization space
- **Dependency injection**: Dependencies passed as parameters, available in closure
- **Initialization control**: Setup work done once during handler creation
- **Clean separation**: Handler logic separate from HTTP interface

**Important Considerations:**
- **Read-only shared data**: Only read shared data in handlers
- **Thread safety**: Use mutexes if handlers need to modify shared state
- **Stateless design**: Avoid storing program state in handlers
- **Cloud-native**: Servers may shut down/restart unpredictably
- **Horizontal scaling**: Multiple instances may run simultaneously
- **Persistent storage**: Use databases/external APIs for state that must survive restarts

**Usage in Routes:**
```go
func addRoutes(mux *http.ServeMux, logger *Logger, store *Store) {
    mux.Handle("/api/something", handleSomething(logger, store))
    mux.Handle("/api/other", handleOther(logger, store))
}
```
