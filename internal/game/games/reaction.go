package games

import (
	"clipet/internal/game"
	"fmt"
	"math/rand"
	"time"
)

// reactionSpeedGame implements a reaction time test game.
type reactionSpeedGame struct {
	state      GameState
	targetTime time.Time
	startTime  time.Time
	result     *GameResult
	inputChan  chan string
}

// GameState defines the states for the reaction game.
type GameState int

const (
	StateWaiting GameState = iota
	StateReady
	StateRunning
	StateFinished
)

// newReactionSpeedGame creates a new reaction speed game.
func newReactionSpeedGame() MiniGame {
	return &reactionSpeedGame{
		state:     StateWaiting,
		inputChan: make(chan string, 1),
	}
}

// GetConfig returns the game's configuration.
func (g *reactionSpeedGame) GetConfig() GameConfig {
	return GameConfig{
		Type:          GameReactionSpeed,
		Name:          "反应速度测试",
		Description:   "当出现 GO! 时，尽快按键！测试你的反应速度。",
		Duration:      10 * time.Second,
		MinEnergy:     5,
		MaxEnergyCost: 8,
		WinnerEnergy:  -3, // less energy cost on win
		WinnerHappiness: 15,
		LoserHappiness:  -5,
	}
}

// Play executes the game logic.
func (g *reactionSpeedGame) Play(pet *game.Pet) (*GameResult, bool) {
	g.state = StateWaiting
	g.result = &GameResult{
		GameType:   GameReactionSpeed,
		PlayerName: "你",
		PetName:    pet.Name,
		Timestamp:  time.Now(),
		AttrChange: make(map[string][2]int),
	}

	// Random delay before showing "GO!"
	delay := time.Duration(rand.Intn(4000)+2000) * time.Millisecond // 2-6 seconds
	
	for {
		select {
		case <-time.After(100 * time.Millisecond):
			if g.state == StateWaiting {
				// Show waiting message
				fmt.Println(g.Render())
				time.Sleep(delay)
				g.state = StateReady
				g.targetTime = time.Now()
			} else if g.state == StateReady {
				g.startTime = time.Now()
				g.state = StateRunning
			} else if g.state == StateRunning {
				// Timeout
				g.finishGame(false, 0, pet)
				return g.result, false
			}
		case _ = <-g.inputChan: // Unused input variable
			if g.state == StateRunning {
				reactionTime := time.Since(g.targetTime).Milliseconds()
				g.finishGame(true, int(reactionTime), pet)
				return g.result, false
			}
		}
		
		// Check if game should continue based on state
		if g.state == StateFinished {
			return g.result, false
		}
	}
}

// finishGame ends the game and calculates results.
func (g *reactionSpeedGame) finishGame(won bool, score int, pet *game.Pet) {
	g.state = StateFinished
	g.result.Won = won
	g.result.Score = score
	
	oldH := pet.Happiness
	oldE := pet.Energy
	
	config := g.GetConfig()
	
	if won {
		// Win: less energy cost, more happiness
		pet.Energy = clamp(pet.Energy+config.WinnerEnergy, 0, 100)
		pet.Happiness = clamp(pet.Happiness+config.WinnerHappiness, 0, 100)
	} else {
		// Loss: base energy cost, less happiness
		pet.Happiness = clamp(pet.Happiness+config.LoserHappiness, 0, 100)
	}
	
	g.result.AttrChange["happiness"] = [2]int{oldH, pet.Happiness}
	g.result.AttrChange["energy"] = [2]int{oldE, pet.Energy}
	
	// Show final result
	fmt.Printf("\n%s\n", g.Render())
	if won {
		fmt.Printf("反应时间: %d 毫秒 (%s)\n", score, getReactionRating(score))
	}
	fmt.Printf("按 Enter 继续...\n")
	
	// Wait for enter to continue
	for {
		select {
		case input := <-g.inputChan:
			if input == "enter" {
				return
			}
		}
	}
}

// Render displays the game UI.
func (g *reactionSpeedGame) Render() string {
	switch g.state {
	case StateWaiting:
		return `🎮 反应速度测试 🎮

准备... 看到下面出现 GO! 时尽快按 Enter！

当前状态: 等待中...`
		
	case StateReady:
		return `🎮 反应速度测试 🎮
         ^
         |
         GO! ⚡
         |
         v
当前状态: 快速按 Enter！`
		
	case StateRunning:
		elapsed := time.Since(g.targetTime).Milliseconds()
		return fmt.Sprintf(`🎮 反应速度测试 🎮
正在计时... %d 毫秒`, elapsed)
		
	case StateFinished:
		if g.result.Won {
			return fmt.Sprintf(`🎮 游戏结束 🎮
✅ 恭喜获胜！
反应时间: %d 毫秒 (%s)
属性变化:
  快乐度: %d → %d
  精力: %d → %d`,
				g.result.Score, getReactionRating(g.result.Score),
				g.result.AttrChange["happiness"][0],
				g.result.AttrChange["happiness"][1],
				g.result.AttrChange["energy"][0],
				g.result.AttrChange["energy"][1])
		} else {
			return fmt.Sprintf(`🎮 游戏结束 🎮
❌ 超时失败！
属性变化:
  快乐度: %d → %d
  精力: %d → %d`,
				g.result.AttrChange["happiness"][0],
				g.result.AttrChange["happiness"][1],
				g.result.AttrChange["energy"][0],
				g.result.AttrChange["energy"][1])
		}
		
	default:
		return "游戏状态未知"
	}
}

// getReactionRating returns a rating based on reaction time.
func getReactionRating(ms int) string {
	if ms < 200 {
		return "超快！🚀"
	} else if ms < 300 {
		return "很快！⚡"
	} else if ms < 400 {
		return "不错！👍"
	} else if ms < 500 {
		return "一般😐"
	} else {
		return "需要练习🐌"
	}
}