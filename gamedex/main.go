package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func main() {
	port := flag.Int("port", 8080, "Port d'écoute du serveur")
	flag.Parse()

	client := newAPIClient()
	baseURL := "http://localhost:8081"

	games, err := FetchGames(client, baseURL)
	if err != nil {
		log.Fatalf("Erreur FetchGames : %v", err)
	}

	tournaments, err := FetchTournaments(client, baseURL)
	if err != nil {
		log.Fatalf("Erreur FetchTournaments : %v", err)
	}

	counts := make(map[int]int)
	for _, t := range tournaments {
		counts[t.GameID]++
	}

	for _, g := range games {
		fmt.Printf("%s (%d) - %d tournoi(s)\n", g.Name, g.Year, counts[g.ID])
	}

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!DOCTYPE html>
<html lang="fr">
<head>
	<meta charset="UTF-8">
	<title>GameDex</title>
	<link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
	<h1>Bienvenue sur GameDex</h1>
	<p><a href="/jeux">Voir la liste des jeux</a></p>
</body>
</html>`)
	})

	mux.HandleFunc("GET /jeux", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		for _, g := range games {
			fmt.Fprintf(w, "[%d] %s (%d)\n", g.ID, g.Name, g.Year)
		}
	})

	mux.HandleFunc("GET /jeux/{id}", func(w http.ResponseWriter, r *http.Request) {
		rawID := r.PathValue("id")
		id, err := strconv.Atoi(rawID)
		if err != nil {
			http.Error(w, "ID invalide", http.StatusBadRequest)
			return
		}

		for _, g := range games {
			if g.ID == id {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				fmt.Fprintf(w, "Jeu #%d : %s (%d) - %d tournoi(s)\n", g.ID, g.Name, g.Year, counts[g.ID])
				return
			}
		}

		http.Error(w, "Jeu introuvable", http.StatusNotFound)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Serveur GameDex démarré sur http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
