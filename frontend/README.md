# MC Manage Frontend

Vue 3 + Vite admin console for managing Minecraft server instances.

## Structure

- `src/views/auth` contains public authentication pages.
- `src/views/admin` contains routed admin console pages.
- `src/components/common` contains shared UI components.
- `src/components/servers` contains server-management UI pieces.
- `src/api/client.js` owns Axios setup, auth headers, timeout, and 401 redirects.
- `src/api/index.js` exposes endpoint methods used by stores and views.
- `src/plugins/index.js` registers app-wide Vue plugins.
- `src/assets/themes` contains design tokens, global styles, and Element Plus overrides.

Use `@/...` imports for anything under `src`.

## Scripts

```bash
bun install
bun run build
```

`npm` is not available in the current Codex desktop shell, so Bun is the verified local package manager for this workspace.
