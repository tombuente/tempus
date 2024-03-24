package bot

type creatorChannel struct {
	ID        int64 `db:"id"`
	GuildID   int64 `db:"guild_id"`
	UserLimit int64 `db:"user_limit"`
}

type creatorChannelParams struct {
	id        int64
	guildID   int64
	userLimit int64
}

type creatorChannelFilter struct {
	id      int64
	guildID int64
	limit   uint64
}

type temporaryVoiceChannel struct {
	ID        int64 `db:"id"`
	GuildID   int64 `db:"guild_id"`
	UserCount int64 `db:"user_count"`
	OwnerID   int64 `db:"owner_id"`
}

type temporaryVoiceChannelParams struct {
	id        int64
	guildID   int64
	userCount int64
	ownerID   int64
}

type temporaryVoiceChannelFilter struct {
	id        int64
	guildID   int64
	userCount int64
	limit     uint64
}
