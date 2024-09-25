-- sqlite thing
--PRAGMA foreign_keys = ON;
-- Remove conflicting tables
DROP TABLE IF EXISTS cleni;
DROP TABLE IF EXISTS schuzky;
-- End of removing

CREATE TABLE cleni (
    id INTEGER PRIMARY KEY,
    discord_id TEXT NOT NULL,
    jmeno TEXT NOT NULL
);

CREATE TABLE schuzky (
    id INTEGER PRIMARY KEY,
    cleni_id INTEGER,
    nazev TEXT NOT NULL,
    kdy TEXT NOT NULL,
    FOREIGN KEY(cleni_id) REFERENCES cleni(id)
);

