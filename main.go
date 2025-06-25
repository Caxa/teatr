package main

import (
	"kola/backend"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Инициализация базы данных
	db, err := backend.OpenDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	backend.SetDB(db)

	// Регистрация обработчиков
	registerHandlers()

	port := getEnv("PORT", "8080")
	log.Printf("Server running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func registerHandlers() {
	http.HandleFunc("/", backend.IndexHandler)
	http.HandleFunc("/performances", backend.PerformancesHandler)
	http.HandleFunc("/posters", backend.PostersHandler)
	http.HandleFunc("/tickets", backend.TicketsHandler)
	http.HandleFunc("/book", backend.BookHandler)
	http.HandleFunc("/booking", backend.BookingHandler)
	http.HandleFunc("/actor_plays", backend.ActorPlaysHandler)
	http.HandleFunc("/schedule", backend.ScheduleHandler)

	// Админские обработчики
	http.HandleFunc("/admin", backend.AdminHandler)
	http.HandleFunc("/admin/create_performance", backend.CreatePerformanceHandler)
	http.HandleFunc("/admin/create_scene", backend.CreateSceneHandler)
	http.HandleFunc("/admin/create_actor", backend.CreateActorHandler)
	http.HandleFunc("/admin/create_poster", backend.CreatePosterHandler)
	http.HandleFunc("/admin/generate_tickets", backend.GenerateTicketsHandler)
	http.HandleFunc("/admin/execute-sql", backend.ExecuteSQLHandler)

}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
