# [Project Name] - Standard Structure

This document defines the standard architecture and directory structure for the project. Use this as a reference when initializing new modules or services.

## 📁 Root Directory Layout

```text
[project-root]/
├── backend/            # Go backend service
├── frontend/           # Vue.js frontend (Vite)
├── database/           # Database migrations, seeds, and schemas
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

---

## 🎨 Frontend (Vue.js + Vite)

The frontend is structured for scalability and maintainability using Vue 3 and Pinia.

### Directory Structure: `frontend/src/`

- **`main.js`**: Initializes Vue, Pinia, Router, and global plugins.
- **`App.vue`**: Root component of the application.
- **`views/`**: Components representing full pages (routed).
- **`components/`**: Atomic and molecular UI components (Shared).
- **`layouts/`**: Structural templates (e.g., Sidebar + Header, Fullscreen).
- **`router/`**: Route definitions and navigation guards.
- **`stores/`**: Global state management (Pinia).
- **`plugins/`**: External library configurations (Axios, UI Frameworks).
- **`assets/`**: Images, fonts, and global style sheets.
- **`utils/`**: Reusable JS logic.
- **`constants/`**: API endpoints and static config tokens.

---

## 🛠️ Project Initialization Checklist

1. **Environment Config**: Copy `.env.example` to `.env` in both `frontend/` and `backend/`.
2. **Dependencies**:
   - Backend: Run `go mod tidy`.
   - Frontend: Run `npm install` or `bun install`.
3. **Database**: Apply migrations located in `database/`.
4. **Naming Standard**:
   - **Backend**: `PascalCase` for Exports, `camelCase` for internals.
   - **Frontend**: `PascalCase` for `.vue` files, `kebab-case` for assets.
5. **Consistency**: Ensure backend `models` match frontend `interfaces` or `types`.
