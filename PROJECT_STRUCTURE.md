# [Project Name] - Standard Structure

This document defines the standard architecture and directory structure for the project. Use this as a reference when initializing new modules or services.

## 📁 Root Directory Layout

```text
[project-root]/
├── backend/            # Go backend service
├── frontend/           # Vue.js frontend (Vite)
├── database/           # MongoDB Docker Compose and database docs
├── docs/               # Technical documentation and API specs
├── .gitignore          # Git ignore rules
└── PROJECT_STRUCTURE.md# This template
```

---

## 🚀 Backend (Go)

The backend follows a **Service-Controller** pattern to ensure separation of concerns.

### Directory Structure: `backend/src/`

- **`main.go`**: App entry point. Initializes config, DB, and the server.
- **`routes.go`**: Centralized route definitions.
- **`controllers/`**: HTTP Request/Response handling. *No business logic here.*
- **`services/`**: Core business logic and database interactions.
- **`models/`**: Database entity definitions and DTOs (Data Transfer Objects).
- **`interfaces/`**: Contracts for services to enable easy mocking in tests.
- **`middlewares/`**: Logic that runs before/after controllers (Auth, Logging, CORS).
- **`utils/`**: Independent utility functions (Time, Crypto, Strings).
- **`constants/`**: Fixed application-wide values (Error codes, Status strings).

### Persistence

- **MongoDB** stores application metadata: servers, users, schedules, settings, and audit logs.
- **Filesystem** stores Minecraft runtime assets: server directories, plugins, backups, config files, and backend log files.

---

## 🎨 Frontend (Vue.js + Vite)

The frontend follows the same organization style as the Room SOL frontend: routed pages are grouped by product area, reusable UI is grouped by domain, app-level setup lives in plugins, and global styling is split into theme modules.

### Directory Structure: `frontend/src/`

- **`main.js`**: Initializes Vue, Pinia, Router, and global plugins.
- **`App.vue`**: Root component of the application.
- **`views/auth/`**: Public authentication pages.
- **`views/admin/`**: Admin console pages for dashboard, servers, files, plugins, logs, users, port monitor, and settings.
- **`components/common/`**: Shared UI components used across feature areas.
- **`components/servers/`**: Server-management components such as cards, modals, and plugin upload controls.
- **`layouts/`**: Structural templates (e.g., Sidebar + Header, Fullscreen).
- **`router/`**: Route definitions and navigation guards.
- **`stores/`**: Global state management (Pinia).
- **`plugins/`**: App-level plugin registration such as Pinia, Router, and Element Plus.
- **`api/client.js`**: Axios client setup, auth header injection, timeout, and 401 handling.
- **`api/index.js`**: Endpoint-specific API methods used by views and stores.
- **`assets/themes/`**: Theme tokens, global styles, and Element Plus overrides.
- **`utils/`**: Reusable JS logic.
- **`constants/`**: API endpoints and static config tokens.

Frontend source imports should use the `@` alias for `frontend/src` (for example, `@/api`, `@/stores/auth`, or `@/views/admin/DashboardView.vue`).

---

## 🛠️ Project Initialization Checklist

1. **Environment Config**: Copy `.env.example` to `.env` in both `frontend/` and `backend/`.
2. **Dependencies**:
   - Backend: Run `go mod tidy`.
   - Frontend: Run `npm install` or `bun install`.
3. **Database**: Start MongoDB with `database/docker-compose.yml` or provide `MONGO_URI`.
4. **Naming Standard**:
   - **Backend**: `PascalCase` for Exports, `camelCase` for internals.
   - **Frontend**: `PascalCase` for `.vue` files, `kebab-case` for assets.
5. **Consistency**: Ensure backend `models` match frontend `interfaces` or `types`.
