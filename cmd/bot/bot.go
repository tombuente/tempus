package main

import (
	"errors"
	"log/slog"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tombuente/tempus/internal/bot"
)

func main() {
	token := os.Getenv("TEMPUS_TOKEN")
	guildID := os.Getenv("TEMPUS_GUILD_ID")

	if err := os.Mkdir("data", os.ModePerm); err != nil {
		switch {
		case errors.Is(err, os.ErrExist):
			slog.Info("Data directory already exists")
		default:
			slog.Error("Unable to create data directory")
			return
		}
	}
	slog.Info("Filesystem setup complete")

	db := sqlx.MustConnect("sqlite3", "./data/bot.db")

	bot.Run(token, guildID, db)
}
