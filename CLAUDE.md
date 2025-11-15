# CLAUDE.md - AI Assistant Guide for Bunkai

## Project Overview

**Bunkai** (分解 - Japanese for "decomposition/analysis") is a full-stack web application for managing text sentences/snippets with associated URLs. It features user authentication and basic CRUD operations.

**Project Status**: Legacy/Learning Project (circa 2014-2015)
**Primary Language**: Go (backend), JavaScript/Ember.js (frontend)
**Database**: PostgreSQL with GORP ORM

## Codebase Structure

```
bunkai/
├── server.go                    # Main Go application (232 lines, monolithic)
├── .gitignore                   # Ignores: gin-bin, node_modules, build/, components/
├── client/                      # Frontend application
│   ├── package.json            # npm build dependencies (Grunt, JSHint, Uglify)
│   ├── bower.json              # Frontend dependencies (Ember 1.5, Ember Data)
│   ├── Gruntfile.js            # Build automation configuration
│   ├── .bowerrc                # Bower config (components directory)
│   └── src/
│       ├── test.js             # Sample/demo JavaScript file
│       └── components/         # Bower-managed libraries (gitignored)
├── public/                      # Static assets directory
│   └── test.html               # Test/placeholder file
└── templates/                   # Go server-side templates
    ├── layout.tmpl             # Base HTML layout (Bootstrap 3.1.0 CDN)
    └── home.tmpl               # Home page template
```

## Tech Stack & Dependencies

### Backend (Go)
- **Web Framework**: Martini (DEPRECATED - abandoned project)
- **ORM**: GORP (gopkg.in/gorp.v1)
- **Database Driver**: lib/pq (PostgreSQL)
- **Rendering**: martini-contrib/render
- **Sessions**: martini-contrib/sessions (cookie-based)
- **Password Hashing**: MD5 (INSECURE - see Security Considerations)

### Frontend
- **Framework**: Ember.js v1.5 (OUTDATED - current is 5.x)
- **Data Layer**: Ember Data v1.0.0-beta.4
- **UI Framework**: Bootstrap 3.1.0 (loaded via CDN)
- **Build System**: Grunt v0.4.5 (DEPRECATED - prefer Vite/Webpack/esbuild)
- **Package Managers**:
  - Bower for frontend libs (DEPRECATED - prefer npm)
  - npm for build tools

### Database
- **RDBMS**: PostgreSQL
- **Database Name**: `bunkai` (hardcoded)
- **Connection**: No SSL, local connection assumed
- **Schema Management**: Auto-create tables on startup via GORP

## Architecture & Patterns

### Application Architecture
**Type**: Monolithic hybrid application
- Single Go file contains entire backend logic
- Server-side rendered initial HTML via Go templates
- Client-side Ember.js app handles dynamic interactions
- RESTful API for AJAX communication

### Code Organization Pattern
```
server.go structure:
├── Constants (lines 22-26)
├── Database Setup (lines 28-40)
├── Utility Functions (lines 42-52)
├── Models/Structs (lines 54-85)
├── Validation (lines 87-100)
├── Main & Routing (lines 102-136)
├── Middleware (lines 138-164)
└── HTTP Handlers (lines 166-231)
```

### Data Models

**Sentence Model** (server.go:54-69)
```go
type Sentence struct {
    Id        int64
    UserId    int64  // Foreign key to User
    Text      string
    Url       string
    CreatedAt int64  // Unix nanoseconds
}
```

**User Model** (server.go:71-85)
```go
type User struct {
    Id        int64
    Email     string
    Password  string  // MD5 hash (INSECURE)
    CreatedAt int64
    UpdatedAt int64   // Not currently used
}
```

### API Endpoints

**Public Routes:**
- `GET /` - Home page (renders home.tmpl)

**Authentication:**
- `POST /api/login` - User login (sets session cookie)
- `POST /api/users` - User registration

**Protected Routes** (require authentication):
- `POST /api/sentences` - Create sentence
- `GET /api/sentences` - List user's sentences
- `DELETE /api/sentences/:id` - Delete sentence by ID
- `POST /api/users/logout` - Logout (clears session)
- `GET /api/users/me` - Get current user info

### Middleware & Dependency Injection

Martini uses dependency injection for handlers:
```go
func SentenceCreate(ren render.Render, req *http.Request, dbmap *gorp.DbMap, usr User) {
    // Dependencies automatically injected by Martini
}
```

**Key Middleware:**
- `RequireLogin` (server.go:153-164) - Session-based auth guard
- `AssetMap` (server.go:138-140) - Dynamic JS component discovery
- `sessions.Sessions` - Cookie session management

## Development Workflow

### Prerequisites
```bash
# System dependencies
- Go (pre-1.11, no modules support)
- PostgreSQL
- Node.js & npm
- Bower (install globally: npm install -g bower)
```

### Initial Setup
```bash
# 1. Create PostgreSQL database
createdb bunkai

# 2. Install Go dependencies
go get github.com/coopernurse/gorp
go get github.com/go-martini/martini
go get github.com/lib/pq
go get github.com/martini-contrib/render
go get github.com/martini-contrib/sessions

# 3. Install frontend dependencies
cd client
npm install
bower install

# 4. Build frontend
grunt
```

