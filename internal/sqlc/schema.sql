-- foreign keys must be disabled to drop tables with foreign keys
PRAGMA foreign_keys = OFF;

DROP TABLE IF EXISTS magic_links;
DROP TABLE IF EXISTS migrations;
DROP TABLE IF EXISTS paths;
DROP TABLE IF EXISTS users;

-- foreign keys must be enabled with every database connection
PRAGMA foreign_keys = ON;

CREATE TABLE magic_links
(
    link   TEXT      NOT NULL,
    handle TEXT      NOT NULL,
    crdttm TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (link),
    UNIQUE (handle)
);

-- paths to files needed by the server
CREATE TABLE paths
(
    name TEXT NOT NULL, -- absolute path to the assets directory
    path TEXT NOT NULL, -- absolute path to the templates directory
    PRIMARY KEY (name)
);

-- users stores information about the end-users (the players)
CREATE TABLE users
(
    id              INTEGER PRIMARY KEY,
    handle          TEXT NOT NULL,             -- forced to lowercase
    hashed_password TEXT NOT NULL,             -- hashed and hex-encoded
    clan            TEXT NOT NULL,             -- formatted as 0138
    magic           TEXT NOT NULL,             -- magic key, uuid
    enabled         TEXT NOT NULL DEFAULT 'N', -- Y or N
    UNIQUE (handle),
    UNIQUE (clan),
    UNIQUE (magic)
);

-- migrations stores information about the migration scripts that have
-- been applied to the database. makes some dangerous assumptions about
-- developers keeping their migrations in a predictable order.
CREATE TABLE migrations
(
    id     TEXT      NOT NULL,
    crdttm TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);
