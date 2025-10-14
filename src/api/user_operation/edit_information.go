package user_operation

import (
	"database/sql"
	"src/api/database"
	"time"
)

func edit_discord(discord, name string) error {
	db := database.InitDB()
	defer db.Close()

	_, err := db.Exec("UPDATE users SET discordID = ? WHERE username = ?;", discord, name)

	if err != nil {
		return err
	}

	return nil
}
