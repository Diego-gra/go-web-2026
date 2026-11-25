package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type jeuxPageData struct {
	Title string
	Query string
	Games []Game
}

type jeuDetailPageData struct {
	Title       string
	Game        Game
	Tournaments []Tournament
}

type erreurPageData struct {
	Title      string
	StatusCode int
	Message    string
}

func renderErrorPage(w http.ResponseWriter, tmpl *template.Template, status int, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	data := erreurPageData{
		Title:      fmt.Sprintf("Erreur %d", status),
		StatusCode: status,
		Message:    message,
	}
	if err := tmpl.ExecuteTemplate(w, "erreur", data); err != nil {
		http.Error(w, message, status)
	}
}

func main() {
	port := flag.Int("port", 8080, "Port d'écoute du serveur GameDex")
	apiURL := flag.String("api-url", "http://localhost:8081", "URL de base de la mini API")
	flag.Parse()

	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Erreur chargement templates : %v", err)
	}

	client := newAPIClient()
	baseURL := strings.TrimRight(*apiURL, "/")

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/jeux", http.StatusFound)
	})

	// 2. & 5. /jeux avec ?q=
	mux.HandleFunc("GET /jeux", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")

		games, err := FetchGames(client, baseURL, query)
		if err != nil {
			var apiErr *APIError
			if errors.As(err, &apiErr) && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 {
				renderErrorPage(w, tmpl, apiErr.StatusCode, "Requête refusée par l'API.")
				return
			}
			renderErrorPage(w, tmpl, http.StatusServiceUnavailable, "L'API GameDex est actuellement indisponible.")
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := jeuxPageData{
			Title: "Jeux",
			Query: query,
			Games: games,
		}
		if err := tmpl.ExecuteTemplate(w, "jeux", data); err != nil {
			renderErrorPage(w, tmpl, http.StatusInternalServerError, "Erreur d'affichage de la page.")
		}
	})

	// 3. & 4. /jeux/{id} + 400 / 404 / API indisponible
	mux.HandleFunc("GET /jeux/{id}", func(w http.ResponseWriter, r *http.Request) {
		rawID := r.PathValue("id")
		id, err := strconv.Atoi(rawID)
		if err != nil {
			renderErrorPage(w, tmpl, http.StatusBadRequest, "L'identifiant du jeu doit être un nombre entier valide.")
			return
		}

		game, err := FetchGameByID(client, baseURL, id)
		if err != nil {
			var apiErr *APIError
			if errors.As(err, &apiErr) {
				if apiErr.StatusCode == http.StatusNotFound {
					renderErrorPage(w, tmpl, http.StatusNotFound, "Ce jeu n'existe pas.")
					return
				}
				if apiErr.StatusCode == http.StatusBadRequest {
					renderErrorPage(w, tmpl, http.StatusBadRequest, "Identifiant invalide.")
					return
				}
			}
			renderErrorPage(w, tmpl, http.StatusServiceUnavailable, "L'API GameDex est actuellement indisponible.")
			return
		}

		tournaments, err := FetchTournamentsByGameID(client, baseURL, id)
		if err != nil {
			renderErrorPage(w, tmpl, http.StatusServiceUnavailable, "Impossible de récupérer les tournois pour le moment.")
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := jeuDetailPageData{
			Title:       game.Name,
			Game:        game,
			Tournaments: tournaments,
		}
		if err := tmpl.ExecuteTemplate(w, "jeu_detail", data); err != nil {
			renderErrorPage(w, tmpl, http.StatusInternalServerError, "Erreur d'affichage de la fiche du jeu.")
		}
	})

	// Route de secours pour les chemins inconnus (404)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderErrorPage(w, tmpl, http.StatusNotFound, "La page demandée est introuvable.")
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Serveur GameDex démarré sur http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
