package bot

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	"github.com/tombuente/tempus/internal/sql"
)

type commandWithHandler struct {
	command *dgo.ApplicationCommand
	handler func(s *dgo.Session, i *dgo.InteractionCreate)
}

type bot struct {
	db database

	session *dgo.Session
}

func Run(token string, guildID string, db *sqlx.DB) {
	_, err := db.Exec(sql.BotSchema)
	if err != nil {
		slog.Error("Unable to load database schema", "error", err)
		return
	}

	session, err := dgo.New(fmt.Sprintf("Bot %v", token))
	if err != nil {
		slog.Error("Invalid bot parameters", "error", err)
		return
	}

	b := bot{
		db:      newDatabase(db),
		session: session,
	}

	commands := map[string]commandWithHandler{
		"add": {
			command: &dgo.ApplicationCommand{
				Name:        "add",
				Description: "Add creator voice channel",
			},
			handler: b.interactionCreateAdd,
		},
	}

	b.session.AddHandler(func(s *dgo.Session, i *dgo.InteractionCreate) {
		if command, ok := commands[i.ApplicationCommandData().Name]; ok {
			command.handler(s, i)
		}
	})

	b.session.AddHandler(b.voiceStateUpdate)

	err = b.session.Open()
	if err != nil {
		slog.Error("Cannot open session", "error", err)
		return
	}
	defer b.session.Close()

	slog.Info("Adding application commands")
	for _, command := range commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, guildID, command.command)
		if err != nil {
			slog.Error("Cannot register application command",
				"command", command.command.Name,
				"error", err)
		}
	}
	slog.Info("Added application commands")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	slog.Info("Press Ctrl+C to exit")
	<-stop

	slog.Info("Shutting down")
}

func (b *bot) interactionRespondWithMessage(message string, s *dgo.Session, e *dgo.InteractionCreate) {
	s.InteractionRespond(e.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{
			Content: message,
		},
	})
}

func (b *bot) interactionCreateAdd(s *dgo.Session, e *dgo.InteractionCreate) {
	guild, ok, err := b.db.guild(guildFilter{guildSnowflakeID: e.GuildID})
	if err != nil {
		b.interactionRespondWithMessage("Internal error", s, e)
		slog.Error("Unable to get guild from database",
			"event", "InteractionCreate",
			"command", "add",
			"error", err)
		return
	}
	if !ok {
		guild, err = b.db.createGuild(guildParams{guildSnowflakeID: e.GuildID})
		if err != nil {
			b.interactionRespondWithMessage("Internal error", s, e)
			slog.Error("Unable to create guild in database",
				"event", "InteractionCreate",
				"command", "add",
				"error", err)
			return
		}
	}

	channel, err := s.GuildChannelCreate(e.GuildID, "Create Voice Channel", dgo.ChannelTypeGuildVoice)
	if err != nil {
		b.interactionRespondWithMessage("Unable to create channel", s, e)
		slog.Error("Unable to create creator channel in database",
			"event", "InteractionCreate",
			"command", "add",
			"error", err)
		return
	}

	_, err = b.db.createCreatorChannel(creatorChannelParams{guildID: guild.ID, channelSnowflakeID: channel.ID})
	if err != nil {
		b.interactionRespondWithMessage("Internal error", s, e)
		slog.Error("Unable to store creator channel in database",
			"event", "InteractionCreate",
			"command", "add",
			"error", err)
		return
	}

	b.interactionRespondWithMessage("Channel created successfully", s, e)
	slog.Info("Added creator channel",
		"event", "InteractionCreate",
		"command", "add")
}

func (b *bot) voiceStateUpdate(s *dgo.Session, e *dgo.VoiceStateUpdate) {
	guild, ok, err := b.db.guild(guildFilter{guildSnowflakeID: e.GuildID})
	if err != nil {
		slog.Error("Unable to get guild from database",
			"event", "VoiceStateUpdate",
			"error", err)
		return
	}
	if !ok {
		slog.Info("Guild has not been set up yet",
			"event", "VoiceStateUpdate",
			"guild_snowflake_id", e.GuildID)
		return
	}

	userLeft := e.ChannelID == ""
	userJoined := e.BeforeUpdate == nil
	userMoved := !userLeft && !userJoined && e.ChannelID != e.BeforeUpdate.ChannelID

	if userLeft || userMoved { // User left channel
		slog.Info("User left voice channel",
			"event", "VoiceStateUpdate",
			"channel_snowflake_id", e.ChannelID)
		b.voiceStateUpdateLeave(s, e)
	}

	if userJoined || userMoved { // User joined channel
		slog.Info("User joined voice channel",
			"event", "VoiceStateUpdate",
			"channel_snowflake_id", e.ChannelID)
		b.voiceStateUpdateJoin(guild, s, e)
	}
}

func (b *bot) voiceStateUpdateJoin(guild guild, s *dgo.Session, e *dgo.VoiceStateUpdate) {
	_, ok, err := b.db.creatorChannel(creatorChannelFilter{channelSnowflakeID: e.ChannelID})
	if err != nil {
		slog.Error("Unable to get creator channel from database",
			"event", "VoiceStateUpdate",
			"action", "Join",
			"error", err)
		return
	}
	if ok {
		b.voiceStateUpdateJoinCreatorChannel(guild, s, e)
		return
	}

	temporaryVoiceChannel, ok, err := b.db.temporaryVoiceChannel(temporaryVoiceChannelFilter{channelSnowflakeID: e.ChannelID})
	if err != nil {
		slog.Error("Unable to get temporary voice channel from database",
			"event", "VoiceStateUpdate",
			"action", "Join",
			"error", err)
		return
	}
	if ok {
		b.voiceStateUpdateJoinTemporaryVoiceChannel(temporaryVoiceChannel)
		return
	}
}

