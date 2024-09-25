-- sqlite thing
--PRAGMA foreign_keys = ON;
-- Remove conflicting tables
DROP TABLE IF EXISTS cleni;
DROP TABLE IF EXISTS schuzky;
-- End of removing

CREATE TABLE cleni (
    id SERIAL NOT NULL PRIMARY KEY,
    discord_id VARCHAR(256) NOT NULL,
    jmeno VARCHAR(256) NOT NULL
);

CREATE TABLE schuzky (
    id SERIAL NOT NULL PRIMARY KEY,
    cleni_id INTEGER,
    nazev VARCHAR(256) NOT NULL,
    kdy DATETIME NOT NULL,
    FOREIGN KEY(cleni_id) REFERENCES cleni(id)
);

