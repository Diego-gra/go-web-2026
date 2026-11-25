package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func main() {
	// Étape 5 : option -port avec le package flag
	port := flag.Int("port", 8080, "Port d'écoute du serveur")
	flag.Parse()

	mux := http.NewServeMux()

	// Données fictives pour tester /jeux/{id}
	jeux := map[int]string{
		1: "The Legend of Zelda: Breath of the Wild",
		2: "Hades",
		3: "Celeste",
	}

	// Bonus : page d'accueil en HTML qui utilise style.css
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
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

	// Étape 2 : GET /jeux renvoie du texte
	mux.HandleFunc("GET /jeux", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "1 - The Legend of Zelda: Breath of the Wild")
		fmt.Fprintln(w, "2 - Hades")
		fmt.Fprintln(w, "3 - Celeste")
	})

	// Étape 2 & 3 : GET /jeux/{id} avec gestion du 400 et du 404
	mux.HandleFunc("GET /jeux/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")

		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			http.Error(w, "ID invalide (400 Bad Request)", http.StatusBadRequest)
			return
		}

		nomJeu, ok := jeux[id]
		if !ok {
			http.Error(w, "Jeu introuvable (404 Not Found)", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "Jeu #%d : %s\n", id, nomJeu)
	})

	// Étape 4 : servir static/css/style.css
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Serveur lancé sur http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
