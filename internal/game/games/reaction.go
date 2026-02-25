package games

import (
	"fmt"
	"math/rand"
	"time"
)

// reactionSpeedGame 实现反应速度测试游戏（纯状态机）。
type reactionSpeedGame struct {
	state     GameState
	startedAt time.Time     // 游戏开始时间
	readyAt   time.Time     // GO! 出现时间
	delay     time.Duration // 随机等待时长
	score     int           // 反应时间（ms）
	won       bool
	confirmed bool
}

func newReactionSpeedGame() MiniGame {
	return &reactionSpeedGame{}
}

func (g *reactionSpeedGame) GetConfig() GameConfig {
	return GameConfig{
		Type:          GameReactionSpeed,
		Name:          "反应速度测试",
		Description:   "当出现 GO! 时，尽快按键！",
		MinEnergy:     5,
		EnergyCost:    8,
		WinHappiness:  15,
		LoseHappiness: -5,
	}
}

func (g *reactionSpeedGame) Start() {
	g.state = StateWaiting
	g.startedAt = time.Now()
	g.delay = time.Duration(rand.Intn(4000)+2000) * time.Millisecond // 2-6秒
	g.readyAt = time.Time{}
	g.score = 0
	g.won = false
	g.confirmed = false
}

func (g *reactionSpeedGame) HandleKey(key string) {
	switch g.state {
	case StateWaiting:
		// 还没出现 GO! 就按了 → 失败
		g.state = StateFinished
		g.won = false
		g.score = 0

	case StateRunning:
		// GO! 出现后按键 → 计算反应时间
		g.score = int(time.Since(g.readyAt).Milliseconds())
		g.won = g.score < 1000 // 1秒内算赢
		g.state = StateFinished

	case StateFinished:
		if key == "enter" || key == " " {
			g.confirmed = true
		}
	}
}

func (g *reactionSpeedGame) Tick() {
	switch g.state {
	case StateWaiting:
		if time.Since(g.startedAt) >= g.delay {
			g.state = StateRunning
			g.readyAt = time.Now()
		}
	case StateRunning:
		// 3秒超时
		if time.Since(g.readyAt) > 3*time.Second {
			g.state = StateFinished
			g.won = false
			g.score = 3000
		}
	}
}

func (g *reactionSpeedGame) View() string {
	switch g.state {
	case StateWaiting:
		elapsed := time.Since(g.startedAt)
		dots := ""
		n := int(elapsed.Seconds()) % 4
		for i := 0; i < n; i++ {
			dots += "."
		}
		return fmt.Sprintf(
			"⚡ 反应速度测试\n\n"+
				"  准备%s\n\n"+
				"  看到 GO! 时按任意键！\n\n"+
				"  ⚠ 别按太早哦！",
			dots)

	case StateRunning:
		return "⚡ 反应速度测试\n\n" +
			"  ┌──────────────┐\n" +
			"  │   ⚡ GO! ⚡   │\n" +
			"  └──────────────┘\n\n" +
			"  快！按任意键！"

	case StateFinished:
		return g.finishedView()

	default:
		return ""
	}
}

func (g *reactionSpeedGame) finishedView() string {
	if g.won {
		return fmt.Sprintf(
			"⚡ 反应速度测试 — 结果\n\n"+
				"  ✅ 反应时间: %d 毫秒 (%s)\n\n"+
				"  按 Enter 继续",
			g.score, reactionRating(g.score))
	}
	if g.score == 0 {
		return "⚡ 反应速度测试 — 结果\n\n" +
			"  ❌ 太早了！还没出现 GO! 就按了\n\n" +
			"  按 Enter 继续"
	}
	return fmt.Sprintf(
		"⚡ 反应速度测试 — 结果\n\n"+
			"  ❌ 超时了！(%d 毫秒)\n\n"+
			"  按 Enter 继续",
		g.score)
}

func (g *reactionSpeedGame) IsFinished() bool  { return g.state == StateFinished }
func (g *reactionSpeedGame) IsConfirmed() bool { return g.confirmed }

func (g *reactionSpeedGame) GetResult() *GameResult {
	msg := ""
	if g.won {
		msg = fmt.Sprintf("反应时间 %dms (%s)", g.score, reactionRating(g.score))
	} else if g.score == 0 {
		msg = "按太早了！"
	} else {
		msg = "超时了！"
	}
	return &GameResult{
		GameType: GameReactionSpeed,
		Won:      g.won,
		Score:    g.score,
		Message:  msg,
	}
}

func reactionRating(ms int) string {
	switch {
	case ms < 200:
		return "超快！🚀"
	case ms < 300:
		return "很快！⚡"
	case ms < 400:
		return "不错！👍"
	case ms < 500:
		return "一般 😐"
	default:
		return "慢了 🐌"
	}
}
