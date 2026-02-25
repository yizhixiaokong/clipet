// Package games 提供迷你游戏的框架和实现。
// 所有游戏均为纯状态机，不包含任何阻塞 I/O，
// 通过 Start/HandleKey/Tick/View 接口与 Bubble Tea TUI 事件循环集成。
package games

// GameType 表示游戏类型。
type GameType string

const (
	GameReactionSpeed GameType = "reaction_speed"
	GameGuessNumber   GameType = "guess_number"
)

// GameState 表示游戏的内部状态。
type GameState int

const (
	StateIdle     GameState = iota // 未开始
	StateWaiting                   // 倒计时 / 准备阶段
	StateRunning                   // 游戏进行中
	StateFinished                  // 游戏结束，展示结果
)

// GameResult 保存游戏的结果（不修改宠物属性，由调用方处理）。
type GameResult struct {
	GameType GameType
	Won      bool
	Score    int    // 游戏特定分数（反应时间 ms / 猜测次数）
	Message  string // 格式化的结果描述
}

// GameConfig 定义游戏的配置参数。
type GameConfig struct {
	Type          GameType
	Name          string
	Description   string
	MinEnergy     int // 所需最低精力
	EnergyCost    int // 玩一次消耗的精力
	WinHappiness  int // 赢了增加的快乐度
	LoseHappiness int // 输了减少的快乐度（通常为负数）
}

// MiniGame 定义所有迷你游戏的接口（纯状态机，无阻塞 I/O）。
type MiniGame interface {
	// GetConfig 返回游戏配置。
	GetConfig() GameConfig

	// Start 初始化/重置游戏状态，准备开始。
	Start()

	// HandleKey 处理一次按键输入。
	HandleKey(key string)

	// Tick 处理一次时钟脉冲（约 500ms），用于倒计时等。
	Tick()

	// View 返回当前游戏画面（纯字符串，由 TUI 层包裹样式）。
	View() string

	// IsFinished 返回游戏是否已结束。
	IsFinished() bool

	// IsConfirmed 返回玩家是否在结果画面按了确认键。
	IsConfirmed() bool

	// GetResult 返回游戏结果（仅在 IsFinished 后有效）。
	GetResult() *GameResult
}
