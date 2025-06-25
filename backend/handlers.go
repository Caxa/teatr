package backend

import (
	"database/sql"
	"fmt"
	"strconv"
	"text/template"

	"net/http"
	"strings"
	"time"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "index.html", nil)
}

func PerformancesHandler(w http.ResponseWriter, r *http.Request) {
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

func TicketsHandler(w http.ResponseWriter, r *http.Request) {
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

func ActorPlaysHandler(w http.ResponseWriter, r *http.Request) {
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

func ScheduleHandler(w http.ResponseWriter, r *http.Request) {
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

// Обработчик страницы бронирования
func BookingPageHandler(w http.ResponseWriter, r *http.Request) {
	posterID := r.URL.Query().Get("id")
	if posterID == "" {
		http.Error(w, "Не указан ID мероприятия", http.StatusBadRequest)
		return
	}

	// Получаем минимальную цену для мероприятия
	var minPrice int
	err := db.QueryRow(`
		SELECT MIN(price) 
		FROM ticket 
		WHERE id_poster = $1`, posterID).Scan(&minPrice)
	if err != nil {
		http.Error(w, "Ошибка получения данных о мероприятии", http.StatusInternalServerError)
		return
	}

	// Формируем данные для шаблона
	data := struct {
		PosterID string
		MinPrice int
	}{
		PosterID: posterID,
		MinPrice: minPrice,
	}

	// Рендерим шаблон
	tmpl := template.Must(template.ParseFiles("booking_simple.html"))
	tmpl.Execute(w, data)
}

func BookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	posterID := r.FormValue("poster_id")
	fullName := strings.TrimSpace(r.FormValue("full_name"))
	seatStr := r.FormValue("seat")

	if posterID == "" || fullName == "" || seatStr == "" {
		http.Error(w, "Не все обязательные поля заполнены", http.StatusBadRequest)
		return
	}

	seat, err := strconv.Atoi(seatStr)
	if err != nil || seat <= 0 {
		http.Error(w, "Некорректный номер места", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли уже билет с таким местом для этого мероприятия
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM ticket WHERE id_poster = $1 AND seat = $2
		)
	`, posterID, seat).Scan(&exists)

	if err != nil {
		http.Error(w, "Ошибка проверки билета", http.StatusInternalServerError)
		return
	}

	if exists {
		// Обновляем владельца, если место свободно
		var isFree bool
		err = db.QueryRow(`
			SELECT ticket_owner_full_name IS NULL 
			FROM ticket 
			WHERE id_poster = $1 AND seat = $2
		`, posterID, seat).Scan(&isFree)

		if err != nil {
			http.Error(w, "Ошибка проверки статуса места", http.StatusInternalServerError)
			return
		}

		if !isFree {
			http.Error(w, "Место уже занято", http.StatusConflict)
			return
		}

		_, err = db.Exec(`
			UPDATE ticket 
			SET ticket_owner_full_name = $1
			WHERE id_poster = $2 AND seat = $3
		`, fullName, posterID, seat)
	} else {
		// Получаем минимальную цену на билеты для указанного мероприятия
		var price int
		err = db.QueryRow(`
			SELECT MIN(price) FROM ticket WHERE id_poster = $1
		`, posterID).Scan(&price)

		if err != nil {
			http.Error(w, "Ошибка получения цены", http.StatusInternalServerError)
			return
		}

		// Добавляем новый билет
		_, err = db.Exec(`
			INSERT INTO ticket (price, seat, ticket_owner_full_name, id_poster)
			VALUES ($1, $2, $3, $4)
		`, price, seat, fullName, posterID)
	}

	if err != nil {
		http.Error(w, "Ошибка при бронировании", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/success?seat=%d&name=%s", seat, fullName), http.StatusSeeOther)
}

func BookingHandler(w http.ResponseWriter, r *http.Request) {
	posterID := r.URL.Query().Get("poster_id")
	if posterID == "" {
		http.Error(w, "Poster ID is required", http.StatusBadRequest)
		return
	}

	var eventInfo struct {
		PosterID         string
		PerformanceTitle string
		StartTime        time.Time
		SceneName        string
		MinPrice         int
		MaxPrice         int
	}

	err := db.QueryRow(`
        SELECT p.id_poster, pf.performance_title, p.start_time, s.scene_name,
               MIN(t.price) as min_price, MAX(t.price) as max_price
        FROM poster p
        JOIN performance pf ON p.id_performance = pf.id_performance
        JOIN scene s ON p.id_scene = s.id_scene
        JOIN ticket t ON p.id_poster = t.id_poster
        WHERE p.id_poster = $1
        GROUP BY p.id_poster, pf.performance_title, p.start_time, s.scene_name`, posterID).
		Scan(&eventInfo.PosterID, &eventInfo.PerformanceTitle, &eventInfo.StartTime,
			&eventInfo.SceneName, &eventInfo.MinPrice, &eventInfo.MaxPrice)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := db.Query(`
        SELECT seat, price, ticket_owner_full_name as owner_name
        FROM ticket 
        WHERE id_poster = $1
        ORDER BY seat`, posterID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var seats []struct {
		Seat      int
		Price     int
		OwnerName string
		IsFree    bool
	}

	for rows.Next() {
		var s struct {
			Seat      int
			Price     int
			OwnerName string
			IsFree    bool
		}
		if err := rows.Scan(&s.Seat, &s.Price, &s.OwnerName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.IsFree = (s.OwnerName == "")
		seats = append(seats, s)
	}

	tmpl.ExecuteTemplate(w, "booking.html", struct {
		PosterID         string
		PerformanceTitle string
		StartTime        time.Time
		SceneName        string
		MinPrice         int
		MaxPrice         int
		Seats            []struct {
			Seat      int
			Price     int
			OwnerName string
			IsFree    bool
		}
	}{
		PosterID:         eventInfo.PosterID,
		PerformanceTitle: eventInfo.PerformanceTitle,
		StartTime:        eventInfo.StartTime,
		SceneName:        eventInfo.SceneName,
		MinPrice:         eventInfo.MinPrice,
		MaxPrice:         eventInfo.MaxPrice,
		Seats:            seats,
	})
}
func PostersHandler(w http.ResponseWriter, r *http.Request) {
	performanceID := r.URL.Query().Get("performance_id")
	if performanceID == "" {
		http.Error(w, "Performance ID is required", http.StatusBadRequest)
		return
	}

	// Получаем название спектакля для заголовка
	var performanceTitle string
	err := db.QueryRow("SELECT performance_title FROM performance WHERE id_performance = $1", performanceID).Scan(&performanceTitle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	// Добавляем название спектакля в данные для шаблона
	tmpl.ExecuteTemplate(w, "posters.html", struct {
		PerformanceID    string
		PerformanceTitle string
		Posters          []struct {
			ID        int
			StartTime time.Time
			SceneName string
		}
	}{
		PerformanceID:    performanceID,
		PerformanceTitle: performanceTitle,
		Posters:          posters,
	})
}
