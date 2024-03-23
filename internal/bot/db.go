package bot

import (
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type database struct {
	db *sqlx.DB
}

func newDatabase(db *sqlx.DB) database {
	return database{
		db: db,
	}
}

func (db database) guild(filter guildFilter) (guild, bool, error) {
	filter.limit = 1

	servers, ok, err := db.guilds(filter)
	if err != nil {
		return guild{}, ok, nil
	}
	if !ok {
		return guild{}, ok, nil
	}

	return servers[0], ok, nil
}

func (db database) guilds(filter guildFilter) ([]guild, bool, error) {
	builder := squirrel.Select("*").From("guilds")

	if filter.id != 0 {
		builder = builder.Where(squirrel.Eq{"id": filter.id})
	}
	if filter.guildID != "" {
		builder = builder.Where(squirrel.Eq{"guild_id": filter.guildID})
	}

	if filter.limit != 0 {
		builder = builder.Limit(filter.limit)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return []guild{}, false, joinErrors(ErrSQLInternal, err)
	}

	rows, err := db.db.Queryx(query, args...)
	if err != nil {
		return []guild{}, false, joinErrors(ErrSQLInternal, err)
	}
	defer rows.Close()

	var guilds []guild
	for rows.Next() {
		var i guild
		err := rows.StructScan(&i)
		if err != nil {
			return []guild{}, false, joinErrors(ErrSQLInternal, err)
		}

		guilds = append(guilds, i)
	}

	if len(guilds) == 0 {
		return []guild{}, false, nil
	}

	return guilds, true, nil
}

func (db database) createGuild(params guildParams) (guild, error) {
	query, args, err := squirrel.Insert("guilds").
		Columns("guild_id").
		Values(params.guildID).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return guild{}, joinErrors(ErrSQLInternal, err)
	}

	row := db.db.QueryRowx(query, args...)

	var i guild
	err = row.StructScan(&i)
	if err != nil {
		return guild{}, joinErrors(ErrSQLInternal, err)
	}

	return i, nil
}

func (db database) creatorChannel(filter creatorChannelFilter) (creatorChannel, bool, error) {
	filter.limit = 1

	channels, ok, err := db.creatorChannels(filter)
	if err != nil {
		return creatorChannel{}, ok, nil
	}
	if !ok {
		return creatorChannel{}, ok, nil
	}

	return channels[0], ok, nil
}

func (db database) creatorChannels(filter creatorChannelFilter) ([]creatorChannel, bool, error) {
	builder := squirrel.Select("*").From("creator_channels")

	if filter.id != 0 {
		builder = builder.Where(squirrel.Eq{"id": filter.id})
	}
	if filter.guildID != 0 {
		builder = builder.Where(squirrel.Eq{"guild_id": filter.guildID})
	}
	if filter.channelSnowflakeID != "" {
		builder = builder.Where(squirrel.Eq{"channel_id": filter.channelSnowflakeID})
	}

	if filter.limit != 0 {
		builder = builder.Limit(filter.limit)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return []creatorChannel{}, false, joinErrors(ErrSQLInternal, err)
	}

	rows, err := db.db.Queryx(query, args...)
	if err != nil {
		return []creatorChannel{}, false, joinErrors(ErrSQLInternal, err)
	}
	defer rows.Close()

	var channels []creatorChannel
	for rows.Next() {
		var i creatorChannel
		err := rows.StructScan(&i)
		if err != nil {
			return []creatorChannel{}, false, joinErrors(ErrSQLInternal, err)
		}

		channels = append(channels, i)
	}

	if len(channels) == 0 {
		return []creatorChannel{}, false, nil
	}

	return channels, true, nil
}

func (db database) createCreatorChannel(params creatorChannelParams) (creatorChannel, error) {
	query, args, err := squirrel.Insert("creator_channels").
		Columns("guild_id, channel_id, max_users").
		Values(params.guildID, params.channelSnowflakeID, params.maxUsers).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return creatorChannel{}, joinErrors(ErrSQLInternal, err)
	}

	row := db.db.QueryRowx(query, args...)

	var i creatorChannel
	err = row.StructScan(&i)
	if err != nil {
		return creatorChannel{}, joinErrors(ErrSQLInternal, err)
	}

	return i, nil
}

func (db database) temporaryVoiceChannel(filter temporaryVoiceChannelFilter) (temporaryVoiceChannel, bool, error) {
	filter.limit = 1

	channels, ok, err := db.temporaryVoiceChannels(filter)
	if err != nil {
		return temporaryVoiceChannel{}, ok, nil
	}
	if !ok {
		return temporaryVoiceChannel{}, ok, nil
	}

	return channels[0], ok, nil
}

func (db database) temporaryVoiceChannels(filter temporaryVoiceChannelFilter) ([]temporaryVoiceChannel, bool, error) {
	builder := squirrel.Select("*").From("temporary_voice_channels")

	if filter.id != 0 {
		builder = builder.Where(squirrel.Eq{"id": filter.id})
	}
	if filter.guildID != 0 {
		builder = builder.Where(squirrel.Eq{"guild_id": filter.guildID})
	}
	if filter.channelSnowflakeID != "" {
		builder = builder.Where(squirrel.Eq{"channel_id": filter.channelSnowflakeID})
	}
	if filter.ownerSnowflakeID != "" {
		builder = builder.Where(squirrel.Eq{"owner_id": filter.ownerSnowflakeID})
	}

	if filter.limit != 0 {
		builder = builder.Limit(filter.limit)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return []temporaryVoiceChannel{}, false, joinErrors(ErrSQLInternal, err)
	}

	rows, err := db.db.Queryx(query, args...)
	if err != nil {
		return []temporaryVoiceChannel{}, false, joinErrors(ErrSQLInternal, err)
	}
	defer rows.Close()

	var channels []temporaryVoiceChannel
	for rows.Next() {
		var i temporaryVoiceChannel
		err := rows.StructScan(&i)
		if err != nil {
			return []temporaryVoiceChannel{}, false, joinErrors(ErrSQLInternal, err)
		}

		channels = append(channels, i)
	}

	if len(channels) == 0 {
		return []temporaryVoiceChannel{}, false, nil
	}

	return channels, true, nil
}

func (db database) createTemporaryVoiceChannel(params temporaryVoiceChannelParams) (temporaryVoiceChannel, error) {
	query, args, err := squirrel.Insert("temporary_voice_channels").
		Columns("guild_id, channel_id, owner_id").
		Values(params.guildID, params.channelSnowflakeID, params.ownerSnowflakeID).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return temporaryVoiceChannel{}, joinErrors(ErrSQLInternal, err)
	}

	row := db.db.QueryRowx(query, args...)

	var i temporaryVoiceChannel
	err = row.StructScan(&i)
	if err != nil {
		return temporaryVoiceChannel{}, joinErrors(ErrSQLInternal, err)
	}

	return i, nil
}
