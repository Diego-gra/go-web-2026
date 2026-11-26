package server

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"gamedex/internal/api"
	"gamedex/internal/catalog"
)

type Server struct {
	cat     *catalog.Catalog
	tmpl    *template.Template
	webRoot string
}

type jeuxPageData struct {
	Title string
	Query string
	Games []api.Game
}

type jeuDetailPageData struct {
	Title       string
	Game        api.Game
	Tournaments []api.Tournament
}

type erreurPageData struct {
	Title      string
	StatusCode int
	Message    string
}

func New(cat *catalog.Catalog, webRoot string) (*Server, error) {
	tmpl, err := template.ParseGlob(webRoot + "/templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("chargement templates impossible : %w", err)
	}

	return &Server{
		cat:     cat,
		tmpl:    tmpl,
		webRoot: webRoot,
	}, nil
}

func (s *Server) renderError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	data := erreurPageData{
		Title:      fmt.Sprintf("Erreur %d", status),
		StatusCode: status,
		Message:    message,
	}
	if err := s.tmpl.ExecuteTemplate(w, "erreur", data); err != nil {
		http.Error(w, message, status)
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir(s.webRoot + "/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/jeux", http.StatusFound)
	})

	mux.HandleFunc("GET /jeux", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		games := s.cat.SearchGames(query)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := jeuxPageData{
			Title: "Jeux",
			Query: query,
			Games: games,
		}
		if err := s.tmpl.ExecuteTemplate(w, "jeux", data); err != nil {
			s.renderError(w, http.StatusInternalServerError, "Erreur d'affichage de la page.")
		}
	})

	mux.HandleFunc("GET /jeux/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			s.renderError(w, http.StatusBadRequest, "L'identifiant du jeu doit être un entier valide.")
			return
		}

		game, tournaments, err := s.cat.GameWithTournaments(id)
		if err != nil {
			if errors.Is(err, catalog.ErrNotFound) {
				s.renderError(w, http.StatusNotFound, "Ce jeu n'existe pas.")
				return
			}
			s.renderError(w, http.StatusInternalServerError, "Erreur interne.")
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := jeuDetailPageData{
			Title:       game.Name,
			Game:        game,
			Tournaments: tournaments,
		}
		if err := s.tmpl.ExecuteTemplate(w, "jeu_detail", data); err != nil {
			s.renderError(w, http.StatusInternalServerError, "Erreur d'affichage de la fiche du jeu.")
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.renderError(w, http.StatusNotFound, "Page introuvable.")
	})

	return mux
}
