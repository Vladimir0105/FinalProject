package functions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const dbFileName = "scheduler.db"

// Функция InitDB открывает или создаёт(если БД не существует) новую базу данных
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

// Функция createTables создаёт в базе данных нужную нам таблицу
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

// Функция createIndexes создаёт индекс по столбцу date
func createIndexes(db *sql.DB) error {
	sql := `
		CREATE INDEX scheduler_date ON scheduler (date);`

	_, err := db.Exec(sql)
	return err
}

// Функция NextDate вычисляет следующую дату по правилу repeat
func NextDate(now time.Time, date string, repeat string) (string, error) {

	dateTime, err := time.Parse("20060102", date)
	if err != nil {
		return "", err
	}

	repeatSlice := strings.Split(repeat, " ")

	if len(repeatSlice) == 0 {
		return "", fmt.Errorf("значений в repeat нету")
	}

	switch repeatSlice[0] {
	case "y":
		dateTime = dateTime.AddDate(1, 0, 0)
		for !dateTime.After(now) {
			dateTime = dateTime.AddDate(1, 0, 0)
		}
		return dateTime.Format("20060102"), nil

	case "d":
		if len(repeatSlice) < 2 {
			return "", fmt.Errorf("неккоректный формат данных")
		}
		numberOfDays, err := strconv.Atoi(repeatSlice[1])
		if err != nil {
			log.Println("Ошибка преобразования строки в число")
			return "", err
		}
		if numberOfDays < 1 || numberOfDays > 400 {
			return "", fmt.Errorf("некоректное количество дней")
		}

		dateTime = dateTime.AddDate(0, 0, numberOfDays)
		for !dateTime.After(now) {
			dateTime = dateTime.AddDate(0, 0, numberOfDays)
		}
		return dateTime.Format("20060102"), nil
	default:
		return "", fmt.Errorf("неккоректный формат данных")
	}
}

type Schedule struct {
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Функция ExtractValues извлекает значения из базы данных
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

// Функция SearchTaskById проверяет наличие задачи по идентификатору
func SearchTaskById(w http.ResponseWriter, r *http.Request, db *sql.DB) (bool, error) {

	var task Schedule
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM scheduler WHERE id = ?)", task.Id).Scan(&exists)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return false, json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при поиске задачи по идентификатору"})
	}

	if !exists {
		w.WriteHeader(http.StatusInternalServerError)

		return false, json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при поиске задачи"})
	}
	return true, nil
}

// Функция UpdateTaskDateInDB обновляет значение даты по правилу в базе данных
func UpdateTaskDateInDB(w http.ResponseWriter, r *http.Request, db *sql.DB) (Schedule, error) {

	var scheduler Schedule

	err := json.NewDecoder(r.Body).Decode(&scheduler)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return Schedule{}, json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат json"})
	}

	if scheduler.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		return Schedule{}, json.NewEncoder(w).Encode(map[string]string{"error": "Поле title обязательно"})
	}

	if scheduler.Date == "" {
		scheduler.Date = time.Now().Format("20060102")
	} else {
		_, err := time.Parse("20060102", scheduler.Date)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return Schedule{}, json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат даты, используйте формат 20060102"})
		}
	}

	if scheduler.Repeat != "" && scheduler.Date != time.Now().Format("20060102") {
		nextDate, err := NextDate(time.Now(), scheduler.Date, scheduler.Repeat)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return Schedule{}, json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка вычисления следущей даты"})
		}
		scheduler.Date = nextDate
	} else {
		scheduler.Date = time.Now().Format("20060102")
	}

	if scheduler.Id == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return Schedule{}, json.NewEncoder(w).Encode(map[string]string{"error": "id не может быть пустым"})
	}

	return scheduler, nil
}
