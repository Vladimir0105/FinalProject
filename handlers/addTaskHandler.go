package handlers

import (
	"database/sql"
	"encoding/json"
	"finalproject/functions"
	"net/http"
)

// Функция AddTaskHandler добавляет новую задачу
func AddTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	var scheduler functions.Schedule

	scheduler, err := functions.UpdateTaskDateInDB(w, r, db)

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
