CREATE TABLE IF NOT EXISTS creator_channels (
    id         INTEGER PRIMARY KEY,
    guild_id   INTEGER NOT NULL,
    user_limit INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS temporary_voice_channels (
    id         INTEGER PRIMARY KEY,
    guild_id   INTEGER NOT NULL,
    user_count INTEGER NOT NULL,
    owner_id   INTEGER NOT NULL
);
