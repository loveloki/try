package gui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
)

// openURL 用系统默认方式打开 URL 或本地路径。
func openURL(ctx context.Context, target string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", target)
	case "windows":
		cmd = exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", target)
	default:
		cmd = exec.CommandContext(ctx, "xdg-open", target)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open %q: %w", target, err)
	}
	_ = cmd.Process.Release()
	return nil
}
