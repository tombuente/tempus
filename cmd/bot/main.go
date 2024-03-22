package main

import (
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tombuente/tempus/internal/bot"
)

func main() {
	token := os.Getenv("TEMPUS_TOKEN")
	guildID := os.Getenv("TEMPUS_GUILD_ID")

	db := sqlx.MustConnect("sqlite3", "data.sqlite3")

	bot.Run(token, guildID, db)
}
