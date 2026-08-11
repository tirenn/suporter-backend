# Suporter Backend Service

The Golang backend application powering real-time OBS Studio Browser Source overlays, user management, JWT authentication, and project alert event streaming.

---

## 🛠️ Tech Stack

- **Go**: `v1.22+`
- **Web Framework**: [Gin Framework](https://github.com/gin-gonic/gin)
- **ORM**: [GORM](https://gorm.io/) with PostgreSQL driver
- **Database**: PostgreSQL 14+
- **Migrations**: [Goose](https://github.com/pressly/goose/v3)
- **Configuration**: [Viper](https://github.com/spf13/viper) (`.env` file)
- **API Documentation**: [Swag / Swagger UI](https://github.com/swaggo/swag)

---

## ⚙️ Environment Configuration (`.env`)

```env
PORT=8080

# PostgreSQL Settings
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=suporter
DB_SSLMODE=disable

# JWT Security
JWT_SECRET=suporter-super-secret-jwt-key-2026
JWT_EXPIRY_HOURS=24
```

---

## 🚀 Quick Execution Commands

```bash
# 1. Run migrations
make migrate-up

# 2. Regenerate Swagger API specs
make swagger

# 3. Start server
make run

# 4. Run tests
make test
```

---

## 📚 Key Endpoints

- **Swagger UI**: `GET http://localhost:8080/swagger/index.html`
- **Dashboard**: `GET http://localhost:8080/dashboard`
- **Register**: `POST /api/v1/auth/register`
- **Login**: `POST /api/v1/auth/login`
- **Create Project**: `POST /api/v1/projects`
- **OBS Overlay Widget**: `GET /overlay/:project_id`
- **Trigger Alert**: `POST /api/v1/projects/:project_id/alert`
