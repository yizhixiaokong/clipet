package games

import (
	"fmt"
	"strconv"
	"strings"
)

// guessEntry 记录一次猜测和提示。
type guessEntry struct {
	guess int
	hint  string
}

// guessNumberGame 实现猜数字游戏（纯状态机）。
type guessNumberGame struct {
	state       GameState
	targetNum   int
	attempts    int
	maxAttempts int
	inputBuf    string       // 玩家正在输入的数字
	history     []guessEntry // 猜测历史
	won         bool
	confirmed   bool
}

func newGuessNumberGame() MiniGame {
	return &guessNumberGame{
		maxAttempts: 7,
	}
}

func (g *guessNumberGame) GetConfig() GameConfig {
	return GameConfig{
		Type:          GameGuessNumber,
		Name:          "猜数字",
		Description:   "猜一个 1-100 的数字，最多 7 次！",
		MinEnergy:     3,
		EnergyCost:    5,
		WinHappiness:  20,
		LoseHappiness: -8,
	}
}

func (g *guessNumberGame) Start() {
	g.state = StateRunning
	g.targetNum = randIntn(100) + 1
	g.attempts = 0
	g.maxAttempts = 7
	g.inputBuf = ""
	g.history = nil
	g.won = false
	g.confirmed = false
}

func (g *guessNumberGame) HandleKey(key string) {
	if g.state == StateFinished {
		if key == "enter" || key == " " {
			g.confirmed = true
		}
		return
	}
	if g.state != StateRunning {
		return
	}

	switch {
	case key >= "0" && key <= "9":
		if len(g.inputBuf) < 3 {
			g.inputBuf += key
		}
	case key == "backspace":
		if len(g.inputBuf) > 0 {
			g.inputBuf = g.inputBuf[:len(g.inputBuf)-1]
		}
	case key == "enter":
		g.submitGuess()
	}
}

func (g *guessNumberGame) submitGuess() {
	if g.inputBuf == "" {
		return
	}
	guess, err := strconv.Atoi(strings.TrimSpace(g.inputBuf))
	g.inputBuf = ""
	if err != nil || guess < 1 || guess > 100 {
		return
	}

	g.attempts++
	if guess == g.targetNum {
		g.won = true
		g.history = append(g.history, guessEntry{guess, "✅ 猜中了！"})
		g.state = StateFinished
	} else if guess < g.targetNum {
		g.history = append(g.history, guessEntry{guess, "太小了 ↑"})
	} else {
		g.history = append(g.history, guessEntry{guess, "太大了 ↓"})
	}
	if g.attempts >= g.maxAttempts && !g.won {
		g.state = StateFinished
	}
}

func (g *guessNumberGame) Tick() {
	// 猜数字不需要时钟驱动逻辑
}

func (g *guessNumberGame) View() string {
	var b strings.Builder
	b.WriteString("🎲 猜数字 (1-100)\n\n")

	for _, e := range g.history {
		b.WriteString(fmt.Sprintf("  %3d  %s\n", e.guess, e.hint))
	}

	if g.state == StateRunning {
		remaining := g.maxAttempts - g.attempts
		b.WriteString(fmt.Sprintf("\n  剩余机会: %d/%d\n", remaining, g.maxAttempts))
		b.WriteString(fmt.Sprintf("  输入数字: %s▌\n", g.inputBuf))
		b.WriteString("\n  输入数字后按 Enter 确认")
	} else {
		b.WriteString("\n")
		if g.won {
			b.WriteString(fmt.Sprintf("  ✅ %d 次猜中！(%s)\n", g.attempts, guessRating(g.attempts)))
		} else {
			b.WriteString(fmt.Sprintf("  ❌ 答案是 %d\n", g.targetNum))
		}
		b.WriteString("\n  按 Enter 继续")
	}

	return b.String()
}

func (g *guessNumberGame) IsFinished() bool  { return g.state == StateFinished }
func (g *guessNumberGame) IsConfirmed() bool { return g.confirmed }

func (g *guessNumberGame) GetResult() *GameResult {
	msg := ""
	if g.won {
		msg = fmt.Sprintf("%d 次猜中 (%s)", g.attempts, guessRating(g.attempts))
	} else {
		msg = fmt.Sprintf("答案是 %d", g.targetNum)
	}
	return &GameResult{
		GameType: GameGuessNumber,
		Won:      g.won,
		Score:    g.attempts,
		Message:  msg,
	}
}

func guessRating(attempts int) string {
	switch {
	case attempts == 1:
		return "天才！🧠"
	case attempts <= 3:
		return "很棒！👏"
	case attempts <= 5:
		return "不错！👍"
	default:
		return "刚好过关 😊"
	}
}
