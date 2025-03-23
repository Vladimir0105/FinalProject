package handlers

import (
	"database/sql"
	"encoding/json"
	"finalproject/functions"
	"log"
	"net/http"
	"strconv"
)

const (
	limit = 10
)

// Функция UpcomingTaskHandler возвращает список , отсортированных по дате задач
func UpcomingTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	rows, err := db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?", limit)
	if err != nil {
		log.Println("Ошибка выполнения запроса", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка выполнения запроса"})
		return
	}

	defer rows.Close()

	var tasks []functions.Schedule = []functions.Schedule{}

	for rows.Next() {
		var task functions.Schedule = functions.Schedule{}
		var id int

		err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			log.Println("Ошибка сканирования строки", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка сканирования строки"})
			return
		}

		task.Id = strconv.Itoa(id)
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		log.Println("Ошибка перебора строк", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка перебора строк"})
		return

	}

	response := map[string][]functions.Schedule{"tasks": tasks}

	jsonData, err := json.MarshalIndent(response, " ", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка сериализации"})
		return

	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
