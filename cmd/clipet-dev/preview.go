package main

import (
	"clipet/internal/plugin"
	"clipet/internal/tui/dev"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	tea "charm.land/bubbletea/v2"
)

func newPreviewCmd() *cobra.Command {
	var fps int

	cmd := &cobra.Command{
		Use:   "preview <pack-dir> [stage-id] [anim-state]",
		Short: "[开发] 预览 ASCII 动画帧",
		Long: `交互式 TUI 预览物种包的 ASCII 动画帧。

左侧动画预览，右侧树形列表选择。
可选参数用于定位初始位置：
  preview <dir>                    — 从第一个动画开始
  preview <dir> <stage-id>         — 定位到该阶段的 idle
  preview <dir> <stage-id> <anim>  — 定位到指定动画

操作: ↑↓/jk 选择  +/- 调速  q/Esc 退出`,
		Args: cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]

			info, err := os.Stat(dir)
			if err != nil {
				return fmt.Errorf("无法访问 %q: %w", dir, err)
			}
			if !info.IsDir() {
				return fmt.Errorf("%q 不是目录", dir)
			}

			pack, err := plugin.ParsePack(os.DirFS(dir), ".")
			if err != nil {
				return fmt.Errorf("解析物种包失败: %w", err)
			}

			var initStage, initAnim string
			if len(args) >= 2 {
				initStage = args[1]
			}
			if len(args) >= 3 {
				initAnim = args[2]
			}

			return runPreviewTUI(pack, fps, initStage, initAnim)
		},
	}

	cmd.Flags().IntVar(&fps, "fps", 2, "帧率 (每秒帧数)")

	return cmd
}

func runPreviewTUI(pack *plugin.SpeciesPack, fps int, initStage, initAnim string) error {
	m := dev.NewPreviewModel(pack, fps, initStage, initAnim, i18nMgr)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
