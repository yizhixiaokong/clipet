package main

import (
	"clipet/internal/plugin"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

func newPreviewCmd() *cobra.Command {
	var fps int

	cmd := &cobra.Command{
		Use:   "preview <pack-dir> <stage-id> <anim-state>",
		Short: "[开发] 预览 ASCII 动画帧",
		Long: `在终端中循环播放指定物种包的 ASCII 动画帧。

anim-state 可选: idle, eating, sleeping, playing, sad, happy
按 Ctrl+C 退出。`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			stageID := args[1]
			animState := args[2]

			// Check directory
			info, err := os.Stat(dir)
			if err != nil {
				return fmt.Errorf("无法访问 %q: %w", dir, err)
			}
			if !info.IsDir() {
				return fmt.Errorf("%q 不是目录", dir)
			}

			// Parse pack
			pack, err := plugin.ParsePack(os.DirFS(dir), ".")
			if err != nil {
				return fmt.Errorf("解析物种包失败: %w", err)
			}

			// Find frames
			key := plugin.FrameKey(stageID, animState)
			frame, ok := pack.Frames[key]
			if !ok {
				// Try fallback to idle
				key = plugin.FrameKey(stageID, "idle")
				frame, ok = pack.Frames[key]
				if !ok {
					fmt.Fprintf(os.Stderr, "未找到帧: %s_%s 或 %s_idle\n\n", stageID, animState, stageID)
					fmt.Fprintln(os.Stderr, "可用帧集:")
					for k, f := range pack.Frames {
						fmt.Fprintf(os.Stderr, "  %s (%d 帧)\n", k, len(f.Frames))
					}
					return fmt.Errorf("no frames found for %s_%s", stageID, animState)
				}
				fmt.Printf("(fallback to idle, %s_%s 不存在)\n", stageID, animState)
			}

			if len(frame.Frames) == 0 {
				return fmt.Errorf("帧集 %s 为空", key)
			}

			fmt.Printf("▶ 播放: %s (%d 帧, %d fps) — Ctrl+C 退出\n\n", key, len(frame.Frames), fps)

			// Setup signal handler
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			interval := time.Second / time.Duration(fps)
			idx := 0

			for {
				select {
				case <-sigCh:
					fmt.Print("\033[?25h") // show cursor
					fmt.Println("\n\n⏹ 预览结束")
					return nil
				default:
					// ANSI: move cursor to home, clear screen
					fmt.Print("\033[H\033[2J")
					fmt.Printf("▶ %s  帧 %d/%d  (%d fps)\n\n", key, idx+1, len(frame.Frames), fps)
					fmt.Print(frame.Frames[idx])
					fmt.Println()

					idx = (idx + 1) % len(frame.Frames)
					time.Sleep(interval)
				}
			}
		},
	}

	cmd.Flags().IntVar(&fps, "fps", 2, "帧率 (每秒帧数)")

	return cmd
}
