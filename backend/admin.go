package backend

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var db *sql.DB
var tmpl = template.Must(template.ParseGlob("frontend/*.html"))

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем список спектаклей для отображения
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

	// Получаем список афиш для отображения
	posterRows, err := db.Query(`
		SELECT p.id_poster, pf.performance_title, p.start_time 
		FROM poster p
		JOIN performance pf ON p.id_performance = pf.id_performance
		ORDER BY p.start_time`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer posterRows.Close()

	var posters []struct {
		ID        int
		Title     string
		StartTime time.Time
	}
	for posterRows.Next() {
		var p struct {
			ID        int
			Title     string
			StartTime time.Time
		}
		err := posterRows.Scan(&p.ID, &p.Title, &p.StartTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		posters = append(posters, p)
	}

	// Передаем данные в шаблон
	data := struct {
		Performances []struct {
			ID    int
			Title string
		}
		Posters []struct {
			ID        int
			Title     string
			StartTime time.Time
		}
	}{
		Performances: performances,
		Posters:      posters,
	}

	tmpl.ExecuteTemplate(w, "admin.html", data)
}

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

func PostersHandler(w http.ResponseWriter, r *http.Request) {
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
func BookingHandler(w http.ResponseWriter, r *http.Request) {
	posterID := r.URL.Query().Get("poster_id")
	if posterID == "" {
		http.Error(w, "Poster ID is required", http.StatusBadRequest)
		return
	}

	// Получаем информацию о мероприятии
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

	// Получаем данные о местах
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

	// Собираем все данные для шаблона
	data := struct {
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
	}

	tmpl.ExecuteTemplate(w, "booking.html", data)
}
func BookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	posterID := r.FormValue("poster_id")
	fullName := r.FormValue("full_name")
	email := r.FormValue("email")
	seats := r.FormValue("selected_seats")

	if posterID == "" || fullName == "" || seats == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Проверяем, что места еще свободны
	var availableSeats int
	err := db.QueryRow(`
        SELECT COUNT(*) FROM ticket 
        WHERE id_poster = $1 AND seat = ANY(string_to_array($2, ',')::int[]) 
        AND ticket_owner_full_name IS NULL`, posterID, seats).Scan(&availableSeats)

	if err != nil || availableSeats != len(strings.Split(seats, ",")) {
		http.Error(w, "Некоторые места уже заняты. Пожалуйста, обновите страницу и выберите другие места.", http.StatusConflict)
		return
	}

	// Бронируем места
	_, err = db.Exec(`
        UPDATE ticket 
        SET ticket_owner_full_name = $1, booking_time = NOW()
        WHERE id_poster = $2 AND seat = ANY(string_to_array($3, ',')::int[])`,
		fullName, posterID, seats)

	if err != nil {
		http.Error(w, "Ошибка при бронировании: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Генерируем номер заказа
	orderID := fmt.Sprintf("ORD-%d-%s", time.Now().Unix(), posterID)

	// Отправляем подтверждение
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
        <h1>Бронирование успешно завершено!</h1>
        <p>Номер вашего заказа: <strong>%s</strong></p>
        <p>Забронированные места: %s</p>
        <p>На имя: %s</p>
        %s
        <p><a href="/">Вернуться на главную</a></p>
    `, orderID, seats, fullName,
		func() string {
			if email != "" {
				return fmt.Sprintf(`<p>На email %s отправлено подтверждение.</p>`, email)
			}
			return ""
		}())
}

func CreateSceneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		name := r.FormValue("name")
		capacity := r.FormValue("capacity")
		address := r.FormValue("address")

		_, err := db.Exec(
			"INSERT INTO scene (scene_name, capacity, address) VALUES ($1, $2, $3)",
			name, capacity, address)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		tmpl.ExecuteTemplate(w, "create_scene.html", nil)
	}
}

func CreateActorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		fullName := r.FormValue("full_name")
		birthDate := r.FormValue("birth_date")
		bio := r.FormValue("bio")

		_, err := db.Exec(
			"INSERT INTO actor (actor_full_name, birth_date, bio) VALUES ($1, $2, $3)",
			fullName, birthDate, bio)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		tmpl.ExecuteTemplate(w, "create_actor.html", nil)
	}
}

func CreatePosterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Получаем списки для выпадающих меню
		performances, _ := db.Query("SELECT id_performance, performance_title FROM performance")
		scenes, _ := db.Query("SELECT id_scene, scene_name FROM scene")

		data := struct {
			Performances *sql.Rows
			Scenes       *sql.Rows
		}{
			Performances: performances,
			Scenes:       scenes,
		}

		tmpl.ExecuteTemplate(w, "create_poster.html", data)
	} else if r.Method == "POST" {
		performanceID := r.FormValue("performance_id")
		sceneID := r.FormValue("scene_id")
		startTime := r.FormValue("start_time")
		basePrice := r.FormValue("base_price")

		_, err := db.Exec(
			"INSERT INTO poster (id_performance, id_scene, start_time, base_price) VALUES ($1, $2, $3, $4)",
			performanceID, sceneID, startTime, basePrice)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func CreatePerformanceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Обработка отправки формы
		title := r.FormValue("title")
		description := r.FormValue("description")
		duration := r.FormValue("duration")
		ageRating := r.FormValue("age_rating")

		// Валидация данных
		if title == "" {
			http.Error(w, "Название спектакля обязательно", http.StatusBadRequest)
			return
		}

		// Преобразование duration в число
		durationInt, err := strconv.Atoi(duration)
		if err != nil {
			http.Error(w, "Неверный формат продолжительности", http.StatusBadRequest)
			return
		}

		// Вставка в базу данных
		_, err = db.Exec(
			"INSERT INTO performance (performance_title, description, duration_minutes, age_rating) VALUES ($1, $2, $3, $4)",
			title, description, durationInt, ageRating)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Перенаправление обратно в админку
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		// Отображение формы
		tmpl.ExecuteTemplate(w, "create_performance.html", nil)
	}
}

func GenerateTicketsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Обработка отправки формы
		posterID := r.FormValue("poster_id")
		seatsCountStr := r.FormValue("seats_count")
		pricePattern := r.FormValue("price_pattern") // "uniform" или "gradient"
		minPriceStr := r.FormValue("min_price")
		maxPriceStr := r.FormValue("max_price")

		// Валидация и преобразование данных
		if posterID == "" || seatsCountStr == "" {
			http.Error(w, "Необходимо указать афишу и количество мест", http.StatusBadRequest)
			return
		}

		seatsCount, err := strconv.Atoi(seatsCountStr)
		if err != nil {
			http.Error(w, "Неверный формат количества мест", http.StatusBadRequest)
			return
		}

		minPrice, err := strconv.Atoi(minPriceStr)
		if err != nil {
			http.Error(w, "Неверный формат минимальной цены", http.StatusBadRequest)
			return
		}

		maxPrice, err := strconv.Atoi(maxPriceStr)
		if err != nil {
			http.Error(w, "Неверный формат максимальной цены", http.StatusBadRequest)
			return
		}

		// Проверка валидности цен
		if minPrice <= 0 || maxPrice <= 0 {
			http.Error(w, "Цены должны быть положительными", http.StatusBadRequest)
			return
		}

		if maxPrice < minPrice {
			http.Error(w, "Максимальная цена не может быть меньше минимальной", http.StatusBadRequest)
			return
		}

		// Генерация билетов
		if pricePattern == "uniform" {
			// Все билеты по одной цене
			for i := 1; i <= seatsCount; i++ {
				_, err := db.Exec("INSERT INTO ticket (id_poster, seat, price) VALUES ($1, $2, $3)",
					posterID, i, minPrice)
				if err != nil {
					http.Error(w, "Ошибка при создании билетов: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
		} else {
			// Билеты с градиентом цен
			if seatsCount > 1 {
				priceStep := (maxPrice - minPrice) / (seatsCount - 1)
				for i := 1; i <= seatsCount; i++ {
					price := minPrice + (i-1)*priceStep
					_, err := db.Exec("INSERT INTO ticket (id_poster, seat, price) VALUES ($1, $2, $3)",
						posterID, i, price)
					if err != nil {
						http.Error(w, "Ошибка при создании билетов: "+err.Error(), http.StatusInternalServerError)
						return
					}
				}
			} else {
				// Для одного билета используем минимальную цену
				_, err := db.Exec("INSERT INTO ticket (id_poster, seat, price) VALUES ($1, 1, $2)",
					posterID, minPrice)
				if err != nil {
					http.Error(w, "Ошибка при создании билета: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		// Перенаправление обратно в админку
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		// Получение списка афиш для выпадающего меню
		posters, err := db.Query(`
			SELECT p.id_poster, pf.performance_title, p.start_time 
			FROM poster p
			JOIN performance pf ON p.id_performance = pf.id_performance
			ORDER BY p.start_time`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer posters.Close()

		var posterList []struct {
			ID        int
			Title     string
			StartTime time.Time
		}
		for posters.Next() {
			var p struct {
				ID        int
				Title     string
				StartTime time.Time
			}
			if err := posters.Scan(&p.ID, &p.Title, &p.StartTime); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			posterList = append(posterList, p)
		}

		// Отображение формы
		tmpl.ExecuteTemplate(w, "generate_tickets.html", posterList)
	}
}
