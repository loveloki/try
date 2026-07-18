package gui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

// openTarget 用系统默认方式打开本地路径。
func openTarget(ctx context.Context, target string) error {
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

// revealTarget 在系统文件管理器中显示目录。
func revealTarget(ctx context.Context, dir string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", dir)
	case "windows":
		cmd = exec.CommandContext(ctx, "explorer", dir)
	default:
		cmd = exec.CommandContext(ctx, "xdg-open", dir)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("reveal %q: %w", dir, err)
	}
	_ = cmd.Process.Release()
	return nil
}

func revealCommand(goos, dir string) (name string, args []string) {
	switch goos {
	case "darwin":
		return "open", []string{dir}
	case "windows":
		return "explorer", []string{dir}
	default:
		return "xdg-open", []string{dir}
	}
}

func (s *service) revealInFileManager(path string) error {
	if err := s.requireAllowed(path); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return revealTarget(ctx, path)
}
