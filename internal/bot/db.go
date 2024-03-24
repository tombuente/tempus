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
		Columns("id, guild_id, user_limit").
		Values(params.id, params.guildID, params.userLimit).
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
	if filter.userCount != 0 {
		builder = builder.Where(squirrel.Eq{"user_count": filter.userCount})
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
		Columns("id, guild_id, user_count, owner_id").
		Values(params.id, params.guildID, params.userCount, params.ownerID).
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

func (db database) updateTemporaryVoiceChannel(id int64, params temporaryVoiceChannelParams) (temporaryVoiceChannel, error) {
	query, args, err := squirrel.Update("temporary_voice_channels").
		Set("id", params.id).
		Set("guild_id", params.guildID).
		Set("user_count", params.userCount).
		Set("owner_id", params.ownerID).
		Where(squirrel.Eq{"id": id}).
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

func (db database) deleteTemporaryVoiceChannel(id int64) (temporaryVoiceChannel, error) {
	query, args, err := squirrel.Delete("temporary_voice_channels").
		Where(squirrel.Eq{"id": id}).
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
