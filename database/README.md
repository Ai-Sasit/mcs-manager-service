# MC Manage MongoDB

MongoDB stores MC Manage application metadata: servers, users, schedules, settings, and audit logs.

Minecraft runtime files, uploaded plugins, backups, and backend logs remain on disk.

## Local Start

Create `database/.env`:

```sh
MONGO_INITDB_PORT=27017
MONGO_INITDB_ROOT_PASSWORD=change-me
MONGO_INITDB_DATABASE=mc-manage
```

Start MongoDB:

```sh
cd database
docker compose up -d
```

Backend `.env` should include:

```sh
MONGO_DB_NAME=mc-manage
MONGO_URI=mongodb://root:change-me@localhost:27017/mc-manage?authSource=admin
```

## Import Existing JSON Data

From `backend/`, run:

```sh
go run ./cmd/migrate-json-to-mongo
```

Optionally pass a custom data directory:

```sh
go run ./cmd/migrate-json-to-mongo ./data
```
