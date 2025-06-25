package backend

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
)

// TemplateData структура для передачи данных в шаблон
type TemplateData struct {
	Mode   string
	Actors []Actor
	Actor  Actor
}

// ActorsHandler универсальный обработчик для всех операций с актерами
func ActorsHandler(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	id := r.URL.Query().Get("id")

	tmpl, err := template.ParseFiles("frontend/actors_management.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	switch mode {
	case "create":
		handleCreateMode(w, r, tmpl)
	case "edit":
		handleEditMode(w, r, tmpl, id)
	default:
		handleListMode(w, r, tmpl)
	}
}

func handleListMode(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	actors, err := getAllActors()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := TemplateData{
		Mode:   "list",
		Actors: actors,
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func handleCreateMode(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	if r.Method == http.MethodPost {
		fullName := r.FormValue("full_name")
		troupe := r.FormValue("troupe")

		if fullName == "" || troupe == "" {
			data := TemplateData{
				Mode: "create",
			}
			tmpl.Execute(w, data)
			return
		}

		_, err := db.Exec("INSERT INTO actor (actor_full_name, troupe) VALUES ($1, $2)", fullName, troupe)
		if err != nil {
			data := TemplateData{
				Mode: "create",
			}
			tmpl.Execute(w, data)
			return
		}

		http.Redirect(w, r, "/admin/actors?mode=list", http.StatusSeeOther)
		return
	}

	data := TemplateData{Mode: "create"}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func handleEditMode(w http.ResponseWriter, r *http.Request, tmpl *template.Template, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid actor ID", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		fullName := r.FormValue("full_name")
		troupe := r.FormValue("troupe")

		if fullName == "" || troupe == "" {
			data := TemplateData{
				Mode: "edit",
			}
			tmpl.Execute(w, data)
			return
		}

		_, err := db.Exec("UPDATE actor SET actor_full_name=$1, troupe=$2 WHERE id_actor=$3", fullName, troupe, id)
		if err != nil {
			data := TemplateData{
				Mode: "edit",
			}
			tmpl.Execute(w, data)
			return
		}

		http.Redirect(w, r, "/admin/actors?mode=list", http.StatusSeeOther)
		return
	}

	actor, err := getActorByID(id)
	if err != nil {
		http.Error(w, "Actor not found", http.StatusNotFound)
		return
	}

	data := TemplateData{
		Mode:  "edit",
		Actor: actor,
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

// Вспомогательные функции для работы с БД
func getAllActors() ([]Actor, error) {
	rows, err := db.Query("SELECT id_actor, actor_full_name, troupe FROM actor ORDER BY actor_full_name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actors []Actor
	for rows.Next() {
		var actor Actor
		err := rows.Scan(&actor.ID, &actor.FullName, &actor.Troupe)
		if err != nil {
			return nil, err
		}
		actors = append(actors, actor)
	}
	return actors, nil
}

func getActorByID(id int) (Actor, error) {
	var actor Actor
	err := db.QueryRow("SELECT id_actor, actor_full_name, troupe FROM actor WHERE id_actor = $1", id).
		Scan(&actor.ID, &actor.FullName, &actor.Troupe)
	return actor, err
}

// DeleteActorHandler обработчик удаления актера
func DeleteActorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "Invalid actor ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM actor WHERE id_actor = $1", id)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/actors?mode=list", http.StatusSeeOther)
}
