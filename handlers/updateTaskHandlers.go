package handlers

import (
	"database/sql"
	"encoding/json"
	"finalproject/functions"
	"net/http"
	"time"
)

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {

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

	if scheduler.Id == "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "id не может быть пустым"})
		return
	}

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM scheduler WHERE id = ?)", scheduler.Id).Scan(&exists)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "ошибка при поиске задачи по идентификатору"})
		return
	}

	if !exists {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Задача не найдена"})
		return
	}

	_, err = db.Exec(`
	       UPDATE scheduler
	       SET date = ?, title = ?, comment = ?, repeat = ?
	       WHERE id = ?
	                    `, scheduler.Date, scheduler.Title, scheduler.Comment, scheduler.Repeat, scheduler.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка обновления задачи"})
		return

	}

	w.Header().Set("Content-Type", "application/json;charset = UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "данные задачи обновлены"})
}