### Running the Application
```bash
# Development mode (from project root)
go run server.go

# Production build
go build -o bunkai server.go
./bunkai

# Frontend build (in client/ directory)
cd client && grunt
```

### Environment Configuration
**Hardcoded values in server.go:22-26:**
- Cookie Secret: `"secretedesse"` (INSECURE)
- Database Name: `"bunkai"`
- Bower Components Path: `"./client/src/components/"`

⚠️ **Note**: No environment variable support. All config is hardcoded.

## Code Conventions

### Go Code Style
1. **Naming**:
   - Exported types: PascalCase (`Sentence`, `User`)
   - Functions: PascalCase for exported, camelCase for private
   - Constants: PascalCase with descriptive names

2. **Error Handling**:
   - `PanicIf(err)` utility used throughout (server.go:42-46)
   - Critical errors cause panic (not graceful)
   - HTTP errors return status codes (401, 400, 403)

3. **Database Queries**:
   - Raw SQL with positional parameters (`$1`, `$2`)
   - GORP ORM for basic CRUD
   - No prepared statements (potential SQL injection risk)

4. **Validation**:
   - Regex-based validation (server.go:87-100)
   - Model method: `(sen *Sentence) Validate()`
   - URL validation regex: `http(s)?://([\w-]+\.)+[\w-]+(/[\w- ./?%&=]*)?`

### Frontend Code Style
- **Ember 1.x Conventions** (outdated)
- **Bower Components**: Auto-discovered via `JsComponent()` function
- **Build**: Grunt-based minification (test.js → test.min.js)

### File Organization
- **Single-file backend**: All Go code in server.go
- **Template-based views**: Go HTML templates (not JSON API)
- **Static assets**: Served from public/ and client/src/components/

## Testing & Quality

### Current State
⚠️ **No test coverage**
- No Go test files (`*_test.go`)
- No frontend tests
- JSHint configured but minimal usage
- Nodeunit dependency present but unused

### Quality Tools Available
- `grunt-contrib-jshint` - JavaScript linting
- `grunt-contrib-nodeunit` - Test framework (not configured)

### Recommended Testing Approach (for future development)
```bash
# Go testing
go test ./...

# Frontend testing (if added)
cd client && npm test
```

## Security Considerations

### Critical Security Issues

1. **MD5 Password Hashing** (server.go:48-52, 82)
   - MD5 is cryptographically broken
   - No salt used
   - **Recommendation**: Migrate to bcrypt or argon2
   ```go
   // Current (INSECURE):
   Password: Md5(password)

   // Should be:
   // import "golang.org/x/crypto/bcrypt"
   // hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
   ```

2. **Hardcoded Secrets** (server.go:23)
   ```go
   const CookieSecret = "secretedesse"  // INSECURE
   ```
   - **Recommendation**: Use environment variables

3. **No HTTPS Enforcement**
   - Cookies transmitted over HTTP
   - Credentials sent in plain text

4. **SQL Injection Risk** (server.go:175, 207, etc.)
   - Using string interpolation in some queries
   - Example: `dbmap.Exec("DELETE FROM sentences WHERE id= $1", params["id"])`
   - **Recommendation**: Always use parameterized queries

5. **No CSRF Protection**
   - POST endpoints lack CSRF tokens
   - Session cookies vulnerable to CSRF attacks

6. **No Input Sanitization**
   - User input directly inserted into database
   - XSS potential in rendered content

7. **Error Disclosure**
   - Panics may expose stack traces in production
   - Database errors returned to client

### Security Best Practices for AI Assistants
When working on this codebase:
- ✅ **DO** recommend security improvements
- ✅ **DO** use parameterized queries
- ✅ **DO** validate and sanitize all user input
- ❌ **DON'T** introduce new hardcoded secrets
- ❌ **DON'T** use MD5 for new password-related code
- ❌ **DON'T** expose sensitive data in API responses

## Common Tasks & Commands

### Database Operations
```bash
# Create database
createdb bunkai

# Drop and recreate (reset)
dropdb bunkai && createdb bunkai

# Connect via psql
psql bunkai

# View tables
\dt

# View sentences
SELECT * FROM sentences;

# View users
SELECT * FROM users;
```

### Development Commands
```bash
# Run server (auto-creates tables)
go run server.go

# Build binary
go build -o bunkai server.go

# Install new Go dependency (pre-modules)
go get github.com/package/name

# Frontend build
cd client && grunt

# Install new Bower component
cd client && bower install <package> --save
```

### Git Workflow
```bash
# Create feature branch with required prefix
git checkout -b claude/feature-name-<session-id>

# Commit changes
git add .
git commit -m "Descriptive commit message"

# Push to remote (use -u for new branches)
git push -u origin claude/feature-name-<session-id>
```

## AI Assistant Guidelines

### When Making Changes

1. **Understand the Monolith**
   - All backend logic is in server.go
   - Changes affect the entire application
   - No module boundaries to respect

