package user_operation

import (
	"database/sql"
)

type User struct {
	Nickname string
	Email    string
	Rules    string
	Discord  string
	UUID     string
}

func GetUserProfile(db *sql.DB, username string) (*User, error) {
	u := &User{}

	query := `
        SELECT 
            username,
            email,
            rules,
            discordID,
        	uuid
        FROM users
        WHERE username = ?
    `
	// Сканим date_registry в dateStr, а не сразу в time.Time
	err := db.QueryRow(query, username).Scan(
		&u.Nickname,
		&u.Email,
		&u.Rules,
		&u.Discord,
		&u.UUID,
	)
	if err != nil {
		return nil, err
	}

	return u, nil
}
