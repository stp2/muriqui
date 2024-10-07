-- sqlite thing for foreign keys support
PRAGMA foreign_keys = ON;
-- Remove conflicting tables
DROP TABLE IF EXISTS cleni;
DROP TABLE IF EXISTS porada;
DROP TABLE IF EXISTS schuzky;
-- End of removing

CREATE TABLE cleni (
    id INTEGER PRIMARY KEY,
    jmeno TEXT NOT NULL UNIQUE,
    discord_id TEXT NOT NULL UNIQUE
);

CREATE TABLE schuzky (
    id INTEGER PRIMARY KEY,
    nazev TEXT NOT NULL,
    kdy INTEGER NOT NULL, -- čas v UTC
    upozorneno BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE porada (
    cleni_id INTEGER NOT NULL,
    schuzky_id INTEGER NOT NULL,
    zprava_id TEXT NOT NULL DEFAULT '',
    PRIMARY KEY(cleni_id, schuzky_id),
    FOREIGN KEY(cleni_id) REFERENCES cleni(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY(schuzky_id) REFERENCES schuzky(id) ON DELETE CASCADE ON UPDATE CASCADE
);