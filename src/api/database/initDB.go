package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

func InitDB() *sql.DB {
	dbUsername, ok := os.LookupEnv("DB_USERNAME")
	if !ok {
		log.Fatal("Не указана переменная окружения DB_USERNAME")
	}

	dbPassword, ok := os.LookupEnv("DB_PASSWORD")
	if !ok {
		log.Fatal("Не указана переменная окружения DB_PASSWORD")
	}

	dbName, ok := os.LookupEnv("DB_NAME")
	if !ok {
		log.Fatal("Не указана переменная окружения DB_NAME")
	}

	// Задаём параметры подключения: юзер, пароль, хост, порт, имя базы
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?parseTime=true",
		dbUsername, dbPassword, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Ошибка при открытии соединения с MySQL:", err)
	}

	// Проверим, что база реально доступна
	if err := db.Ping(); err != nil {
		log.Fatal("MySQL не отвечает:", err)
	}

	return db
}
