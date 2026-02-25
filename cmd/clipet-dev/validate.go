package main

import (
	"clipet/internal/plugin"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <pack-dir>",
		Short: "[开发] 验证物种包目录",
		Long:  "对指定目录的物种包执行完整解析与校验，报告所有问题。",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]

			// Check directory exists
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
				fmt.Fprintf(os.Stderr, "✗ 解析失败: %v\n", err)
				os.Exit(1)
			}

			// Validate
			errs := plugin.Validate(pack)
			if len(errs) == 0 {
				stageCount := len(pack.Stages)
				evoCount := len(pack.Evolutions)
				frameCount := len(pack.Frames)
				dlgCount := len(pack.Dialogues)
				advCount := len(pack.Adventures)

				fmt.Printf("✓ 物种包 %q 校验通过\n", pack.Species.ID)
				fmt.Printf("  名称: %s (v%s)\n", pack.Species.Name, pack.Species.Version)
				fmt.Printf("  阶段: %d, 进化路径: %d\n", stageCount, evoCount)
				fmt.Printf("  帧集: %d, 对话组: %d, 冒险: %d\n", frameCount, dlgCount, advCount)
				return nil
			}

			fmt.Fprintf(os.Stderr, "✗ 物种包 %q 校验失败 (%d 个问题):\n", pack.Species.ID, len(errs))
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "  - %s: %s\n", e.Field, e.Message)
			}
			os.Exit(1)
			return nil
		},
	}
}
