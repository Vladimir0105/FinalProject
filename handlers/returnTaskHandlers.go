package handlers

import (
	"database/sql"
	"encoding/json"
	"finalproject/functions"
	"net/http"
)

// Функция ReturnTaskHandler возвращает задачу по идентификатору
func ReturnTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	var returnTask functions.Schedule

	idStr := r.URL.Query().Get("id")

	err := db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", idStr).Scan(&returnTask.Id, &returnTask.Date, &returnTask.Title, &returnTask.Comment, &returnTask.Repeat)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка выполнения запроса"})
		return
	}

	w.Header().Set("Content-Type", "application/json;charset = UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(returnTask)
}
