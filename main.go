package main

import (
	"encoding/json"
	"finalproject/functions"
	"finalproject/handlers"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	_ "modernc.org/sqlite"
)

const (
	defaultPort = "7540"
	webDir      = "./web"
)

func main() {

	db, err := functions.InitDB()
	if err != nil {
		log.Fatal("Ошибка инициализации БД:", err)
	}

	defer db.Close()

	r := mux.NewRouter()

	r.HandleFunc("/api/nextdate", handlers.NextDateHandler)
	r.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.AddTaskHandler(w, r, db)
		case http.MethodGet:
			handlers.ReturnTaskHandler(w, r, db)
		case http.MethodPut:
			handlers.UpdateTaskHandler(w, r, db)
		case http.MethodDelete:
			handlers.DeleteTaskTaskHandler(w, r, db)
		default:
			w.WriteHeader((http.StatusMethodNotAllowed))
			json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
			return
		}
	})
	r.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpcomingTaskHandler(w, r, db)
	})
	r.HandleFunc("/api/task/done", func(w http.ResponseWriter, r *http.Request) {
		handlers.CompletTaskHandler(w, r, db)
	})

	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	log.Printf("Запускаем сервер на порту %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
