package bot

type guild struct {
	ID               int64  `db:"id"`
	GuildSnowflakeID string `db:"guild_snowflake_id"`
}

type guildParams struct {
	guildSnowflakeID string
}

type guildFilter struct {
	id               int64
	guildSnowflakeID string
	limit            uint64
}

type creatorChannel struct {
	ID                 int64  `db:"id"`
	GuildID            int64  `db:"guild_id"`
	ChannelSnowflakeID string `db:"channel_snowflake_id"`
	UserLimit          int64  `db:"user_limit"`
}

type creatorChannelParams struct {
	guildID            int64
	channelSnowflakeID string
	userLimit          int64
}

type creatorChannelFilter struct {
	id                 int64
	guildID            int64
	channelSnowflakeID string
	limit              uint64
}

type temporaryVoiceChannel struct {
	ID                 int64  `db:"id"`
	GuildID            int64  `db:"guild_id"`
	ChannelSnowflakeID string `db:"channel_snowflake_id"`
	UserCount          int64  `db:"user_count"`
	OwnerSnowflakeID   string `db:"owner_snowflake_id"`
}

type temporaryVoiceChannelParams struct {
	guildID            int64
	channelSnowflakeID string
	userCount          int64
	ownerSnowflakeID   string
}

type temporaryVoiceChannelFilter struct {
	id                 int64
	guildID            int64
	channelSnowflakeID string
	ownerSnowflakeID   string
	limit              uint64
}
