package backend

import (
	"database/sql"

	"net/http"
	"strconv"

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
