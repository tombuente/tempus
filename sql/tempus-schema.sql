CREATE TABLE IF NOT EXISTS guilds (
    id INTEGER PRIMARY KEY,
    guild_id VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS creator_channels (
    id INTEGER PRIMARY KEY,
    guild_id INTEGER, -- References the primary key of `guilds`, not the snowflake id
    channel_id VARCHAR(255) UNIQUE NOT NULL,
    max_users INTEGER,

    FOREIGN KEY(guild_id) REFERENCES guilds(id)
)
