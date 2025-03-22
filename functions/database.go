package functions

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const dbFileName = "scheduler.db"

func InitDB() (*sql.DB, error) {
	appPath, err := os.Getwd()

	log.Printf("Текущая директория : %s\n", appPath)
	if err != nil {
		log.Fatal("Ошибка получения текущей дириктории:", err)
	}

	dbFile := filepath.Join(appPath, dbFileName)
	log.Printf("Файл базы данных будет создан по пути: %s\n", dbFile)

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal("Ошибка подключения к БД, создаём новую", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Ошибка подключения к базе данных", err)
	}

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		if err := createTables(db); err != nil {
			log.Fatal("Ошибка создания таблиц:", err)
		}
		if err := createIndexes(db); err != nil {
			log.Fatal("Ошибка создания индексов:", err)
		}
	} else {
		var tableName string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type ='table' AND name='scheduler' ").Scan(&tableName)
		if err != nil && err != sql.ErrNoRows {
			log.Fatal("Ошибка при проверке существования таблицы", err)
		}

		if tableName == "" {
			if err := createTables(db); err != nil {
				log.Fatal("Ошибка создания таблиц:", err)
			}
			if err := createIndexes(db); err != nil {
				log.Fatal("Ошибка создания индексов:", err)
			}

		}
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	sql := `
		CREATE TABLE scheduler (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date CHAR(8) NOT NULL DEFAULT '',
			title VARCHAR(256) NOT NULL DEFAULT '',
			comment VARCHAR(256) NOT NULL DEFAULT '',
			repeat VARCHAR(256) NOT NULL DEFAULT ''
		);`

	_, err := db.Exec(sql)

	log.Printf("Таблица scheduler в базе данных создана")

	return err
}

func createIndexes(db *sql.DB) error {
	sql := `
		CREATE INDEX scheduler_date ON scheduler (date);`

	_, err := db.Exec(sql)
	return err
}

type Schedule struct {
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func ExtractValues(db *sql.DB) ([]Schedule, error) {

	rows, err := db.Query("SELECT repeat, date FROM scheduler")
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос: %v", err)
	}

	defer rows.Close()

	var schedules []Schedule

	for rows.Next() {
		var s Schedule

		err := rows.Scan(&s.Repeat, &s.Date)
		if err != nil {
			return nil, fmt.Errorf("не удалось отсканировать строку: %v", err)
		}
		schedules = append(schedules, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после сканирования строк: %v", err)
	}

	return schedules, nil
}
