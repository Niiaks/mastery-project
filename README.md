# Item Api

A simple Go CRUD API designed for deployment on a VPS.

## Overview

This is a RESTful API built with Go, featuring user authentication, session management, and item CRUD operations with file upload support. The application follows a clean architecture pattern with separate layers for handlers, services, and repositories.

## Tech Stack

- **Language:** Go 1.25
- **Router:** [Chi](https://github.com/go-chi/chi) - Lightweight, idiomatic HTTP router
- **Database:** PostgreSQL with [pgx](https://github.com/jackc/pgx) driver
- **Migrations:** [golang-migrate](https://github.com/golang-migrate/migrate)
- **Validation:** [go-playground/validator](https://github.com/go-playground/validator)
- **Rate Limiting:** [go-chi/httprate](https://github.com/go-chi/httprate)
- **Password Hashing:** bcrypt

## Project Structure

```
mastery-project/
├── cmd/
│   └── mastery-project/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/               # Configuration management
│   ├── database/             # Database connection & migrations
│   │   └── migrations/       # SQL migration files
│   ├── handler/              # HTTP request handlers
│   ├── middleware/           # Authentication middleware
│   ├── model/                # Data models
│   ├── repository/           # Database operations
│   ├── router/               # Route definitions
│   ├── server/               # HTTP server setup
│   └── service/              # Business logic
```

## Features

- **User Authentication**

  - User registration with email validation
  - Secure login with bcrypt password hashing
  - Session-based authentication with HTTP-only cookies
  - CSRF protection with SameSite cookie policy

- **Items Management (CRUD)**

  - Create items with file upload (images)
  - Read single or all items
  - Update item details
  - Delete items with associated files

- **Security**
  - Rate limiting (10 requests/minute)
  - Session-based authentication
  - Secure file upload with MIME type validation
  - Allowed file types: JPEG, PNG (max 5MB)

## API Endpoints

### Public Routes

| Method | Endpoint                | Description       |
| ------ | ----------------------- | ----------------- |
| GET    | `/api/v1/health`        | Health check      |
| POST   | `/api/v1/auth/register` | Register new user |
| POST   | `/api/v1/auth/login`    | User login        |

### Protected Routes (Requires Authentication)

| Method | Endpoint             | Description     |
| ------ | -------------------- | --------------- |
| GET    | `/api/v1/items`      | Get all items   |
| POST   | `/api/v1/items`      | Create new item |
| GET    | `/api/v1/items/{id}` | Get item by ID  |
| PATCH  | `/api/v1/items/{id}` | Update item     |
| DELETE | `/api/v1/items/{id}` | Delete item     |

### Static Files

| Path         | Description          |
| ------------ | -------------------- |
| `/uploads/*` | Serve uploaded files |

## Environment Variables

Create a `.env` file in the project root:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASS=your_db_password
DB_NAME=your_db_name
SSL_MODE=disable

# Server Configuration
SERVER_PORT=8080
READ_TIMEOUT=15
WRITE_TIMEOUT=15
IDLE_TIMEOUT=60

# Environment
ENV=development
```

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Items Table

```sql
CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    file_path TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Sessions Table

```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Getting Started

### Prerequisites

- Go 1.25 or higher
- PostgreSQL 13+

### Installation

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd mastery-project
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Set up PostgreSQL database**

   ```bash
   createdb your_db_name
   ```

5. **Run the application**

   ```bash
   go run cmd/mastery-project/main.go
   ```

   Migrations will run automatically on startup.

## API Usage Examples

### Register a User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

### Create an Item (with file upload)

```bash
curl -X POST http://localhost:8080/api/v1/items \
  -b cookies.txt \
  -F "title=My Item" \
  -F "description=Item description" \
  -F "file=@/path/to/image.jpg"
```

### Get All Items

```bash
curl -X GET http://localhost:8080/api/v1/items \
  -b cookies.txt
```

### Update an Item

```bash
curl -X PATCH http://localhost:8080/api/v1/items/{item-id} \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "title": "Updated Title",
    "description": "Updated description"
  }'
```

### Delete an Item

```bash
curl -X DELETE http://localhost:8080/api/v1/items/{item-id} \
  -b cookies.txt
```
