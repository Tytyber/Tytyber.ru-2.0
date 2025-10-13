package user_operation

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"math/big"
	"src/api/database"
	"time"
)

var (
	ErrUserExists   = errors.New("пользователь уже существует")
	ErrInvalidInput = errors.New("некорректные данные")
)

func RegisterUser(name, password, mail string) error {
	if name == "" || password == "" {
		return ErrInvalidInput
	}

	db := database.InitDB()
	defer db.Close()

	var exists int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", name).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if exists > 0 {
		return ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (username, passwd, email, discordID, datereg, uuid, rules) VALUES (?, ?, ?, ?, ?, ?, ?)",
		name, hashedPassword, mail, " ", time.Now(), gen_key(), 3)
	if err != nil {
		return err
	}

	userID, err := database.GetUserIDByUsername(db, name)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO wallets (userid, money, uuid) VALUES (?, ?, ?)",
		userID, 0, genWallet())
	if err != nil {
		return err
	}

	return nil
}

func gen_key() string {
	prefix := "t_userkey-"
	characters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	key := make([]byte, 50)
	for i := 0; i < 50; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(characters))))
		if err != nil {
			return ""
		}
		key[i] = characters[n.Int64()]
	}
	return prefix + string(key)
}

func genWallet() string {
	db := database.InitDB()
	defer db.Close()

	// Забираем MAX(id) и смотрим, есть ли вообще значение
	var maxID sql.NullInt64
	err := db.QueryRow("SELECT MAX(id) FROM users").Scan(&maxID)
	if err != nil {
		fmt.Errorf("ошибка при запросе MAX(id): %w", err)
		return ""
	}
	if !maxID.Valid {
		maxID.Int64 = 0
	}

	wallet := fmt.Sprintf("T-WALLET-%d", maxID.Int64)
	return wallet
}
