package handlers

import (
	"database/sql"
	"encoding/json"
	"finalproject/functions"
	"net/http"
	"time"
)

// Функция AddTaskHandler добавляет задачу
func AddTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	var scheduler functions.Schedule

	err := json.NewDecoder(r.Body).Decode(&scheduler)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат json"})
	}

	if scheduler.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Поле title обязательно"})
		return
	}

	if scheduler.Date == "" {
		scheduler.Date = time.Now().Format("20060102")
	} else {
		_, err := time.Parse("20060102", scheduler.Date)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат даты, используйте формат 20060102"})
			return
		}
	}

	if scheduler.Repeat != "" && scheduler.Date != time.Now().Format("20060102") {
		nextDate, err := functions.NextDate(time.Now(), scheduler.Date, scheduler.Repeat)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка вычисления следущей даты"})
			return
		}
		scheduler.Date = nextDate
	} else {
		scheduler.Date = time.Now().Format("20060102")
	}

	query := `
	          INSERT INTO scheduler (date, title, comment, repeat)
			  VALUES ($1, $2, $3, $4)
			  RETURNING id`
	var schedulerID int

	err = db.QueryRow(query, scheduler.Date, scheduler.Title, scheduler.Comment, scheduler.Repeat).Scan(&schedulerID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при добавлении задачи в базу данных"})
		return
	}

	response := map[string]int{"id": schedulerID}
	w.Header().Set("Content-Type", "application/json;charset = UTF-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

}
