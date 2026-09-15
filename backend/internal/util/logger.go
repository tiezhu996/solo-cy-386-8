package util

import (
	"log/slog"
	"os"
)

// NewLogger 创建 slog 结构化日志器（屎山约束：日志模块单独管理但全栈引用）。
func NewLogger(env string) *slog.Logger {
	level := slog.LevelInfo
	if env == "development" {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
