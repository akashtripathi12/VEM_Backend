# TBO Backend - Installation & Setup Guide

## Prerequisites

- **Go 1.25.5** or higher
- **Git** (already installed)

---

## Step 1: Install Go

### Windows Installation

1. Download Go from: https://go.dev/dl/
2. Download the Windows installer (`.msi` file)
3. Run the installer and follow the prompts
4. Verify installation:
   ```powershell
   go version
   ```

---

## Step 2: Install Dependencies

Navigate to your project directory and run:

```powershell
cd t:\TBO_BACKEND
go mod download
```

This will download all Fiber v2 dependencies.

---

## Step 3: Run the Server

```powershell
go run cmd/api/main.go
```

Expected output:
```
🚀 Starting TBO Backend [env: development]
✅ Database store initialized
✅ Repository initialized
✅ Middleware stack configured
✅ Routes registered
🌐 Server listening on :8080
```

---

## Step 4: Test the API

### Health Check
```powershell
curl http://localhost:8080/health
```

### Get Events
```powershell
curl http://localhost:8080/api/v1/events
```

### Create Event
```powershell
curl -X POST http://localhost:8080/api/v1/events -H "Content-Type: application/json" -d '{\"name\":\"Test Event\"}'
```

---

## Step 5: Build for Production

```powershell
go build -o tbo-backend.exe cmd/api/main.go
```

Run the executable:
```powershell
.\tbo-backend.exe
```

---

## Troubleshooting

### "go: command not found"
- Go is not installed or not in PATH
- Restart your terminal after installing Go
- Verify with: `go version`

### Port 8080 already in use
- Change port in `internal/config/config.go`
- Or kill the process using port 8080

### Module errors
- Run: `go mod tidy`
- Then: `go mod download`

---

## Next Steps

1. Implement actual database (replace MockStore)
2. Add JWT authentication logic
3. Add request validation
4. Write tests
5. Deploy to production

---

## Useful Commands

```powershell
# Run server
go run cmd/api/main.go

# Build executable
go build -o tbo-backend.exe cmd/api/main.go

# Run tests (when added)
go test ./...

# Format code
go fmt ./...

# Check for issues
go vet ./...

# Update dependencies
go mod tidy
```