2. **Database Schema Changes**
   - Modify struct definitions (lines 54-85)
   - Update GORP mappings in SetupDB() (lines 28-40)
   - Tables auto-recreate on restart (DATA LOSS in production)

3. **Adding New Endpoints**
   ```go
   // Add to appropriate group in main() (lines 115-133)
   m.Group("/api", func(m martini.Router) {
       m.Post("/new-endpoint", NewHandler)
   })
   ```

4. **Authentication Requirements**
   - Public endpoints: Place outside RequireLogin middleware
   - Protected endpoints: Place inside RequireLogin group
   - User object auto-injected as parameter in protected routes

5. **Error Handling Pattern**
   ```go
   // Critical errors (database, etc.)
   PanicIf(err)

   // User errors (validation, etc.)
   ren.JSON(400, map[string]string{"error": err.Error()})
   ```

### Code Reading Tips

1. **Entry Point**: Start at `main()` (server.go:102)
2. **Routing**: All routes defined in main() function
3. **Database Access**: Search for `dbmap.` calls
4. **Session Management**: Look for `s.Get()` and `s.Set()` calls
5. **Request Handling**: Functions with `render.Render` parameter

### Common Patterns

**Handler Signature**:
```go
func HandlerName(ren render.Render, req *http.Request, dbmap *gorp.DbMap, usr User) {
    // ren: Render JSON/HTML responses
    // req: HTTP request (form values, etc.)
    // dbmap: Database connection
    // usr: Current authenticated user (if RequireLogin middleware active)
}
```

**JSON Response**:
```go
ren.JSON(statusCode, data)
```

**HTML Response**:
```go
ren.HTML(200, "template-name", data)
```

**Form Value Access**:
```go
value := req.FormValue("field-name")
```

**Database Insert**:
```go
err := dbmap.Insert(&model)
PanicIf(err)
```

**Database Query**:
```go
var result Type
err := dbmap.SelectOne(&result, "SQL QUERY", param1, param2)
```

### Modernization Opportunities

If asked to improve the project:

1. **Backend**:
   - Migrate to Go modules (`go mod init`)
   - Replace Martini with Gin or Echo (active projects)
   - Implement proper password hashing (bcrypt)
   - Add environment variable configuration
   - Separate concerns (handlers, models, database in different files)
   - Add comprehensive error handling

2. **Frontend**:
   - Upgrade to modern Ember (5.x) or migrate to React/Vue
   - Replace Bower with npm
   - Replace Grunt with Vite or esbuild
   - Add TypeScript for type safety

3. **Security**:
   - HTTPS enforcement
   - CSRF protection
   - Input sanitization
   - Rate limiting
   - Security headers
   - Audit logging

4. **Testing**:
   - Add Go unit tests
   - Add integration tests
   - Add frontend tests
   - CI/CD pipeline

5. **Infrastructure**:
   - Dockerfile for containerization
   - docker-compose.yml for local development
   - Environment-based configuration
   - Migration system (goose, migrate)

### Important Notes for AI Assistants

✅ **Safe Operations**:
- Reading code and explaining functionality
- Adding new features following existing patterns
- Improving error handling
- Adding validation
- Writing tests
- Documentation

⚠️ **Proceed with Caution**:
- Database schema changes (no migration system)
- Changing authentication logic
- Modifying core middleware
- Dependency updates (compatibility issues likely)

❌ **Avoid**:
- Breaking changes to API contracts (no versioning)
- Removing existing functionality without discussion
- Major refactoring without user approval
- Changing database connection logic without testing

### Questions to Ask Users

Before implementing changes, consider asking:
- "Should I preserve backward compatibility with the existing API?"
- "Do you want me to implement proper security (bcrypt, env vars)?"
- "Should I add tests for this functionality?"
- "Is this a learning project or production code?"
- "Do you want me to modernize dependencies or keep them as-is?"

## Project History & Context

This appears to be a learning/prototype project from 2014-2015 based on:
- Technology versions (Ember 1.5, Grunt, Bower)
- Pre-Go modules era
- Martini framework (deprecated ~2015)
- Simple CRUD application pattern

The project demonstrates fundamental concepts:
- RESTful API design
- Session-based authentication
- ORM usage
- Template rendering
- Form validation
- Database relationships

## Quick Reference

### File Locations
- Main server: `/home/user/bunkai/server.go`
- Frontend config: `/home/user/bunkai/client/package.json`
- Templates: `/home/user/bunkai/templates/*.tmpl`
- Gitignore: `/home/user/bunkai/.gitignore`

### Key Constants
```go
CookieSecret       = "secretedesse"
DBName             = "bunkai"
BowerComponentPath = "./client/src/components/"
```

### Database Tables
- `sentences` - User text snippets with URLs
- `users` - Authentication and user data

### Port
- Default: Martini runs on port 3000
- Override: Set `PORT` environment variable

---

**Last Updated**: 2025-11-15
**Repository**: bunkai (ukitazume/bunkai)
**AI Assistant**: Claude Code by Anthropic
