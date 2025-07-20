# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Commands

### Build
```bash
# Production build (Linux AMD64)
make release

# Regular build
go build -o promptpal main.go
```

### Testing
```bash
# Run all tests with race detection and coverage
go test -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run a single test file
go test -v ./routes/prompt_test.go

# Run tests in a specific package
go test -v ./service/...
```

### Code Generation
```bash
# Install mockery if not present
go install github.com/vektra/mockery/v2@v2.42.0

# Generate all code (mocks, ent entities)
go generate ./...
```

### Development
```bash
# Run the application locally
go run main.go

# Docker deployment
docker run -v $(pwd)/.env:/usr/app/.env -p 7788:7788 annatarhe/prompt-pal:latest
```

## Architecture Overview

PromptPal is a monolithic Go API server for prompt management with AI integration. The architecture follows a layered design pattern:

```
HTTP Request → Routes (Gin handlers) → GraphQL Schema/Resolvers → Services → Database (Ent ORM)
```

### Key Components

1. **Routes Layer** (`routes/`)
   - HTTP handlers using Gin framework
   - Authentication middleware for JWT/OAuth/Web3
   - GraphQL endpoint handler
   - SSE streaming for AI responses

2. **GraphQL Layer** (`schema/`)
   - Type definitions in `schema/types/`
   - Resolver implementations for queries/mutations
   - Context-based authentication and authorization

3. **Service Layer** (`service/`)
   - Business logic implementation
   - AI provider abstractions (OpenAI, Gemini)
   - Database operations via Ent ORM
   - Redis caching integration

4. **Database Layer** (`ent/`)
   - Auto-generated entity code
   - Schema definitions in `ent/schema/`
   - Support for PostgreSQL, MySQL, and SQLite

### Authentication Flow

The system supports three authentication methods:
1. **JWT Tokens**: For API access (project tokens)
2. **OAuth2**: Google SSO integration
3. **Web3**: Ethereum wallet authentication via signed messages

Authentication middleware in `routes/auth.middleware.go` validates requests and injects user context.

### AI Service Integration

AI services implement a common interface allowing provider switching:
- OpenAI integration in `service/ai.openai.go`
- Google Gemini integration in `service/ai.gemini.go`
- Streaming responses via SSE in `routes/prompt.go`

### Testing Strategy

- Unit tests alongside implementation files (`*_test.go`)
- Integration tests use PostgreSQL in CI
- Mocks generated via mockery for service interfaces
- Test environment variables in CI workflow

### Database Migrations

Ent handles schema migrations automatically. To modify database schema:
1. Edit entity schemas in `ent/schema/`
2. Run `go generate ./...` to regenerate code
3. Ent will handle migration on startup

### Environment Configuration

Required environment variables (see `.env` template):
- `JWT_TOKEN_KEY`: Secret for JWT signing
- `DB_TYPE` and `DB_DSN`: Database configuration
- `ADMIN_LIST`: Comma-separated admin addresses
- `OPENAI_BASE_URL`: OpenAI API endpoint

## Commit Rules

When creating commits in this repository, follow these conventions:

### Commit Message Format
Use conventional commit format:
```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Commit Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, missing semi-colons, etc)
- `refactor`: Code refactoring without functionality changes
- `test`: Adding or updating tests
- `chore`: Maintenance tasks, dependency updates
- `perf`: Performance improvements
- `ci`: Changes to CI configuration

### Examples
```
feat(auth): add Web3 wallet authentication support
fix(api): resolve GraphQL query timeout issues
docs: update API documentation for new endpoints
test(service): add unit tests for AI service providers
```

### Guidelines
- Keep the first line under 50 characters
- Use imperative mood in the description ("add" not "added")
- Include scope when applicable (auth, api, service, etc.)
- Reference issue numbers in the footer when applicable
- Always run tests before committing