# go-social

A simple social media REST API built with Go's standard `net/http` package. Built this as a simple go practice project, i never implemented the other planned endpoints cause i kinda already got the memo.

## Features

- Authentication with JWT (access + refresh token rotation)
- Cookie-based session management
- Rate limiting
- Redis caching
- Email notifications via Resend
- Cursor-based pagination
- Swagger documentation

## Architecture

The project follows a **hexagonal (ports and adapters)** style:

- **Handlers** define interfaces for their dependencies (repos, cache, mailer)
- **Repositories** implement those interfaces against PostgreSQL
- **Services** (cache, mailer) implement their own interfaces
- **main.go** is the composition root

Swapping out a dependency only requires a new implementation that satisfies the interface, with no changes to handlers or business logic.

## Directory Structure

```
.
├── cmd/
│   └── api/
│       └── main.go           # Entry point, wires dependencies
├── internal/
│   ├── cache/
│   │   ├── interface.go      # CacheService interface
│   │   └── redis.go          # Redis implementation
│   ├── db/
│   │   └── db.go             # Database connection
│   ├── handlers/
│   │   ├── interfaces.go     # UsersRepo, PostsRepo interfaces
│   │   ├── auth.go           # Auth handlers
│   │   └── posts.go          # Posts handlers
│   ├── mailer/
│   │   ├── interface.go      # MailerService interface + Mail model
│   │   └── mailer.go         # Resend implementation
│   ├── middlewares/
│   │   ├── is-authorized.go  # JWT auth middleware
│   │   ├── rate-limiter.go   # Rate limiting middleware
│   │   └── request-logger.go # Request logging middleware
│   ├── models/               # Domain models
│   ├── repository/
│   │   ├── users.go          # UsersRepository (implements UsersRepo)
│   │   └── posts.go          # PostsRepository (implements PostsRepo)
│   ├── routers/              # Route registration
│   └── templates/            # Email HTML templates
└── pkg/
    └── utils/                # Shared utilities (JWT, hashing, validation)
```

## Getting Started

Copy `.env.example` to `.env` and fill in the values, then:

```bash
go run cmd/api/main.go
```

Swagger UI is available at `http://localhost:{PORT}/swagger/index.html` in non-production environments.

## Docker

Build the image:

```bash
docker build -t go-social:local .
```

Run the container:

```bash
docker run --env-file .env -p 3000:3000 go-social
```
