package handlers

import (
	"database/sql"
	"encoding/json"
	"finalproject/functions"
	"net/http"
)

// Функция UpdateTaskHandler обновляет данные задачи
func UpdateTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	var scheduler functions.Schedule

	functions.SearchTaskById(w, r, db)

	scheduler, err := functions.UpdateTaskDateInDB(w, r, db)

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
