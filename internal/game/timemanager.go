package game

import (
	"sort"
	"time"
)

// TimeManager 管理时间演进和回调分发
type TimeManager struct {
	hooks []hookEntry
}

type hookEntry struct {
	hook     TimeHook
	priority TimeHookPriority
}

// NewTimeManager 创建时间管理器
func NewTimeManager() *TimeManager {
	return &TimeManager{}
}

// RegisterHook 注册时间回调（按优先级排序）
func (tm *TimeManager) RegisterHook(hook TimeHook, priority TimeHookPriority) {
	entry := hookEntry{hook: hook, priority: priority}
	tm.hooks = append(tm.hooks, entry)

	// 按优先级降序排序（高优先级先执行）
	sort.Slice(tm.hooks, func(i, j int) bool {
		return tm.hooks[i].priority > tm.hooks[j].priority
	})
}

// AdvanceTime 统一的时间演进入口
func (tm *TimeManager) AdvanceTime(elapsed time.Duration, pet *Pet) {
	for _, entry := range tm.hooks {
		entry.hook.OnTimeAdvance(elapsed, pet)
	}
}
