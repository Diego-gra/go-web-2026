package catalog

import (
	"errors"
	"strings"

	"gamedex/internal/api"
)

var ErrNotFound = errors.New("jeu introuvable")

type Catalog struct {
	games           []api.Game
	gamesByID       map[int]api.Game
	tournamentsByID map[int][]api.Tournament
}

func New(games []api.Game, tournaments []api.Tournament) *Catalog {
	byID := make(map[int]api.Game, len(games))
	for _, g := range games {
		byID[g.ID] = g
	}

	byGame := make(map[int][]api.Tournament)
	for _, t := range tournaments {
		byGame[t.GameID] = append(byGame[t.GameID], t)
	}

	return &Catalog{
		games:           games,
		gamesByID:       byID,
		tournamentsByID: byGame,
	}
}

func (c *Catalog) SearchGames(query string) []api.Game {
	q := strings.TrimSpace(strings.ToLower(query))
	if q == "" {
		return c.games
	}

	var filtered []api.Game
	for _, g := range c.games {
		if strings.Contains(strings.ToLower(g.Name), q) ||
			strings.Contains(strings.ToLower(g.Studio), q) {
			filtered = append(filtered, g)
		}
	}
	return filtered
}

func (c *Catalog) GameWithTournaments(id int) (api.Game, []api.Tournament, error) {
	g, ok := c.gamesByID[id]
	if !ok {
		return api.Game{}, nil, ErrNotFound
	}
	return g, c.tournamentsByID[id], nil
}
