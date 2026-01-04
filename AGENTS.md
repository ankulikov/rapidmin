# Requirements

## Goal
Build a simple web app that renders pages and widgets from a YAML config and fetches widget data from backend APIs.

## High-Level Flow
1. Backend loads config file once at startup and registers data providers.
2. User opens the app in the browser.
3. Frontend (SPA) loads the HTML shell, fetches config, resolves current page by URL, builds widgets, and fetches data with filters.

## Backend Requirements
- Language: Go.
- Single production binary.
- Endpoints:
  - `GET /` serves minified HTML shell.
  - `GET /api/config` returns config JSON 
  - `GET /api/widgets/:id` returns widget data with filters from query params.
  - Cursor pagination uses `offset` as the cursor value when `provider.sql.pagination` is set.
  - Widget data response includes `next_cursor` when pagination is enabled.
  - Widget data response includes `has_more` when a limit is provided.
  - Default `limit` is 50 when not specified.
  - Filter query format: `filter_name[.operator]=value` with repeated key/value pairs for multi-values or between boundaries.
  - Operator tokens: `gt`, `lt`, `contains`, `in` (for multi choice filter), `between`, `before`, `after`. Equality is implicit when no operator is provided.
- Data providers:
  - Start with `db` provider using `sqlx`.
  - Support SQLite and Postgres.
- No auth for now.
- No hot-reload of config.

## API Examples
- `GET /api/config`

```http
HTTP/1.1 200 OK
Content-Type: application/json
```

```json
{
  "title": "NoCode MVP",
  "menu": [
    { "title": "Users", "page": "users" },
    { "title": "Reports", "href": "/reports" }
  ],
  "pages": [
    {
      "slug": "users",
      "title": "Users",
      "widgets": [
        {
          "id": "users_table",
          "title": "Users",
          "type": "table",
          "table": {
            "columns": [
              { "id": "id", "title": "ID" },
              {
                "id": "name",
                "title": "Name",
                "render": {
                  "type": "link",
                  "text": "{{name}}",
                  "url": "/users/{{id}}"
                }
              },
              {
                "id": "email",
                "title": "Email",
                "render": {
                  "type": "link",
                  "text": "{{email}}",
                  "url": "mailto:{{email}}",
                  "external": true
                }
              }
            ]
          }
        }
      ]
    }
  ]
}
```

- `GET /api/widgets/users_table?limit=25&offset=10&name.contains=ann&age.gt=18`

```json
{
  "data": [
    { "id": 1, "name": "Ann", "email": "ann@example.com" },
    { "id": 2, "name": "Anna", "email": "anna@example.com" }
  ],
  "total": 2,
  "next_cursor": "2",
  "has_more": true
}
```

- `GET /api/widgets/users_table?created.between=2024-01-01&created.between=2024-12-31`

- `GET /api/widgets/users_table?tags=vip&tags=active`

## Frontend Requirements
- Language: TypeScript + React.
- SPA with client-side routing based on `pages[].slug`.
- Builds UI from config and renders widgets.
- Widget data requests include filters from URL and widget state.
- No localization layer; labels are rendered as-is.
- Dev mode must include a stub server that mimics backend APIs.
- Release build should inline all JS/CSS into a single minified HTML file used by the backend shell.

## Config Notes (YAML)
- Top-level: `title`, `providers`, `menu[]`, `pages[]`.
- `providers` is a map keyed by provider name; each value includes a typed block (`sql`) with `driver`/`dsn`.
- Provider config values can use `{{env.VAR_NAME}}` to read from environment variables during config load.
- Menu items may link by `page` (slug) or `href` (external link).
- Pages have `slug`, `title`, `widgets[]`.
- Widgets include `id`, `title`, `type`, `provider`, and `table` display config.
- Provider SQL supports `query`, `bindings` (e.g., `query.limit`), and optional `pagination` with `column` and `order`.
- `table.columns` supports either a string column name or an object with `id`, optional `title`, and optional `render`.
- `render` supports `type: link`, `text` and `url` templates (e.g., `{{id}}`), and `external: true` for external links.

## Project structure

```
backend/
  cmd/
  internal/
    handlers/
    models/
    providers/
      sql/
  go.mod
frontend/
  package.json
  .env.example
```

## Frontend Dev Env

- `VITE_API_PROXY` sets the Vite dev proxy target for `/api` (default: `http://localhost:4173`).
