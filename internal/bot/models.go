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
	ID        int64  `db:"id"`
	GuildID   int64  `db:"guild_id"` // guild.ID not guild.GuildID
	ChannelID string `db:"channel_id"`
	MaxUsers  int64  `db:"max_users"`
}

type creatorChannelParams struct {
	guildID   int64 // guild.ID not guild.GuildID
	channelID string
	maxUsers  int64
}

// type creatorChannelFilter struct {
// 	id        int64
// 	guildID   int64 // guild.ID not guild.GuildID
// 	channelID string
// 	limit     uint64
// }
