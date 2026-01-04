PRAGMA foreign_keys = ON;

-- =========================
-- films
-- =========================
CREATE TABLE films (
                       id INTEGER PRIMARY KEY,
                       imdb_id TEXT,
                       popularity REAL,
                       budget INTEGER,
                       revenue INTEGER,
                       title TEXT NOT NULL,
                       release_year INTEGER
);

CREATE INDEX idx_films_imdb_id
    ON films (imdb_id);

CREATE INDEX idx_films_release_year
    ON films (release_year);

-- =========================
-- actors
-- =========================
CREATE TABLE actors (
                        id INTEGER PRIMARY KEY,
                        name TEXT NOT NULL UNIQUE
);

-- =========================
-- genres
-- =========================
CREATE TABLE genres (
                        id INTEGER PRIMARY KEY,
                        name TEXT NOT NULL UNIQUE
);

-- =========================
-- films_actors (many-to-many)
-- =========================
CREATE TABLE films_actors (
                              film_id INTEGER NOT NULL,
                              actor_id INTEGER NOT NULL,

                              PRIMARY KEY (film_id, actor_id),

                              FOREIGN KEY (film_id)
                                  REFERENCES films (id)
                                  ON DELETE CASCADE
                                  ON UPDATE CASCADE,

                              FOREIGN KEY (actor_id)
                                  REFERENCES actors (id)
                                  ON DELETE CASCADE
                                  ON UPDATE CASCADE
);

CREATE INDEX idx_films_actors_opposite
    ON films_actors (actor_id, film_id);

-- =========================
-- films_genres (many-to-many)
-- =========================
CREATE TABLE films_genres (
                              film_id INTEGER NOT NULL,
                              genre_id INTEGER NOT NULL,

                              PRIMARY KEY (film_id, genre_id),

                              FOREIGN KEY (film_id)
                                  REFERENCES films (id)
                                  ON DELETE CASCADE
                                  ON UPDATE CASCADE,

                              FOREIGN KEY (genre_id)
                                  REFERENCES genres (id)
                                  ON DELETE CASCADE
                                  ON UPDATE CASCADE
);

CREATE INDEX idx_films_genres_opposite
    ON films_genres (genre_id, film_id);