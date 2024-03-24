CREATE TABLE IF NOT EXISTS guilds (
    id INTEGER PRIMARY KEY,
    guild_snowflake_id VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS creator_channels (
    id INTEGER PRIMARY KEY,
    guild_id INTEGER,
    channel_snowflake_id VARCHAR(255) UNIQUE NOT NULL,
    user_limit INTEGER NOT NULL,

    FOREIGN KEY(guild_id) REFERENCES guilds(id)
);

CREATE TABLE IF NOT EXISTS temporary_voice_channels (
    id INTEGER PRIMARY KEY,
    guild_id INTEGER,
    channel_snowflake_id VARCHAR(255) UNIQUE NOT NULL,
    user_count INTEGER NOT NULL,
    owner_snowflake_id VARCHAR(255) NOT NULL,

    FOREIGN KEY(guild_id) REFERENCES guilds(id)
);