func (b *bot) voiceStateUpdateJoinCreatorChannel(guild guild, s *dgo.Session, e *dgo.VoiceStateUpdate) {
	username := e.Member.Nick
	if username == "" {
		username = e.Member.User.Username
	}

	channelCreateData := dgo.GuildChannelCreateData{
		Name: fmt.Sprintf("%v's Channel", username),
		Type: dgo.ChannelTypeGuildVoice,
	}
	tempVoiceChannel, err := s.GuildChannelCreateComplex(e.GuildID, channelCreateData)
	if err != nil {
		slog.Error("Unable to create Discord voice channel",
			"event", "VoiceStateUpdate",
			"action", "JoinCreatorChannel",
			"error", err)
		return
	}

	_, err = b.db.createTemporaryVoiceChannel(temporaryVoiceChannelParams{guildID: guild.ID, channelSnowflakeID: tempVoiceChannel.ID, ownerSnowflakeID: e.UserID})
	if err != nil {
		slog.Error("Unable to create temporary voice channel in database",
			"event", "VoiceStateUpdate",
			"action", "JoinCreatorChannel",
			"error", err)
	}

	err = s.GuildMemberMove(e.GuildID, e.UserID, &tempVoiceChannel.ID)
	if err != nil {
		slog.Error("Unable to move user to temporary voice channel",
			"event", "VoiceStateUpdate",
			"action", "JoinCreatorChannel",
			"error", err)
		return
	}

	slog.Info("Created temporary voice channel and moved user",
		"event", "VoiceStateUpdate",
		"action", "JoinCreatorChannel")
}

func (b *bot) voiceStateUpdateJoinTemporaryVoiceChannel(temporaryVoiceChannel temporaryVoiceChannel) {
	params := temporaryVoiceChannelParams{
		guildID:            temporaryVoiceChannel.GuildID,
		channelSnowflakeID: temporaryVoiceChannel.ChannelSnowflakeID,
		userCount:          temporaryVoiceChannel.UserCount + 1,
		ownerSnowflakeID:   temporaryVoiceChannel.OwnerSnowflakeID,
	}

	_, err := b.db.updateTemporaryVoiceChannel(temporaryVoiceChannel.ID, params)
	if err != nil {
		slog.Error("Unable to update temporary voice channel in database",
			"event", "VoiceStateUpdate",
			"action", "JoinTemporaryVoiceChannel",
			"id", temporaryVoiceChannel.ID,
			"error", err)
		return
	}

	slog.Info("Updated temporary voice channel in database",
		"event", "VoiceStateUpdate",
		"action", "JoinTemporaryVoiceChannel",
		"id", temporaryVoiceChannel.ID)
}

func (b *bot) voiceStateUpdateLeave(s *dgo.Session, e *dgo.VoiceStateUpdate) {
	temporaryVoiceChannel, ok, err := b.db.temporaryVoiceChannel(temporaryVoiceChannelFilter{channelSnowflakeID: e.BeforeUpdate.ChannelID})
	if err != nil {
		slog.Error("Unable to get temporary voice channel from database",
			"event", "VoiceStateUpdate",
			"action", "Leave",
			"error", err)
		return
	}
	if ok {
		b.voiceStateUpdateLeaveTemporaryVoiceChannel(temporaryVoiceChannel, s)
		return
	}
}

func (b *bot) voiceStateUpdateLeaveTemporaryVoiceChannel(temporaryVoiceChannel temporaryVoiceChannel, s *dgo.Session) {
	if temporaryVoiceChannel.UserCount-1 == 0 {
		_, err := s.ChannelDelete(temporaryVoiceChannel.ChannelSnowflakeID)
		if err != nil {
			slog.Error("Unable to delete Discord channel",
				"event", "VoiceStateUpdate",
				"action", "LeaveTemporaryVoiceChannel",
				"channel_snowflake_id", temporaryVoiceChannel.ChannelSnowflakeID,
				"error", err)
		}
		slog.Info("Deleted Discord channel",
			"event", "VoiceStateUpdate",
			"action", "LeaveTemporaryVoiceChannel",
			"channel_snowflake_id", temporaryVoiceChannel.ChannelSnowflakeID)

		_, err = b.db.deleteTemporaryVoiceChannel(temporaryVoiceChannel.ID)
		if err != nil {
			slog.Error("Unable to delete temporary voice channel from database",
				"event", "VoiceStateUpdate",
				"action", "LeaveTemporaryVoiceChannel",
				"id", temporaryVoiceChannel.ID,
				"error", err)
		}

		slog.Info("Deleted temporary voice channel from database",
			"event", "VoiceStateUpdate",
			"action", "LeaveTemporaryVoiceChannel",
			"id", temporaryVoiceChannel.ID)
		return
	}

	params := temporaryVoiceChannelParams{
		guildID:            temporaryVoiceChannel.GuildID,
		channelSnowflakeID: temporaryVoiceChannel.ChannelSnowflakeID,
		userCount:          temporaryVoiceChannel.UserCount - 1,
		ownerSnowflakeID:   temporaryVoiceChannel.OwnerSnowflakeID,
	}

	_, err := b.db.updateTemporaryVoiceChannel(temporaryVoiceChannel.ID, params)
	if err != nil {
		slog.Error("Unable to update temporary voice channel in database",
			"event", "VoiceStateUpdate",
			"action", "JoinTemporaryVoiceChannel",
			"id", temporaryVoiceChannel.ID,
			"error", err)
		return
	}
}
