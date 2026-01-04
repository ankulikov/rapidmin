# Rapidmin

Rapidmin is a small admin-style web app that renders pages and widgets from a YAML config, then fetches widget data from backend APIs. The backend is a single Go binary with a SQL provider, while the frontend is a Vite + React SPA.

## Project layout
- `backend/` Go server, config loader, providers, and integration tests.
- `frontend/` React SPA, API client, and mock stub server for development.
- `AGENTS.md` project requirements and API contracts.

## Run locally
Backend (serves API + HTML shell):
```sh
cd backend
go test ./internal/server -v
go run ./cmd/server
```

Frontend (dev server with proxy):
```sh
cd frontend
npm install
npm run dev
```

Frontend release build (single inlined HTML for backend):
```sh
cd frontend
npm run build:release
```
This overwrites `backend/internal/server/web/index.html` with the inlined, minified build.

Frontend stub API (for UI work without Go backend):
```sh
cd frontend
npm run stub
```

Set `VITE_API_PROXY` in `frontend/.env` to point the dev proxy to your backend (defaults to `http://localhost:4173`).

## Config basics
The YAML config defines menu pages and widgets. Example table columns and link rendering:
```yaml
table:
  columns:
    - id: name
      title: "Name"
      render:
        type: link
        text: "{{name}}"
        url: "/users/{{id}}"
```

Database provider config lives at `providers.db`:
```yaml
providers:
  db:
    sql:
      driver: sqlite3
      dsn: data.db
```
`driver`/`dsn` can use `{{env.VAR_NAME}}` to resolve values from environment variables at load time.

`render.type: link` supports:
- `text`: template for label.
- `url`: template for href.
- `external: true` to open in a new tab.

Templates replace `{{column}}` with row values.

SQL providers can declare type hints for filter targets:
```yaml
provider:
  name: db
  sql:
    types:
      created_at: date
      age: int
```

Filters use `target` to point at the provider field:
```yaml
filters:
  - id: name
    title: "Name contains"
    type: text
    target: name
    mode: contains
```

## API summary
- `GET /api/config` returns config JSON (without provider details).
- `GET /api/widgets/:id` returns widget data.

Filtering uses query params in the format `filter_name[.operator]=value`:
- `age.gt=10`
- `created.between=2024-01-01&created.between=2024-01-31`
- `tags=vip&tags=active`
Equality is implicit when no operator is provided.

Cursor pagination uses `offset` as the cursor value. Response includes `next_cursor` and `has_more`. Default limit is 50.
