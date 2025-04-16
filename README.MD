# go-fr-project

A Firebase-like platform built with Go and PostgreSQL, similar to Supabase.

## Features

- Authentication system with signup, login, and session management
- RESTful API endpoints
- PostgreSQL database integration
- Background task worker system
- Database migrations

## Project Structure

```
/
├── cmd/
│   ├── api/         # API server entry point
│   ├── worker/      # Background job workers
│   └── migrations/  # Database migration tool
├── internal/
│   ├── auth/        # Authentication service
│   └── common/      # Shared utilities and middleware
├── api/
│   └── http/        # HTTP handlers
├── migrations/      # SQL migration files
└── config/          # Configuration files
```

## Getting Started

### Prerequisites

- Go 1.18 or higher
- PostgreSQL 12 or higher

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/go-fr-project.git
   cd go-fr-project
   ```

2. Initialize the Go module:
   ```
   go mod init go-fr-project
   ```

3. Install dependencies:
   ```
   go get github.com/lib/pq
   go get github.com/golang-migrate/migrate/v4
   go get github.com/google/uuid
   go get golang.org/x/crypto/bcrypt
   ```

4. Create a PostgreSQL database:
   ```
   createdb go_firebase
   ```

### Running the Application

1. Run database migrations:
   ```
   make migrate-up
   ```

2. Start the API server:
   ```
   make run-api
   ```

3. In a separate terminal, start the background worker:
   ```
   make run-worker
   ```

## API Endpoints

### Public Endpoints

- `POST /api/auth/signup` - Create a new user account
- `POST /api/auth/login` - Log in and get a session token

### Protected Endpoints

- `GET /api/protected/auth/me` - Get current user information
- `POST /api/protected/auth/logout` - Log out and invalidate session

## Development

### Creating a Migration

To create a new migration:
```
make create-migration
```

### Building the Project

To build all binaries:
```
make build
```

The binaries will be in the `build/bin` directory.

## License

This project is licensed under the MIT License - see the LICENSE file for details.