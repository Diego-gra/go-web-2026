package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Game struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Year   int      `json:"year"`
	Studio string   `json:"studio"`
	Genres []string `json:"genres"`
}

type Tournament struct {
	ID     int     `json:"id"`
	GameID int     `json:"gameId"`
	Name   string  `json:"name"`
	City   string  `json:"city"`
	Date   string  `json:"date"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API HTTP %d : %s", e.StatusCode, e.Message)
}

func newAPIClient() *http.Client {
	return &http.Client{
		Timeout: 3 * time.Second,
	}
}

func getJSON[T any](client *http.Client, endpoint string) (T, error) {
	var zero T

	resp, err := client.Get(endpoint)
	if err != nil {
		return zero, fmt.Errorf("appel à %s impossible : %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return zero, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return zero, fmt.Errorf("JSON invalide sur %s : %w", endpoint, err)
	}

	return result, nil
}

func FetchGames(client *http.Client, baseURL string, query string) ([]Game, error) {
	endpoint := baseURL + "/api/games"
	if query != "" {
		endpoint += "?name=" + url.QueryEscape(query)
	}
	return getJSON[[]Game](client, endpoint)
}

func FetchGameByID(client *http.Client, baseURL string, id int) (Game, error) {
	return getJSON[Game](client, fmt.Sprintf("%s/api/games/%d", baseURL, id))
}

func FetchTournaments(client *http.Client, baseURL string) ([]Tournament, error) {
	return getJSON[[]Tournament](client, baseURL+"/api/tournaments")
}

func FetchTournamentsByGameID(client *http.Client, baseURL string, gameID int) ([]Tournament, error) {
	return getJSON[[]Tournament](client, fmt.Sprintf("%s/api/games/%d/tournaments", baseURL, gameID))
}
