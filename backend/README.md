# MC Manage Backend

A Go-based backend server for managing Minecraft servers with a web interface. This application provides REST API endpoints and WebSocket connections for real-time server management, monitoring, and control.

## Features

- **Server Management**: Create, start, stop, restart, and kill Minecraft servers
- **Real-time Monitoring**: WebSocket connections for server logs and terminal access
- **Plugin Management**: Upload, list, and delete server plugins
- **File Management**: Read and write server configuration files
- **Backup System**: Create and manage server backups
- **Scheduling**: Schedule automated tasks (backups, restarts, etc.)
- **User Management**: Multi-user authentication with role-based access control
- **System Monitoring**: View system information and port usage
- **Audit Logging**: Track all administrative actions

## Tech Stack

- **Go 1.25.5** - Backend programming language
- **Fiber v3** - High-performance web framework
- **MongoDB** - Database for storing server configurations, users, and logs
- **JWT** - JSON Web Tokens for authentication
- **WebSocket** - Real-time communication for logs and terminal
- **Cron** - Scheduled task execution

## Project Structure

```
backend/
├── cmd/                    # Command-line utilities
│   ├── create-user/       # User creation utility
│   └── migrate-json-to-mongo/ # Data migration tool
├── data/                  # Data files and server configurations
├── logs/                  # Application logs
├── src/                   # Source code
│   ├── constants/        # Application constants
│   ├── controllers/      # HTTP request handlers
│   ├── interfaces/       # Interface definitions
│   ├── middlewares/      # HTTP middleware
│   ├── models/           # Data models
│   ├── services/         # Business logic
│   └── utils/            # Utility functions
├── tmp/                  # Temporary build files
├── main.go              # Application entry point
├── routes.go            # Route definitions
├── go.mod              # Go module dependencies
└── .env.example        # Environment variables template
```

## Getting Started

### Prerequisites

- Go 1.25.5 or later
- MongoDB (running locally or accessible)
- Git

### Installation

1. Clone the repository:

   ```bash
   git clone <repository-url>
   cd mc-manage/backend
   ```

2. Install dependencies:

   ```bash
   go mod download
   ```

3. Copy the environment file and configure it:

   ```bash
   cp .env.example .env
   ```

4. Edit `.env` with your configuration:

   ```env
   # Application environment (development, production, etc.)
   APP_ENV=development

   # Application configuration
   VERSION=1.0.0
   PORT=6020
   DEBUG_MODE=true

   # JWT Configuration
   JWT_SECRET=your-secret-key-change-this
   JWT_EXPIRATION_HOURS=24
   JWT_ISSUER=mc-manage
   JWT_ALGORITHM=HS256

   # Database configuration
   MONGO_DB_NAME=mc-manage
   MONGO_URI=mongodb://localhost:27017/mc-manage
   ```

5. Start MongoDB (if not already running):

   ```bash
   # On Windows with MongoDB installed as a service
   net start MongoDB

   # Or using Docker
   docker run -d -p 27017:27017 --name mongodb mongo:latest
   ```

6. Create an initial admin user:

   ```bash
   # Using environment variables
   set CREATE_USERNAME=admin
   set CREATE_PASSWORD=admin123
   set CREATE_ROLE=admin
   go run ./cmd/create-user

   # Or using command-line arguments
   go run ./cmd/create-user admin admin123 admin
   ```

7. Run the application:

   ```bash
   go run main.go
   ```

   The server will start on `http://127.0.0.1:6020` (or the port specified in `.env`).

### Development

For development with hot reloading, you can use Air:

```bash
# Install Air if not already installed
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

## API Documentation

### Authentication

- `POST /api/v1/auth/login` - Login with username/password
- `GET /api/v1/auth/me` - Get current user info (protected)
- `POST /api/v1/auth/logout` - Logout (protected)
- `PUT /api/v1/auth/me/password` - Change password (protected)

### Server Management

- `GET /api/v1/servers` - List all servers (protected)
- `POST /api/v1/servers` - Create a new server (protected)
- `GET /api/v1/servers/:id` - Get server details (protected)
- `DELETE /api/v1/servers/:id` - Delete a server (protected)
- `POST /api/v1/servers/:id/start` - Start a server (protected)
- `POST /api/v1/servers/:id/stop` - Stop a server (protected)
- `POST /api/v1/servers/:id/restart` - Restart a server (protected)
- `POST /api/v1/servers/:id/kill` - Force kill a server (protected)

### WebSocket Endpoints

- `GET /ws/servers/:id/logs` - Stream server logs (requires token query param)
- `GET /ws/servers/:id/terminal` - Server terminal access (requires token query param)
- `GET /ws/server-setup/:job_id` - Stream server setup progress (requires token query param)
- `GET /ws/backend-logs` - Stream backend application logs (requires token query param)

Log and terminal streams support resumable delivery with optional `stream_id`
and `last_event_id` query parameters. Each output envelope contains `type`,
`data`, `ts`, `stream_id`, and `event_id`. A client reconnects using the last
processed cursor; `stream_gap` reports an expired 10,000-event replay window,
and `stream_reset` reports a new backend stream. Slow consumers are closed with
code `1013` so they can resume, while invalid or revoked sessions close with
code `4401` and must not reconnect without signing in again.

Production reverse proxies must forward websocket upgrade traffic for `/ws` to this backend service. For nginx, configure `proxy_http_version 1.1`, `proxy_set_header Upgrade $http_upgrade`, and `proxy_set_header Connection "upgrade"` on the `/ws/` location. A `502` from the browser usually means nginx cannot reach the backend upstream or the `/ws` location is missing upgrade routing.

### Plugin Management

- `GET /api/v1/servers/:id/plugins` - List server plugins (protected)
- `POST /api/v1/servers/:id/plugins` - Upload a plugin (protected)
- `DELETE /api/v1/servers/:id/plugins/:name` - Delete a plugin (protected)

### File Management

- `GET /api/v1/servers/:id/files` - List server files (protected)
- `GET /api/v1/servers/:id/files/read` - Read a server file (protected)
- `POST /api/v1/servers/:id/files/write` - Write to a server file (protected)

### Backup Management

- `GET /api/v1/servers/:id/backups` - List server backups (protected)
- `POST /api/v1/servers/:id/backups` - Create a backup (protected)
- `DELETE /api/v1/servers/:id/backups/:name` - Delete a backup (protected)

### Scheduling

- `GET /api/v1/schedules` - List scheduled tasks (protected)
- `POST /api/v1/schedules` - Create a scheduled task (protected)
- `PUT /api/v1/schedules/:id/toggle` - Toggle a schedule (protected)
- `DELETE /api/v1/schedules/:id` - Delete a schedule (protected)

### User Management

- `GET /api/v1/users` - List all users (admin only)
- `POST /api/v1/users` - Create a new user (admin only)
- `PUT /api/v1/users/:id` - Update a user (admin only)
- `DELETE /api/v1/users/:id` - Delete a user (admin only)

### System

- `GET /api/v1/system/info` - Get system information (protected)
- `GET /api/v1/system/port-lookup` - Look up port usage (protected)
- `POST /api/v1/system/kill-pid` - Kill a process by PID (protected)
- `GET /api/v1/audit-logs` - Get audit logs (protected)
- `GET /api/v1/backend-logs` - List backend log files (protected)
- `GET /api/v1/backend-logs/file` - Get a specific backend log file (protected)

### Settings

- `GET /api/v1/settings` - Get application settings (protected)
- `PUT /api/v1/settings` - Update application settings (admin only)

## Database Schema

### Users Collection

```json
{
  "id": "uuid",
  "username": "string",
  "password_hash": "string",
  "role": "admin|viewer",
  "created_at": "ISO datetime"
}
```

### Servers Collection

```json
{
  "id": "uuid",
  "name": "string",
  "type": "java|bedrock",
  "version": "string",
  "port": "number",
  "path": "string",
  "status": "stopped|starting|running|stopping",
  "created_at": "ISO datetime",
  "updated_at": "ISO datetime"
}
```

## Security

- **JWT Authentication**: All protected endpoints require a valid JWT token
- **Password Hashing**: Passwords are hashed using bcrypt
- **CORS**: Configured to allow cross-origin requests (adjust for production)
- **Input Validation**: All user input is validated using go-playground/validator
- **Role-Based Access Control**: Different permissions for admin and viewer roles

## Deployment

### Building for Production

1. Build the binary:

   ```bash
   go build -o mc-manage-backend.exe main.go
   ```

2. Set production environment variables:

   ```env
   APP_ENV=production
   DEBUG_MODE=false
   JWT_SECRET=<strong-random-secret>
   MONGO_URI=<production-mongodb-uri>
   ```

3. Run the binary:
   ```bash
   mc-manage-backend.exe
   ```

### Docker Deployment

Create a `Dockerfile`:

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o mc-manage-backend main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mc-manage-backend .
COPY --from=builder /app/.env.example .env
EXPOSE 6020
CMD ["./mc-manage-backend"]
```

Build and run:

```bash
docker build -t mc-manage-backend .
docker run -p 6020:6020 --env-file .env mc-manage-backend
```

## Troubleshooting

### Common Issues

1. **MongoDB Connection Failed**
   - Ensure MongoDB is running: `mongod --version`
   - Check connection string in `.env`
   - Verify network connectivity to MongoDB host

2. **Port Already in Use**
   - Change `PORT` in `.env` file
   - Check for other processes using the port: `netstat -ano | findstr :6020`

3. **JWT Authentication Issues**
   - Verify `JWT_SECRET` is set in `.env`
   - Check token expiration time
   - Ensure token is included in `Authorization: Bearer <token>` header

4. **User Creation Fails**
   - Check MongoDB connection
   - Verify username doesn't already exist
   - Ensure password meets minimum length (8 characters)

### Logs

Application logs are stored in the `logs/` directory:

- `logs/YYYY-MM-DD-mc-manage.log` - Application logs
- `logs/YYYY-MM-DD.log` - General logs

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## License

[Add your license here]

## Support

For issues and feature requests, please use the GitHub issue tracker.
