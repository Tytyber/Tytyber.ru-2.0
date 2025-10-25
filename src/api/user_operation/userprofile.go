package user_operation

import (
	"database/sql"
)

type User struct {
	UserID   int64
	Nickname string
	Email    string
	Rules    string
	Discord  string
	UUID     string
	WaletID  string
	Money    float32
}

func GetUserProfile(db *sql.DB, username string) (*User, error) {
	u := &User{}

	query := `
        SELECT
            id,
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
		&u.UserID,
		&u.Nickname,
		&u.Email,
		&u.Rules,
		&u.Discord,
		&u.UUID,
	)
	if err != nil {
		return nil, err
	}
	query = `
        SELECT 
            money,
            uuid
        FROM wallets
        WHERE userid = ?
    `
	err = db.QueryRow(query, &u.UserID).Scan(
		&u.Money,
		&u.WaletID,
	)

	if err != nil {
		return nil, err
	}

	return u, nil
}
