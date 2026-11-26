package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func getJSON[T any](c *Client, endpoint string) (T, error) {
	var zero T

	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return zero, fmt.Errorf("requête vers %s impossible : %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return zero, fmt.Errorf("statut HTTP %d sur %s : %s", resp.StatusCode, endpoint, string(body))
	}

	var data T
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return zero, fmt.Errorf("JSON invalide sur %s : %w", endpoint, err)
	}

	return data, nil
}

func (c *Client) FetchGames() ([]Game, error) {
	return getJSON[[]Game](c, c.baseURL+"/api/games")
}

func (c *Client) FetchTournaments() ([]Tournament, error) {
	return getJSON[[]Tournament](c, c.baseURL+"/api/tournaments")
}
