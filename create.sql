-- sqlite thing for foreign keys support
PRAGMA foreign_keys = ON;
-- Remove conflicting tables
DROP TABLE IF EXISTS cleni;
DROP TABLE IF EXISTS schuzky;
-- End of removing

CREATE TABLE cleni (
    id INTEGER PRIMARY KEY,
    discord_id TEXT NOT NULL,
    jmeno TEXT NOT NULL UNIQUE
);

CREATE TABLE schuzky (
    id INTEGER PRIMARY KEY,
    cleni_id INTEGER,
    nazev TEXT NOT NULL,
    kdy INTEGER NOT NULL, -- čas v UTC
    upozorneno INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY(cleni_id) REFERENCES cleni(id) ON DELETE CASCADE ON UPDATE CASCADE
);

