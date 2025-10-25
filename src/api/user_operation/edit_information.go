package user_operation

import (
	"src/api/database"
)

func Edit_discord(discord, name string) error {
	db := database.InitDB()
	defer db.Close()

	_, err := db.Exec("UPDATE users SET discordID = ? WHERE username = ?;", discord, name)
	return err
}
