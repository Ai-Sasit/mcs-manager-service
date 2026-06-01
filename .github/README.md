# GitHub Actions Deployment

This repository deploys MC Manage from `.github/workflows/deploy.yml`.

The workflow runs when code is pushed to the `deploy` branch. It uses SSH key based SCP/SSH through `appleboy/scp-action` and `appleboy/ssh-action`.

## Jobs

### Frontend

The `deploy-frontend` job:

1. Checks out the repository.
2. Installs Bun.
3. Runs `bun install` in `frontend`.
4. Writes `frontend/.env` with `VITE_API_BASE_URL` and optional `VITE_WS_BASE_URL`.
5. Runs `bun run build`.
6. Uploads `frontend/dist/*` to `/tmp/mc-manage-frontend/` on the VPS.
7. Copies the uploaded files into `/var/www/${DOMAIN_NAME}` with `rsync --delete`.
8. Sets web ownership to `www-data:www-data` and permissions to `755`.

### Database

The `deploy-database` job:

1. Checks out the repository.
2. Prepares `/tmp/mc-manage-database/` on the VPS.
3. Uploads the local `database/` directory.
4. Publishes the compose files into `/opt/mc-manage/database`.
5. Writes `/opt/mc-manage/database/.env`.
6. Runs `docker compose up -d --remove-orphans`.
7. Prints `docker compose ps` so the action log shows MongoDB status.

### Backend

The `deploy-backend` job depends on `deploy-database` and:

1. Checks out the repository.
2. Installs Go using `backend/go.mod`.
3. Builds a Linux amd64 binary named `mc-manage-backend`.
4. Uploads the binary to `/tmp/mc-manage/` on the VPS.
5. Installs the binary under `/opt/mc-manage/backend`.
6. Writes `/opt/mc-manage/backend/.env`.
7. Creates `/etc/systemd/system/mc-manage.service` if it does not already exist.
8. Restarts `mc-manage` and fails the deployment if the service is not active.

## Required GitHub Variables

Configure these in repository settings under **Actions variables**:

- `VPS_HOST`: VPS hostname or IP address.
- `VPS_USER`: SSH username.
- `DOMAIN_NAME`: frontend web root name, used as `/var/www/${DOMAIN_NAME}`.
- `VITE_API_BASE_URL`: frontend API base URL.

## Optional GitHub Variables

- `VPS_PORT`: SSH port. Defaults to `22`.
- `VITE_WS_BASE_URL`: optional websocket base URL. Use this when production websocket traffic should use a different origin than `VITE_API_BASE_URL`, for example `wss://example.com` or `wss://api.example.com`.
- `BACKEND_PORT`: backend service port. Defaults to `8080`.
- `DEBUG_MODE`: backend debug flag. Defaults to `false`.
- `ADMIN_USERNAME`: initial admin username. Defaults to `admin`.
- `MONGO_DB_NAME`: backend MongoDB database name. Defaults to `mc-manage`.
- `MONGO_INITDB_PORT`: exposed MongoDB port. Defaults to `27017`.
- `MONGO_INITDB_DATABASE`: MongoDB database created by the compose stack. Defaults to `mc-manage`.

## Required GitHub Secrets

Configure these in repository settings under **Actions secrets**:

- `VPS_SSH_KEY`: private SSH key for the VPS user.
- `ADMIN_PASSWORD`: initial admin password.
- `JWT_SECRET`: backend JWT signing secret.
- `MONGO_URI`: MongoDB connection string used by the backend.
- `MONGO_INITDB_ROOT_PASSWORD`: MongoDB root password used by the database compose stack.

## Server Expectations

The VPS user must be able to run the required `sudo` commands used by the workflow:

- create and write `/var/www/${DOMAIN_NAME}`
- create and write `/opt/mc-manage/database`
- create and write `/opt/mc-manage/backend`
- write `/etc/systemd/system/mc-manage.service` when the service is missing
- run `systemctl daemon-reload`, `enable`, `stop`, and `start`
- run `rsync`, `chmod`, and `chown`

The server should have `rsync`, `systemd`, Docker Compose, and nginx or another web server already configured to serve `/var/www/${DOMAIN_NAME}`.

## Production WebSockets

The frontend websocket paths are served under `/ws`, while HTTP APIs are served under `/api/v1`. In production, nginx or your reverse proxy must proxy both `/api/v1` and `/ws` to the same backend service port.

For nginx, the `/ws` location must preserve upgrade headers:

```nginx
location /ws/ {
    proxy_pass http://127.0.0.1:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 3600s;
    proxy_send_timeout 3600s;
}
```

Use the backend port from `BACKEND_PORT` if it is not `8080`. If the frontend is served over HTTPS, websocket URLs must use `wss://`, not `ws://`.

Set `VITE_WS_BASE_URL` only when the websocket origin is different from what the frontend can derive from `VITE_API_BASE_URL`. For same-domain deployments, `VITE_WS_BASE_URL=wss://your-domain.com` is usually enough.

## Manual Deployment Checklist

Before pushing to `deploy`:

1. Confirm all required variables and secrets are configured.
2. Confirm `VPS_USER` can SSH to the server with `VPS_SSH_KEY`.
3. Confirm the frontend domain points to the server.
4. Confirm the backend service port matches any reverse proxy configuration.
5. Confirm Docker Compose is available on the server for MongoDB.
6. Push the desired commit to the `deploy` branch.

After deployment:

1. Open the frontend domain and confirm the app loads.
2. Confirm frontend API calls reach `VITE_API_BASE_URL`.
3. Confirm websocket calls use `wss://<domain>/ws/...` in browser devtools.
4. Check MongoDB with `cd /opt/mc-manage/database && docker compose ps`.
5. Check the backend service with `systemctl status mc-manage`.
6. Check recent backend logs with `journalctl -u mc-manage -n 50` if the service does not become active.

## Troubleshooting

- If SSH fails, verify `VPS_HOST`, `VPS_USER`, `VPS_PORT`, and `VPS_SSH_KEY`.
- If frontend files do not update, check `/tmp/mc-manage-frontend/`, `/var/www/${DOMAIN_NAME}`, and web server permissions.
- If MongoDB does not start, check `/opt/mc-manage/database/.env`, `docker compose ps`, and `docker compose logs`.
- If backend deployment fails, check whether `mc-manage-backend` exists in `/tmp/mc-manage/` and whether `/opt/mc-manage/backend` is writable through `sudo`.
- If the service starts then exits, inspect `journalctl -u mc-manage -n 50`.
- If websocket requests fail with `502`, confirm the backend service is active, the proxy upstream port matches `BACKEND_PORT`, `/ws` includes the nginx upgrade headers above, TLS pages are using `wss://`, and the browser console websocket URL points at the production domain rather than `localhost` or the wrong API host.
