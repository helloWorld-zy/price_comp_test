# Quickstart: 邮轮航次各房型价格统计对比工具

**Feature Branch**: `001-cruise-price-compare`  
**Date**: 2026-01-22

## Prerequisites

### System Requirements

- Go 1.25+
- Node.js 20+ (for frontend)
- MariaDB 12
- Ollama (local LLM)
- protoc (Protocol Buffers compiler)

### Install Dependencies

```bash
# Backend dependencies
go mod download

# Frontend dependencies
cd web && npm install

# Protobuf tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Database Setup

```bash
# Create database
mysql -u root -p -e "CREATE DATABASE cruise_price CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# Run migrations
go run cmd/migrate/main.go up
```

### Ollama Setup

```bash
# Install Ollama (see https://ollama.ai/)
# Pull recommended model
ollama pull llama3.2

# Verify Ollama is running
curl http://localhost:11434/api/tags
```

---

## Quick Start

### 1. Generate Protobuf Code

```bash
# Generate Go code from proto files
make proto-gen

# Or manually:
protoc --go_out=api/gen/go --go_opt=paths=source_relative \
       --go-grpc_out=api/gen/go --go-grpc_opt=paths=source_relative \
       api/proto/*.proto
```

### 2. Configure Environment

Create `.env` file:

```env
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=cruise_price

# Server
SERVER_PORT=8080
JWT_SECRET=your-jwt-secret-key

# Ollama
OLLAMA_HOST=http://localhost:11434
OLLAMA_MODEL=llama3.2
```

### 3. Start Backend

```bash
# Start API server
go run cmd/server/main.go

# Start worker (in another terminal)
go run cmd/worker/main.go
```

### 4. Start Frontend

```bash
cd web
npm run dev
```

Access the application at `http://localhost:5173`

---

## Development Commands

### Backend

```bash
# Run tests
go test ./...

# Run with hot reload (using air)
air

# Lint
golangci-lint run

# Generate mocks
go generate ./...
```

### Frontend

```bash
# Development
npm run dev

# Build
npm run build

# Lint
npm run lint

# Type check
npm run type-check
```

### Protobuf

```bash
# Generate all proto code
make proto-gen

# Validate proto files
make proto-lint
```

---

## API Overview

### Authentication

```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/x-protobuf" \
  -d '{"username":"admin","password":"admin123"}'

# Use token
curl -X GET http://localhost:8080/api/v1/user/me \
  -H "Authorization: Bearer <token>"
```

### Key Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/auth/login` | POST | 用户登录 |
| `/api/v1/catalog/cruise-lines` | GET/POST | 邮轮公司管理 |
| `/api/v1/catalog/ships` | GET/POST | 邮轮管理 |
| `/api/v1/catalog/sailings` | GET/POST | 航次管理 |
| `/api/v1/catalog/cabin-types` | GET/POST | 房型管理 |
| `/api/v1/quotes` | GET/POST | 报价管理 |
| `/api/v1/quotes/comparison` | GET | 价格对比 |
| `/api/v1/quotes/trend` | GET | 价格趋势 |
| `/api/v1/import/upload` | POST | 文件上传 |
| `/api/v1/import/text` | POST | 文本提交 |

---

## Default Users

| Username | Password | Role | Description |
|----------|----------|------|-------------|
| `admin` | `admin123` | ADMIN | 管理员账号 |
| `vendor1` | `vendor123` | VENDOR | 测试供应商账号 |

---

## Directory Structure

```text
.
├── api/
│   ├── proto/           # Protobuf definitions
│   └── gen/             # Generated code
├── cmd/
│   ├── server/          # API server entry
│   └── worker/          # Job worker entry
├── internal/
│   ├── app/             # Application bootstrap
│   ├── auth/            # Authentication
│   ├── domain/          # Domain models
│   ├── service/         # Business logic
│   ├── repo/            # Data access
│   ├── jobs/            # Async jobs
│   ├── parsers/         # File parsers
│   ├── llm/             # Ollama client
│   ├── transport/http/  # HTTP handlers
│   └── obs/             # Observability
├── migrations/          # DB migrations
├── web/                 # Vue frontend
└── docs/                # Documentation
```

---

## Common Tasks

### Add New Cruise Line (Admin)

1. Login as admin
2. Navigate to "基础数据" > "邮轮公司"
3. Click "新增"
4. Fill in company name and aliases
5. Save

### Submit Price Quote (Vendor)

1. Login as vendor user
2. Navigate to "报价录入"
3. Option A: Upload PDF/Word file
   - System extracts and identifies data via LLM
   - Review and confirm recognition results
4. Option B: Manual entry
   - Select sailing and cabin type
   - Enter price and details
5. Submit

### View Price Comparison

1. Navigate to "价格对比"
2. Select sailing (by cruise line, ship, date)
3. View comparison table
4. Click cabin type to see trend chart
5. Export to Excel/CSV if needed

---

## Troubleshooting

### Ollama Connection Failed

```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Restart Ollama
ollama serve
```

### Database Connection Failed

```bash
# Check MariaDB status
systemctl status mariadb

# Verify connection
mysql -u root -p -e "SHOW DATABASES;"
```

### PDF Extraction Issues

- Ensure the PDF is text-based (not scanned image)
- Check file size limit (default 10MB)
- View import job logs for detailed errors
