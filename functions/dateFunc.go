package functions

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Функция NextDate высчитывает следующую дату по правилу repeat
func NextDate(now time.Time, date string, repeat string) (string, error) {

	dateTime, err := time.Parse("20060102", date)
	if err != nil {
		return "", err
	}

	repeatSlice := strings.Split(repeat, " ")

	if len(repeatSlice) == 0 {
		return "", fmt.Errorf("значений в repeat нету, задача будет удалена")
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

// Функция UpdateDateBd обновляет значение даты в базе данных
func UpdateDateBd(db *sql.DB) error {
	schedules, err := ExtractValues(db)
	if err != nil {
		log.Print("ошибка при извлечении данных: ", err)
	}

	now := time.Now()

	for _, s := range schedules {

		newTaskDate, err := NextDate(now, s.Date, s.Repeat)
		if err != nil {
			log.Println("Ошибка обновления даты:", err)
			continue
		}

		if newTaskDate == "" {
			_, err = db.Exec("DELETE FROM scheduler WHERE date = ?", s.Date)
			if err != nil {
				log.Println("Ошибка при удалении данных")
			}
		}
		if newTaskDate == s.Date {
			log.Printf("Дата задачи не изменилась %v", newTaskDate)
		} else {
			log.Printf("Новая дата задачи: %v\n", newTaskDate)
			_, err = db.Exec("UPDATE scheduler SET date = ? WHERE date = ?", newTaskDate, s.Date)
			if err != nil {
				log.Println("Ошибка обновления даты:", err)
				continue
			}
			log.Println("Дата в базе данных обновлена")
		}
	}
	return nil
}
