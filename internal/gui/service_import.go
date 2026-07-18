package gui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"github.com/loveloki/try/internal/i18n"
)

// dropResult 汇总拖拽复制结果。
type dropResult struct {
	Copied  int
	Skipped int
}

// DropProgressFunc 报告拖拽复制进度（done/total 为顶层 URI 序号）。
type DropProgressFunc func(done, total int, current string)

// copyDroppedFiles 将外部拖入的 URI 复制到 destDir（不覆盖已存在目标）。
func (s *service) copyDroppedFiles(destDir string, uris []fyne.URI, onProgress DropProgressFunc) (dropResult, error) {
	var result dropResult
	destDir = filepath.Clean(destDir)
	if err := s.requireAllowed(destDir); err != nil {
		return result, err
	}
	total := countDropURIs(uris)
	done := 0
	for _, uri := range uris {
		if uri == nil {
			continue
		}
		src := uriToLocalPath(uri)
		if src == "" {
			result.Skipped++
			done++
			reportDropProgress(onProgress, done, total, "")
			continue
		}
		reportDropProgress(onProgress, done, total, filepath.Base(src))
		n, err := s.copyPathInto(destDir, src)
		if err != nil {
			return result, err
		}
		result.Copied += n.Copied
		result.Skipped += n.Skipped
		done++
		reportDropProgress(onProgress, done, total, filepath.Base(src))
	}
	return result, nil
}

func countDropURIs(uris []fyne.URI) int {
	n := 0
	for _, uri := range uris {
		if uri != nil {
			n++
		}
	}
	return n
}

func reportDropProgress(fn DropProgressFunc, done, total int, current string) {
	if fn != nil {
		fn(done, total, current)
	}
}

func (s *service) copyPathInto(destDir, src string) (dropResult, error) {
	var result dropResult
	info, err := os.Lstat(src)
	if err != nil {
		return result, fmt.Errorf("%s: %w", i18n.Get().GUIErrCopy, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		result.Skipped = 1
		return result, nil
	}
	dest := filepath.Join(destDir, filepath.Base(src))
	if err := s.requireMutable(dest); err != nil {
		return result, err
	}
	if _, err := os.Lstat(dest); err == nil {
		result.Skipped = 1
		return result, nil
	} else if !os.IsNotExist(err) {
		return result, fmt.Errorf("%s: %w", i18n.Get().GUIErrCopy, err)
	}
	if info.IsDir() {
		if err := copyDir(src, dest); err != nil {
			_ = os.RemoveAll(dest)
			return result, err
		}
		result.Copied = 1
		return result, nil
	}
	if err := copyFile(src, dest); err != nil {
		return result, err
	}
	result.Copied = 1
	return result, nil
}

func uriToLocalPath(uri fyne.URI) string {
	if uri.Scheme() != "file" {
		return ""
	}
	path := uri.Path()
	if path == "" {
		return ""
	}
	return filepath.Clean(path)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.Get().GUIErrCopy, err)
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.Get().GUIErrCopy, err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		_ = os.Remove(dst)
		return fmt.Errorf("%s: %w", i18n.Get().GUIErrCopy, err)
	}
	return out.Close()
}

func copyDir(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.Get().GUIErrCopy, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	if err := os.Mkdir(dst, 0o755); err != nil {
		return fmt.Errorf("%s: %w", i18n.Get().GUIErrCopy, err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.Get().GUIErrCopy, err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		from := filepath.Join(src, e.Name())
		to := filepath.Join(dst, e.Name())
		fi, err := os.Lstat(from)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.Get().GUIErrCopy, err)
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			continue
		}
		if fi.IsDir() {
			if err := copyDir(from, to); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(from, to); err != nil {
			return err
		}
	}
	return nil
}
