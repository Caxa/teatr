package backend

import (
	"database/sql"
	"fmt"
	"log"

	"os"
)

// SetDB устанавливает соединение с БД для модуля backend
func SetDB(database *sql.DB) {
	db = database
}

// OpenDatabase создает новое соединение с базой данных
func OpenDatabase() (*sql.DB, error) {
	port := getEnv("DB_PORT", "5432")
	host := getEnv("DB_HOST", "localhost")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "1234")
	dbname := getEnv("DB_NAME", "kola")

	dbConn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", dbConn)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии соединения с БД: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка при проверке соединения с БД: %v", err)
	}

	log.Println("Успешно подключено к базе данных")
	return db, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
