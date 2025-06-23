package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB
var tmpl = template.Must(template.ParseGlob("frontend/*.html"))

func main() {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "1234")
	dbname := getEnv("DB_NAME", "kola")

	dbConn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", dbConn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/performances", performancesHandler)
	http.HandleFunc("/posters", postersHandler)
	http.HandleFunc("/tickets", ticketsHandler)
	http.HandleFunc("/book", bookHandler)
	http.HandleFunc("/actor_plays", actorPlaysHandler)
	http.HandleFunc("/schedule", scheduleHandler)

	port = getEnv("PORT", "8080")
	log.Printf("Server running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "index.html", nil)
}

func performancesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id_performance, performance_title FROM performance")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var performances []struct {
		ID    int
		Title string
	}
	for rows.Next() {
		var p struct {
			ID    int
			Title string
		}
		err := rows.Scan(&p.ID, &p.Title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		performances = append(performances, p)
	}

	tmpl.ExecuteTemplate(w, "performances.html", performances)
}

func postersHandler(w http.ResponseWriter, r *http.Request) {
	performanceID := r.URL.Query().Get("performance_id")
	if performanceID == "" {
		http.Error(w, "Performance ID is required", http.StatusBadRequest)
		return
	}

	query := `
		SELECT p.id_poster, p.start_time, s.scene_name 
		FROM poster p
		JOIN scene s ON p.id_scene = s.id_scene
		WHERE p.id_performance = $1
		ORDER BY p.start_time`

	rows, err := db.Query(query, performanceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posters []struct {
		ID        int
		StartTime time.Time
		SceneName string
	}
	for rows.Next() {
		var p struct {
			ID        int
			StartTime time.Time
			SceneName string
		}
		err := rows.Scan(&p.ID, &p.StartTime, &p.SceneName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		posters = append(posters, p)
	}

	tmpl.ExecuteTemplate(w, "posters.html", struct {
		PerformanceID string
		Posters       []struct {
			ID        int
			StartTime time.Time
			SceneName string
		}
	}{
		PerformanceID: performanceID,
		Posters:       posters,
	})
}

func ticketsHandler(w http.ResponseWriter, r *http.Request) {
	posterID := r.URL.Query().Get("poster_id")
	if posterID == "" {
		http.Error(w, "Poster ID is required", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id_ticket, actor_role_name, price, seat, ticket_owner_full_name 
		FROM ticket 
		WHERE id_poster = $1 AND ticket_owner_full_name IS NULL
		ORDER BY seat`

	rows, err := db.Query(query, posterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tickets []struct {
		ID       int
		RoleName string
		Price    int
		Seat     int
		Owner    sql.NullString
	}
	for rows.Next() {
		var t struct {
			ID       int
			RoleName string
			Price    int
			Seat     int
			Owner    sql.NullString
		}
		err := rows.Scan(&t.ID, &t.RoleName, &t.Price, &t.Seat, &t.Owner)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tickets = append(tickets, t)
	}

	tmpl.ExecuteTemplate(w, "tickets.html", struct {
		PosterID string
		Tickets  []struct {
			ID       int
			RoleName string
			Price    int
			Seat     int
			Owner    sql.NullString
		}
	}{
		PosterID: posterID,
		Tickets:  tickets,
	})
}

func bookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		ticketID := r.FormValue("ticket_id")
		fullName := r.FormValue("full_name")

		_, err := db.Exec(`
			UPDATE ticket SET ticket_owner_full_name = $1 
			WHERE id_ticket = $2 AND ticket_owner_full_name IS NULL`, fullName, ticketID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/performances", http.StatusSeeOther)
		return
	}

	ticketID := r.URL.Query().Get("ticket_id")
	if ticketID == "" {
		http.Error(w, "Ticket ID is required", http.StatusBadRequest)
		return
	}

	var ticket struct {
		ID       int
		RoleName string
		Price    int
		Seat     int
	}
	err := db.QueryRow(`
		SELECT id_ticket, actor_role_name, price, seat 
		FROM ticket WHERE id_ticket = $1`, ticketID).
		Scan(&ticket.ID, &ticket.RoleName, &ticket.Price, &ticket.Seat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "book.html", ticket)
}

func actorPlaysHandler(w http.ResponseWriter, r *http.Request) {
	actorID := r.URL.Query().Get("actor_id")
	if actorID == "" {
		http.Error(w, "Actor ID is required", http.StatusBadRequest)
		return
	}

	query := `
		SELECT DISTINCT pl.play_title
		FROM play pl
		JOIN performance pf ON pf.id_play = pl.id_play
		JOIN performance_role pr ON pr.id_performance = pf.id_performance
		JOIN actor_role ar ON ar.performance_role_id = pr.performance_role_id
		WHERE ar.id_actor = $1`

	rows, err := db.Query(query, actorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var plays []string
	for rows.Next() {
		var title string
		err := rows.Scan(&title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		plays = append(plays, title)
	}
	tmpl.ExecuteTemplate(w, "actor_plays.html", plays)
}

func scheduleHandler(w http.ResponseWriter, r *http.Request) {
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	if start == "" || end == "" {
		http.Error(w, "Start and end dates are required", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse("2006-01-02", start)
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse("2006-01-02", end)
	if err != nil {
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}

	query := `
		SELECT p.id_poster, p.start_time, s.scene_name, pf.performance_title
		FROM poster p
		JOIN scene s ON p.id_scene = s.id_scene
		JOIN performance pf ON p.id_performance = pf.id_performance
		WHERE p.start_time BETWEEN $1 AND $2
		ORDER BY p.start_time`

	rows, err := db.Query(query, startTime, endTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var schedule []struct {
		ID        int
		StartTime time.Time
		SceneName string
		Title     string
	}
	for rows.Next() {
		var item struct {
			ID        int
			StartTime time.Time
			SceneName string
			Title     string
		}
		err := rows.Scan(&item.ID, &item.StartTime, &item.SceneName, &item.Title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		schedule = append(schedule, item)
	}
	tmpl.ExecuteTemplate(w, "schedule.html", schedule)
}
