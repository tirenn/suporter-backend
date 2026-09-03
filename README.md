# Suporter Backend Service

The Golang backend application powering real-time OBS Studio Browser Source overlays, user management, JWT authentication, and project alert event streaming.

---

## 🛠️ Tech Stack

- **Go**: `v1.22+`
- **Web Framework**: [Gin Framework](https://github.com/gin-gonic/gin)
- **ORM**: [GORM](https://gorm.io/) with PostgreSQL driver
- **Database**: PostgreSQL 14+
- **Migrations**: [Goose](https://github.com/pressly/goose/v3)
- **Configuration & Secrets**: [Doppler](https://www.doppler.com/) / [Viper](https://github.com/spf13/viper)
- **API Documentation**: [Swag / Swagger UI](https://github.com/swaggo/swag)

---

## 🔐 Doppler Setup (Backend Service)

1. Create a project named **`suporter-backend`** on [Doppler](https://doppler.com).
2. Import secrets from [`.env.example`](file:///c:/Users/Ryzen/Documents/Projects/suporter/backend/.env.example).
3. Setup and link locally:
```bash
doppler login
doppler setup --project suporter-backend --config dev
```
4. Run locally with Doppler:
```bash
# Run server
doppler run -- make run

# Run tests
doppler run -- make test

# Run migrations
doppler run -- make migrate-up
```

---

## 🤖 GitHub Actions CI/CD for Backend Repo

If this backend repository is hosted independently:
- **`ci.yml`**: Runs `1. Build Binary` ➔ `2. Unit Tests` on every PR and merge to `main`.
- **`deploy.yml`**: Runs `1. Build Binary` ➔ `2. Unit Tests` ➔ `3. Deploy to VPS on Tag` (`git tag v1.0.0 && git push origin v1.0.0`), retrieves secrets via Doppler, connects to your VPS over SSH, and restarts the backend container.

### Required GitHub Secrets for Backend:
- `DOPPLER_TOKEN`: Doppler Service Token for `suporter-backend` (Production config).
- `SSH_HOST`: VPS IP address / hostname.
- `SSH_USER`: SSH username on VPS (e.g. `root`).
- `SSH_KEY`: SSH private key.
- `SSH_PORT`: *(optional, default 22)*.
- `TARGET_DIR`: `/root/Projects/suporter-backend`.

---

## 🚀 Quick Execution Commands (Makefile)

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
