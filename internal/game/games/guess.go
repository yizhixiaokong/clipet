package games

import (
	"clipet/internal/game"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// guessNumberGame implements a number guessing game.
type guessNumberGame struct {
	state       GameState
	targetNum   int
	attempts    int
	maxAttempts int
	result      *GameResult
	inputChan   chan string
}

const (
	maxGuessAttempts = 7
)

// newGuessNumberGame creates a new guess number game.
func newGuessNumberGame() MiniGame {
	return &guessNumberGame{
		state:     StateWaiting,
		maxAttempts: maxGuessAttempts,
		inputChan: make(chan string, 1),
	}
}

// GetConfig returns the game's configuration.
func (g *guessNumberGame) GetConfig() GameConfig {
	return GameConfig{
		Type:          GameGuessNumber,
		Name:          "猜数字",
		Description:   "我想了一个1-100的数字，你能猜中吗？最多7次机会！",
		Duration:      30 * time.Second,
		MinEnergy:     3,
		MaxEnergyCost: 5,
		WinnerEnergy:  -1, // less energy cost on win
		WinnerHappiness: 20,
		LoserHappiness:  -8,
	}
}

// Play executes the game logic.
func (g *guessNumberGame) Play(pet *game.Pet) (*GameResult, bool) {
	g.state = StateWaiting
	g.result = &GameResult{
		GameType:   GameGuessNumber,
		PlayerName: "你",
		PetName:    pet.Name,
		Timestamp:  time.Now(),
		AttrChange: make(map[string][2]int),
	}

	// Generate target number
	g.targetNum = rand.Intn(100) + 1
	g.attempts = 0
	g.maxAttempts = maxGuessAttempts

	// Game loop
	for g.attempts < g.maxAttempts {
		g.state = StateRunning
		fmt.Println(g.Render())
		
		// Wait for input
		input := <-g.inputChan
		
		// Check if input is a number
		guess, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Printf("请输入1-100之间的数字！按 Enter 继续...\n")
			continue
		}
		
		g.attempts++
		if guess == g.targetNum {
			g.finishGame(true, g.attempts, pet)
			return g.result, false
		} else if guess < g.targetNum {
			fmt.Printf("太小了！再猜一次！按 Enter 继续...\n")
		} else {
			fmt.Printf("太大了！再猜一次！按 Enter 继续...\n")
		}
	}
	
	// Ran out of attempts
	g.finishGame(false, g.attempts, pet)
	return g.result, false
}

// finishGame ends the game and calculates results.
func (g *guessNumberGame) finishGame(won bool, attempts int, pet *game.Pet) {
	g.state = StateFinished
	g.result.Won = won
	g.result.Score = attempts
	
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
func (g *guessNumberGame) Render() string {
	switch g.state {
	case StateWaiting:
		return `🎲 猜数字游戏 🎲

我想了一个1-100之间的数字
你有最多7次机会来猜中它！

准备好了吗？按 Enter 开始！`
		
	case StateRunning:
		return fmt.Sprintf(`🎲 猜数字游戏 🎲
尝试次数: %d/%d

我猜的数字是1-100之间的一个数字。
请输入你的猜测:`,
			g.attempts, g.maxAttempts)
		
	case StateFinished:
		if g.result.Won {
			return fmt.Sprintf(`🎲 游戏结束 🎲
✅ 恭喜猜中了！
数字: %d
尝试次数: %d (%s)
属性变化:
  快乐度: %d → %d
  精力: %d → %d`,
				g.targetNum, g.result.Score, getGuessRating(g.result.Score),
				g.result.AttrChange["happiness"][0],
				g.result.AttrChange["happiness"][1],
				g.result.AttrChange["energy"][0],
				g.result.AttrChange["energy"][1])
		} else {
			return fmt.Sprintf(`🎲 游戏结束 🎲
❌ 用光了所有尝试次数！
正确的数字是: %d
属性变化:
  快乐度: %d → %d
  精力: %d → %d`,
				g.targetNum,
				g.result.AttrChange["happiness"][0],
				g.result.AttrChange["happiness"][1],
				g.result.AttrChange["energy"][0],
				g.result.AttrChange["energy"][1])
		}
		
	default:
		return "游戏状态未知"
	}
}

// getGuessRating returns a rating based on number of attempts.
func getGuessRating(attempts int) string {
	if attempts == 1 {
		return "天才！🧠"
	} else if attempts <= 3 {
		return "很棒！👏"
	} else if attempts <= 5 {
		return "不错！👍"
	} else {
		return "刚好过关！😊"
	}
}