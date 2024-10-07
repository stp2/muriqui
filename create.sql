-- sqlite thing for foreign keys support
PRAGMA foreign_keys = ON;
-- Remove conflicting tables
DROP TABLE IF EXISTS cleni CASCADE;
DROP TABLE IF EXISTS porada CASCADE;
DROP TABLE IF EXISTS schuzky CASCADE;
-- End of removing

CREATE TABLE cleni (
    id INTEGER PRIMARY KEY,
    discord_id TEXT NOT NULL UNIQUE,
    jmeno TEXT NOT NULL UNIQUE
);

CREATE TABLE schuzky (
    id INTEGER PRIMARY KEY,
    nazev TEXT NOT NULL,
    kdy INTEGER NOT NULL, -- čas v UTC
    upozorneno BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE porada (
    cleni_id INTEGER PRIMARY KEY NOT NULL,
    schuzky_id INTEGER PRIMARY KEY NOT NULL,
    zprava_id TEXT NOT NULL UNIQUE DEFAULT '',
    FOREIGN KEY(cleni_id, schuzky_id) REFERENCES cleni(id), schuzky(id) ON DELETE CASCADE ON UPDATE CASCADE
);