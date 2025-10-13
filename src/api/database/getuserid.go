package database

import (
	"database/sql"
	"fmt"
)

// GetUserIDByUsername возвращает userid по имени пользователя.
func GetUserIDByUsername(db *sql.DB, username string) (int, error) {
	var userID int

	query := "SELECT id FROM users WHERE username = ? LIMIT 1"

	err := db.QueryRow(query, username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// если пользователь не найден
			return 0, fmt.Errorf("пользователь с именем '%s' не найден", username)
		}
		// любая другая ошибка БД
		return 0, err
	}

	return userID, nil
}
