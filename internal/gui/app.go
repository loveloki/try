package gui

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/loveloki/try/internal/config"
	"github.com/loveloki/try/internal/i18n"
)

// Options 启动 GUI 后端的选项。
type Options struct {
	Path string // 可选，覆盖 tries 根目录（对应 -path）
}

// Run 加载配置、启动本机 HTTP 服务、打开浏览器，并阻塞至收到退出信号。
func Run(opts Options) error {
	cfg, err := loadOrInitConfig()
	if err != nil {
		return err
	}
	triesPath, shipPaths := config.ResolvePaths(opts.Path, cfg)
	locale := config.ResolveLocale("", cfg)
	i18n.Init(locale)

	ensureGUIDirs(triesPath, shipPaths)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	addr := ln.Addr().String()
	url := "http://" + addr + "/"

	srv := newServer(serverConfig{
		triesPath: triesPath,
		shipPaths: shipPaths,
		locale:    locale,
		theme:     config.DetectTheme(),
	})

	httpSrv := &http.Server{
		Handler:           srv.handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpSrv.Serve(ln)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := openURL(ctx, url); err != nil {
		fmt.Fprintf(os.Stderr, "open browser: %v\n", err)
		fmt.Fprintf(os.Stderr, "open %s manually\n", url)
	} else {
		fmt.Fprintf(os.Stderr, "try-gui listening on %s\n", url)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		fmt.Fprintf(os.Stderr, "received %s, shutting down\n", sig)
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server: %w", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = httpSrv.Shutdown(shutdownCtx)
	return nil
}

func loadOrInitConfig() (config.Config, error) {
	cfg, err := config.LoadConfig()
	if err == nil {
		return cfg, nil
	}
	if _, initErr := config.InitConfigFile(); initErr != nil {
		return config.Config{}, fmt.Errorf("init config: %w", initErr)
	}
	cfg, err = config.LoadConfig()
	if err != nil {
		return config.Config{}, fmt.Errorf("load config: %w", err)
	}
	return cfg, nil
}

func ensureGUIDirs(triesPath string, shipPaths []string) {
	_ = os.MkdirAll(triesPath, 0o755)
	for _, sp := range shipPaths {
		_ = os.MkdirAll(sp, 0o755)
	}
}
