package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"gamedex/internal/api"
	"gamedex/internal/catalog"
	"gamedex/internal/server"
)

func main() {
	port := flag.Int("port", 8080, "Port d'écoute du serveur GameDex")
	apiURL := flag.String("api-url", "http://localhost:8081", "URL de base de la mini API")
	webRoot := flag.String("web", "web", "Dossier contenant static/ et templates/")
	flag.Parse()

	client := api.NewClient(*apiURL, 3*time.Second)

	games, err := client.FetchGames()
	if err != nil {
		log.Fatalf("Impossible de charger les jeux : %v", err)
	}

	tournaments, err := client.FetchTournaments()
	if err != nil {
		log.Fatalf("Impossible de charger les tournois : %v", err)
	}

	cat := catalog.New(games, tournaments)

	srv, err := server.New(cat, *webRoot)
	if err != nil {
		log.Fatalf("Impossible d'initialiser le serveur : %v", err)
	}

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("GameDex prêt sur http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
