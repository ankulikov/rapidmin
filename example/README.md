# Example App

This folder contains a minimal runnable example using SQLite.

## Setup

1) Run the example app (migrates and starts the backend):

```sh
cd example
go run .
```

2) Open the app:

```sh
open http://localhost:8080
```

The example app runs the SQLite migration from `example/db/schema.sql` and
`example/db/data.sql`, creates `example/rapidmin.db`, and then starts the
backend with `example/config.yaml`. The frontend is served from the embedded
HTML in `backend/internal/server/web/index.html`.

## Sample Data

The dataset contains several hundred rows across:

- `films` (hundreds of film records)
- `actors` (hundreds of actor records)
- `film_actors` (hundreds of relations)
