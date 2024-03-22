package bot

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"

	"github.com/tombuente/tempus/sql"
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
	_, err := db.Exec(sql.TempusSchema)
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
			slog.Error("Cannot register application command", "command", command.command.Name, "error", err)
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
	guild, ok, err := b.db.guild(guildFilter{guildID: e.GuildID})
	if err != nil {
		b.interactionRespondWithMessage("Internal error", s, e)
		slog.Error("Unable to get database entry", "event", "InteractionCreate", "command", "add", "error", err)
		return
	}
	if !ok {
		guild, err = b.db.createServer(guildParams{guildID: e.GuildID})
		if err != nil {
			b.interactionRespondWithMessage("Internal error", s, e)
			slog.Error("Unable to create guild in database", "event", "InteractionCreate", "command", "add", "error", err)
			return
		}
	}

	channel, err := s.GuildChannelCreate(e.GuildID, "Create Voice Channel", dgo.ChannelTypeGuildVoice)
	if err != nil {
		b.interactionRespondWithMessage("Unable to create voice channel", s, e)
		slog.Error("Unable to create voice channel", "event", "InteractionCreate", "guild_id", e.GuildID, "error", err)
		return
	}

	_, err = b.db.createCreatorChannel(creatorChannelParams{guildID: guild.ID, channelID: channel.ID})
	if err != nil {
		b.interactionRespondWithMessage("Internal error", s, e)
		slog.Error("Unable to store CreatorChannel in database", "event", "InteractionCreate", "guild_id", e.GuildID, "error", err)
		return
	}

	b.interactionRespondWithMessage("Creator channel created succesfully", s, e)
	slog.Info("Added creator channel", "event", "InteractionCreate", "guild_id", e.GuildID)
}
