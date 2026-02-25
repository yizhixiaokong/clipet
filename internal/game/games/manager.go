package games

import "math/rand"

// randIntn 是 rand.Intn 的包内别名，方便测试时替换。
var randIntn = rand.Intn

// GameManager 管理和创建迷你游戏实例。
type GameManager struct {
	registry map[GameType]func() MiniGame
}

// NewGameManager 创建游戏管理器，注册所有可用游戏。
func NewGameManager() *GameManager {
	return &GameManager{
		registry: map[GameType]func() MiniGame{
			GameReactionSpeed: func() MiniGame { return newReactionSpeedGame() },
			GameGuessNumber:   func() MiniGame { return newGuessNumberGame() },
		},
	}
}

// NewGame 创建指定类型的新游戏实例（每次返回全新实例）。
func (gm *GameManager) NewGame(gt GameType) MiniGame {
	factory, ok := gm.registry[gt]
	if !ok {
		return nil
	}
	return factory()
}

// GetConfig 返回指定游戏类型的配置。
func (gm *GameManager) GetConfig(gt GameType) (GameConfig, bool) {
	g := gm.NewGame(gt)
	if g == nil {
		return GameConfig{}, false
	}
	return g.GetConfig(), true
}

// AvailableGames 返回所有已注册的游戏类型。
func (gm *GameManager) AvailableGames() []GameType {
	types := make([]GameType, 0, len(gm.registry))
	for gt := range gm.registry {
		types = append(types, gt)
	}
	return types
}
