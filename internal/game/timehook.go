package game

import "time"

// TimeHook 定义响应时间演进的回调接口
type TimeHook interface {
	// OnTimeAdvance 在时间流逝时调用
	OnTimeAdvance(elapsed time.Duration, pet *Pet)

	// Name 返回模块名称（用于日志和调试）
	Name() string
}

// TimeHookPriority 定义回调优先级（数字越大越先执行）
type TimeHookPriority int

const (
	PriorityLow      TimeHookPriority = 10  // 统计、日志等
	PriorityNormal   TimeHookPriority = 50  // 常规逻辑
	PriorityHigh     TimeHookPriority = 80  // 核心逻辑（属性衰减）
	PriorityCritical TimeHookPriority = 100 // 死亡判定等
)
