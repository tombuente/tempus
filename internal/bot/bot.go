package bot

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"

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

	minValue := float64(0)

	commands := map[string]commandWithHandler{
		"add": {
			command: &dgo.ApplicationCommand{
				Name:        "add",
				Description: "Add creator voice channel",
				Options: []*dgo.ApplicationCommandOption{
					{
						Type:        dgo.ApplicationCommandOptionInteger,
						Name:        "limit",
						Description: "User limit",
						MinValue:    &minValue,
						MaxValue:    100,
						Required:    false,
					},
				},
			},
			handler: b.interactionCreateAdd,
		},
		"kick": {
			command: &dgo.ApplicationCommand{
				Name:        "kick",
				Description: "Kick user from temporary voice channel",
				Options: []*dgo.ApplicationCommandOption{
					{
						Type:        dgo.ApplicationCommandOptionMentionable,
						Name:        "user",
						Description: "User to kick",
						Required:    true,
					},
				},
			},
			handler: b.interactionCreateKick,
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

func snowflakeToID(snowflake string) int64 {
	id, err := strconv.ParseInt(snowflake, 10, 64)
	if err != nil {
		slog.Error("Unable to convert snowflake to ID", "error", err)
	}

	return id
}

func idToSnowflake(id int64) string {
	return fmt.Sprintf("%v", id)
}

func (b *bot) interactionRespondWithMessage(message string, s *dgo.Session, e *dgo.InteractionCreate) {
	s.InteractionRespond(e.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{
			Flags:   dgo.MessageFlagsEphemeral,
			Content: message,
		},
	})
}

func (b *bot) interactionCreateAdd(s *dgo.Session, e *dgo.InteractionCreate) {
	options := make(map[string]*dgo.ApplicationCommandInteractionDataOption, len(e.ApplicationCommandData().Options))
	for _, option := range e.ApplicationCommandData().Options {
		options[option.Name] = option
	}

	data := dgo.GuildChannelCreateData{
		Name: "Create Voice Channel",
		Type: dgo.ChannelTypeGuildVoice,
	}
	if option, ok := options["limit"]; ok {
		data.UserLimit = int(option.IntValue())
	}

	discordChannel, err := s.GuildChannelCreateComplex(e.GuildID, data)
	if err != nil {
		b.interactionRespondWithMessage("Unable to create channel", s, e)
		slog.Error("Unable to create creator channel in database",
			"event", "InteractionCreate",
			"command", "add",
			"error", err)
		return
	}

	params := creatorChannelParams{
		id:      snowflakeToID(discordChannel.ID),
		guildID: snowflakeToID(e.GuildID),
	}
	if option, ok := options["limit"]; ok {
		params.userLimit = option.IntValue()
	}
	_, err = b.db.createCreatorChannel(params)
	if err != nil {
		b.interactionRespondWithMessage("Internal error", s, e)
		slog.Error("Unable to create creator channel in database",
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

func (b *bot) interactionCreateKick(s *dgo.Session, e *dgo.InteractionCreate) {
	options := make(map[string]*dgo.ApplicationCommandInteractionDataOption, len(e.ApplicationCommandData().Options))
	for _, option := range e.ApplicationCommandData().Options {
		options[option.Name] = option
	}
	userID := options["user"].Value.(string)

	temporaryVoiceChannel, ok, err := b.db.temporaryVoiceChannel(temporaryVoiceChannelFilter{ownerID: snowflakeToID(e.Member.User.ID)})
	if err != nil {
		b.interactionRespondWithMessage("Internal error", s, e)
		slog.Error("Unable to get temporary voice channel from database",
			"event", "InteractionCreate",
			"command", "kick",
			"error", err)
		return
	}
	if !ok {
		b.interactionRespondWithMessage("You do not own a channel", s, e)
		slog.Info("User tried to kick but is not a channel owner",
			"event", "InteractionCreate",
			"command", "kick")
		return
	}

	discordGuild, err := s.State.Guild(e.GuildID)
	if err != nil {
		b.interactionRespondWithMessage("Unable to get guild state from Discord", s, e)
		slog.Info("Unable to get guild state from Discord",
			"event", "InteractionCreate",
			"command", "kick")
		return
	}

	for _, state := range discordGuild.VoiceStates {
		if state.UserID != userID {
			fmt.Println("User IDs do not match")
			continue
		}

		if state.ChannelID != idToSnowflake(temporaryVoiceChannel.ID) {
			fmt.Println("Channel IDs do not match")
			continue
		}

		err = s.GuildMemberMove(e.GuildID, userID, nil)
		if err != nil {
			b.interactionRespondWithMessage("Unable to kick user", s, e)
			slog.Error("Unable to kick user",
				"event", "InteractionCreate",
				"command", "kick",
				"error", err)
			return
		}

		b.interactionRespondWithMessage("User kicked", s, e)
		slog.Info("User kicked from temporary voice channel",
			"event", "InteractionCreate",
			"command", "kick")
	}
}

func (b *bot) voiceStateUpdate(s *dgo.Session, e *dgo.VoiceStateUpdate) {
	userLeft := e.ChannelID == ""
	userJoined := e.BeforeUpdate == nil
	userMoved := !userLeft && !userJoined && e.ChannelID != e.BeforeUpdate.ChannelID

	if userLeft || userMoved { // User left channel
		slog.Info("User left voice channel",
			"event", "VoiceStateUpdate",
			"channel_id", e.BeforeUpdate.ChannelID)
		b.voiceStateUpdateLeave(s, e)
	}

	if userJoined || userMoved { // User joined channel
		slog.Info("User joined voice channel",
			"event", "VoiceStateUpdate",
			"channel_id", e.ChannelID)
		b.voiceStateUpdateJoin(s, e)
	}
}

func (b *bot) voiceStateUpdateJoin(s *dgo.Session, e *dgo.VoiceStateUpdate) {
	creatorChannel, ok, err := b.db.creatorChannel(creatorChannelFilter{id: snowflakeToID(e.ChannelID)})
	if err != nil {
		slog.Error("Unable to get creator channel from database",
			"event", "VoiceStateUpdate",
			"action", "Join",
			"error", err)
		return
	}
	if ok {
		b.voiceStateUpdateJoinCreatorChannel(creatorChannel, s, e)
		return
	}

	temporaryVoiceChannel, ok, err := b.db.temporaryVoiceChannel(temporaryVoiceChannelFilter{id: snowflakeToID(e.ChannelID)})
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

func (b *bot) voiceStateUpdateJoinCreatorChannel(creatorChannel creatorChannel, s *dgo.Session, e *dgo.VoiceStateUpdate) {
	username := e.Member.Nick
	if username == "" {
		username = e.Member.User.Username
	}

	channelCreateData := dgo.GuildChannelCreateData{
		Name: fmt.Sprintf("%v's Channel", username),
		Type: dgo.ChannelTypeGuildVoice,
	}
	if creatorChannel.UserLimit > 0 {
		channelCreateData.UserLimit = int(creatorChannel.UserLimit)
	}
	discordChannel, err := s.GuildChannelCreateComplex(e.GuildID, channelCreateData)
	if err != nil {
		slog.Error("Unable to create Discord voice channel",
			"event", "VoiceStateUpdate",
			"action", "JoinCreatorChannel",
			"error", err)
		return
	}

	params := temporaryVoiceChannelParams{
		id:      snowflakeToID(discordChannel.ID),
		guildID: snowflakeToID(e.GuildID),
		ownerID: snowflakeToID(e.UserID),
	}
	_, err = b.db.createTemporaryVoiceChannel(params)
	if err != nil {
		slog.Error("Unable to create temporary voice channel in database",
			"event", "VoiceStateUpdate",
			"action", "JoinCreatorChannel",
			"error", err)
	}

	err = s.GuildMemberMove(e.GuildID, e.UserID, &discordChannel.ID)
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
		id:        temporaryVoiceChannel.ID,
		guildID:   temporaryVoiceChannel.GuildID,
		userCount: temporaryVoiceChannel.UserCount + 1,
		ownerID:   temporaryVoiceChannel.OwnerID,
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
	temporaryVoiceChannel, ok, err := b.db.temporaryVoiceChannel(temporaryVoiceChannelFilter{id: snowflakeToID(e.BeforeUpdate.ChannelID)})
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
		_, err := s.ChannelDelete(idToSnowflake(temporaryVoiceChannel.ID))
		if err != nil {
			slog.Error("Unable to delete Discord channel",
				"event", "VoiceStateUpdate",
				"action", "LeaveTemporaryVoiceChannel",
				"id", temporaryVoiceChannel.ID,
				"error", err)
		}
		slog.Info("Discord channel deleted",
			"event", "VoiceStateUpdate",
			"action", "LeaveTemporaryVoiceChannel",
			"id", temporaryVoiceChannel.ID)

		_, err = b.db.deleteTemporaryVoiceChannel(temporaryVoiceChannel.ID)
		if err != nil {
			slog.Error("Unable to delete temporary voice channel from database",
				"event", "VoiceStateUpdate",
				"action", "LeaveTemporaryVoiceChannel",
				"id", temporaryVoiceChannel.ID,
				"error", err)
		}

		slog.Info("Temporary voice channel deleted from database",
			"event", "VoiceStateUpdate",
			"action", "LeaveTemporaryVoiceChannel",
			"id", temporaryVoiceChannel.ID)
		return
	}

	params := temporaryVoiceChannelParams{
		id:        temporaryVoiceChannel.ID,
		guildID:   temporaryVoiceChannel.GuildID,
		userCount: temporaryVoiceChannel.UserCount - 1,
		ownerID:   temporaryVoiceChannel.OwnerID,
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
