package games

import (
	"clipet/internal/game"
	"math/rand"
)

// GameManager manages and runs mini-games.
type GameManager struct {
	games []MiniGame
}

// NewGameManager creates a new game manager.
func NewGameManager() *GameManager {
	return &GameManager{
		games: []MiniGame{
			newReactionSpeedGame(),
			newGuessNumberGame(),
		},
	}
}

// GetAvailableGames returns games that match the pet's state.
func (gm *GameManager) GetAvailableGames(pet *game.Pet) []MiniGame {
	var available []MiniGame
	
	for _, game := range gm.games {
		config := game.GetConfig()
		if pet.Energy >= config.MinEnergy {
			available = append(available, game)
		}
	}
	
	return available
}

// GetGame returns a game by type, or nil if not found.
func (gm *GameManager) GetGame(gameType GameType) MiniGame {
	for _, game := range gm.games {
		if game.GetConfig().Type == gameType {
			return game
		}
	}
	return nil
}

// PlayGame runs a mini-game with the given pet.
func (gm *GameManager) PlayGame(pet *game.Pet, gameType GameType) (*GameResult, bool) {
	game := gm.GetGame(gameType)
	if game == nil {
		return nil, false
	}
	
	config := game.GetConfig()
	
	// Deduct energy cost upfront
	cost := config.MaxEnergyCost
	if config.MaxEnergyCost > 0 {
		pet.Energy = clamp(pet.Energy-cost, 0, 100)
	}
	
	// Play the game
	result, shouldContinue := game.Play(pet)
	
	// Apply attribute changes if result has them
	if result != nil && result.AttrChange != nil {
		for attr, change := range result.AttrChange {
			switch attr {
			case "happiness":
				pet.Happiness = clamp(change[1], 0, 100)
			case "energy":
				pet.Energy = clamp(change[1], 0, 100)
			}
		}
	}
	
	return result, shouldContinue
}

// GetRandomGame returns a random available game for the pet.
func (gm *GameManager) GetRandomGame(pet *game.Pet) MiniGame {
	games := gm.GetAvailableGames(pet)
	if len(games) == 0 {
		return nil
	}
	return games[rand.Intn(len(games))]
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}