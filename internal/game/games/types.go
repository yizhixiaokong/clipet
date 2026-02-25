package games

import (
	"clipet/internal/game"
	"time"
)

// GameType represents the type of mini-game.
type GameType string

const (
	GameReactionSpeed GameType = "reaction_speed"
	GameGuessNumber   GameType = "guess_number"
)

// GameResult holds the outcome of a mini-game.
type GameResult struct {
	GameType    GameType
	Won         bool
	Score       int // game-specific score (time for reaction, number of attempts for guess)
	PlayerName  string
	PetName     string
	Timestamp   time.Time
	AttrChange map[string][2]int // attribute name -> {old, new}
}

// GameConfig defines configuration for a mini-game.
type GameConfig struct {
	Type          GameType
	Name          string
	Description   string
	Duration      time.Duration // max time allowed
	MinEnergy     int          // required energy to play
	MaxEnergyCost int          // max energy cost on loss
	WinnerEnergy  int          // energy cost reduction on win
	WinnerHappiness int        // happiness gain on win
	LoserHappiness  int        // happiness change on loss
}

// MiniGame interface defines the contract for all mini-games.
type MiniGame interface {
	// GetConfig returns the game's configuration.
	GetConfig() GameConfig
	
	// Play executes the game logic.
	// Returns result and whether the player should continue playing.
	Play(*game.Pet) (*GameResult, bool)
	
	// Render displays the game UI in terminal.
	Render() string
}