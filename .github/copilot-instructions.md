# MC Manage — Copilot Instructions

MC Manage is a Minecraft server management web app. The workspace has four roots:

| Folder | Stack |
|---|---|
| `backend/` | Go 1.25+, Fiber v3, MongoDB (mongo-driver/v2), JWT, WebSocket |
| `frontend/` | Vue 3, Vite, Element Plus, Pinia, Bun |
| `database/` | MongoDB via Docker Compose |
| `.github/` | GitHub Actions (deploys to VPS) |

---

## Backend

### Commands (run from `backend/`)
```sh
go run ./...          # start dev server on :8080
go build ./...        # compile
go run ./cmd/migrate-json-to-mongo  # migrate legacy JSON data to MongoDB
```

### Key conventions
- **All API routes** are prefixed `/api` and protected by `middleware.AuthRequired` (Bearer JWT). Public exceptions: `POST /api/auth/login`.
- **WebSocket routes** live at `/ws/...` — auth uses `?token=` query param, not a header.
- `services.AppState` holds in-memory server state. It is initialised in `main.go` and passed to controllers via `controllers.Init(state)`. Controllers retrieve it with `utils.GetState(c)`.
- **All API responses** use `interfaces.ApiResponse{Success bool, Data any, Message string}`.
- Server runtime files (JARs, plugins, backups, config) live on disk under `data/servers/{id}/`. MongoDB only stores metadata (`ServerConfig`).
- Java plugins → `plugins/` subdirectory; Bedrock add-ons → `addons/` subdirectory.
- **Path traversal** must be prevented on every file operation — see `utils/zip.go` and `FilesController.go` for the pattern.
- JWT claims expose `username` and `role`; Fiber stores them as `c.Locals("username")` / `c.Locals("role")`.

### Env vars
```
PORT           # default 8080
MONGO_URI      # e.g. mongodb://root:pw@localhost:27017/mc-manage?authSource=admin
MONGO_DB_NAME  # e.g. mc-manage
JWT_SECRET
```

### Key files
- [`src/routes.go`](../backend/src/routes.go) — route registration
- [`src/services/server_service.go`](../backend/src/services/server_service.go) — server lifecycle & download logic
- [`src/models/models.go`](../backend/src/models/models.go) — `ServerConfig`, `ServerEdition`, `ServerStatus`
- [`src/constants/constants.go`](../backend/src/constants/constants.go) — ports, RAM, player defaults

---

## Frontend

### Commands (run from `frontend/`)
```sh
bun install      # install deps — use Bun, NOT npm
bun run dev      # start Vite dev server
bun run build    # production build
```

> `npm` is not available in this environment. Always use `bun`.

### Key conventions
- Use `@/` alias for all `src/` imports.
- Element Plus components are **auto-imported** via `unplugin-vue-components` — do not add manual imports.
- Design tokens live in [`src/assets/themes/tokens.scss`](../frontend/src/assets/themes/tokens.scss); use CSS vars (`--color-primary`, `--color-border`, etc.) in styles.
- SCSS files can `@use` tokens directly (load path configured in `vite.config.js`).
- Auth token stored in `localStorage` as `mc_token`; `src/api/client.js` attaches it as a Bearer header and redirects on 401.
- WebSocket URL is derived from `VITE_API_BASE_URL` by swapping `http(s)` → `ws(s)` — see [`src/constants/index.js`](../frontend/src/constants/index.js).
- New views go in `src/views/admin/` (authenticated) or `src/views/auth/` (public). Register them in [`src/router/index.js`](../frontend/src/router/index.js).
- New API calls go in [`src/api/index.js`](../frontend/src/api/index.js).

### Env vars
```
VITE_API_BASE_URL   # e.g. https://api.example.com/api
```

---

## Database

See [`database/README.md`](../database/README.md) for Docker Compose setup and required `.env` values.

---

## CI/CD

[`.github/workflows/deploy.yml`](workflows/deploy.yml) deploys frontend, backend, and database independently to a VPS on push to `main`. Required GitHub vars: `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`, `VPS_PORT`, `VITE_API_BASE_URL`, `DOMAIN_NAME`.
