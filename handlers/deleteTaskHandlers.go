package handlers

import (
	"database/sql"
	"encoding/json"
	"finalproject/functions"
	"net/http"
)

func DeleteTaskTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	var task functions.Schedule

	idStr := r.URL.Query().Get("id")

	err := db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", idStr).Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка выполнения запроса"})
		return
	}

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM scheduler WHERE id = ?)", task.Id).Scan(&exists)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при поиске задачи по идентификатору"})
		return
	}

	if !exists {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при поиске задачи"})
		return
	}

	_, err = db.Exec("DELETE FROM scheduler WHERE id = ?", task.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка удаления задачи"})
		return
	}

	w.Header().Set("Content-Type", "application/json;charset = UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{})
}
