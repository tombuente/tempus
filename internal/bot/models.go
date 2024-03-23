package bot

type guild struct {
	ID      int64  `db:"id"`
	GuildID string `db:"guild_id"`
}

type guildParams struct {
	guildID string
}

type guildFilter struct {
	id      int64
	guildID string
	limit   uint64
}

type creatorChannel struct {
	ID                 int64  `db:"id"`
	GuildID            int64  `db:"guild_id"`
	ChannelSnowflakeID string `db:"channel_id"`
	MaxUsers           int64  `db:"max_users"`
}

type creatorChannelParams struct {
	guildID            int64
	channelSnowflakeID string
	maxUsers           int64
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
	ChannelSnowflakeID string `db:"channel_id"`
	OwnerID            int64  `db:"owner_id"`
}

type temporaryVoiceChannelParams struct {
	guildID            int64
	channelSnowflakeID string
	ownerSnowflakeID   string
}

type temporaryVoiceChannelFilter struct {
	id                 int64
	guildID            uint64
	channelSnowflakeID string
	ownerSnowflakeID   string
	limit              uint64
}
