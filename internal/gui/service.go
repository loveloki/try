package gui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/loveloki/try/internal/git"
	"github.com/loveloki/try/internal/i18n"
	"github.com/loveloki/try/internal/script"
	"github.com/loveloki/try/internal/selector"
)

// service 封装 GUI 视图可调用的业务操作。
type service struct {
	triesPath string
	shipPaths []string
}

func newService(triesPath string, shipPaths []string) *service {
	return &service{triesPath: triesPath, shipPaths: shipPaths}
}

// touchDir 更新目录 mtime，使 selector 按访问时间重排（对齐 TUI execCd）。
func (s *service) touchDir(path string) {
	path = filepath.Clean(path)
	if err := s.requireAllowed(path); err != nil {
		return
	}
	now := time.Now()
	_ = os.Chtimes(path, now, now)
}

func (s *service) listEntries(query, source string) EntriesResult {
	entries := selector.LoadAllEntries(s.triesPath, s.shipPaths)
	sources := selector.SourceOptions(s.shipPaths)
	matched := selector.MatchEntries(entries, query, normalizeSource(source), 0)
	views := make([]EntryView, len(matched))
	for i, m := range matched {
		views[i] = entryToView(m)
	}
	return EntriesResult{
		Entries: views,
		Counts:  selector.SourceCounts(entries, sources),
		Sources: sources,
	}
}

func (s *service) createEntry(name string) (string, error) {
	name = normalizeName(name)
	if name == "" {
		return "", errors.New(i18n.Get().EmptyInputHint)
	}
	if strings.Contains(name, "/") {
		return "", errors.New(i18n.Get().RenameSlash)
	}
	dateSuffix := time.Now().Format("2006-01-02")
	base := git.ResolveUniqueName(s.triesPath, name, dateSuffix)
	path := filepath.Join(s.triesPath, base+"-"+dateSuffix)
	if err := s.requireMutable(path); err != nil {
		return "", err
	}
	err := script.ExecuteSideEffect(&selector.SelectionResult{
		Type: selector.SelectMkdir,
		Path: path,
	})
	return path, err
}

func (s *service) deleteEntries(paths []string) error {
	items, err := s.deleteItems(paths)
	if err != nil {
		return err
	}
	return script.ExecuteSideEffect(&selector.SelectionResult{
		Type:     selector.SelectDelete,
		Paths:    items,
		BasePath: s.triesPath,
	})
}

func (s *service) renameEntry(path, newName string) (string, error) {
	if err := s.requireMutable(path); err != nil {
		return "", err
	}
	newName = normalizeName(newName)
	if newName == "" {
		return "", errors.New(i18n.Get().RenameEmpty)
	}
	if strings.Contains(newName, "/") {
		return "", errors.New(i18n.Get().RenameSlash)
	}
	basePath := filepath.Dir(path)
	oldName := filepath.Base(path)
	err := script.ExecuteSideEffect(&selector.SelectionResult{
		Type:     selector.SelectRename,
		Old:      oldName,
		New:      newName,
		BasePath: basePath,
	})
	return filepath.Join(basePath, newName), err
}

func (s *service) shipEntry(path string, destIndex int) (string, error) {
	if err := s.requireMutable(path); err != nil {
		return "", err
	}
	if destIndex < 0 || destIndex >= len(s.shipPaths) {
		return "", errors.New(i18n.Get().ErrBadRequest)
	}
	basename := filepath.Base(path)
	dest := filepath.Join(s.shipPaths[destIndex], basename)
	if err := s.requireMutable(dest); err != nil {
		return "", err
	}
	err := script.ExecuteSideEffect(&selector.SelectionResult{
		Type:     selector.SelectShip,
		Source:   path,
		Dest:     dest,
		Basename: basename,
		BasePath: s.triesPath,
	})
	return dest, err
}

// cloneEntry 克隆 Git 仓库到 tries 目录，返回目标路径。
func (s *service) cloneEntry(uri, customName string) (string, error) {
	dirName := git.GenerateCloneDirName(uri, customName)
	if dirName == "" {
		return "", fmt.Errorf("%s%s", i18n.Get().ErrParseGitURI, uri)
	}
	targetPath := filepath.Join(s.triesPath, dirName)
	if err := script.ExecClone(io.Discard, io.Discard, targetPath, uri); err != nil {
		return "", err
	}
	return targetPath, nil
}

func (s *service) listFiles(path string) ([]FileEntry, error) {
	path = filepath.Clean(path)
	if err := s.requireAllowed(path); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("%s%w", i18n.Get().ErrReadDir, err)
	}
	files := make([]FileEntry, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if name == "" || name == "." || name == ".." || strings.HasPrefix(name, ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		full := filepath.Join(path, name)
		files = append(files, FileEntry{
			ID:     full,
			Name:   name,
			Type:   fileTypeOf(name, e.IsDir()),
			SizeKB: float64(info.Size()) / 1024,
			Mtime:  info.ModTime(),
			IsDir:  e.IsDir(),
			Path:   full,
		})
	}
	sortFileEntries(files)
	return files, nil
}

// sortFileEntries 目录优先，同类型按名称不区分大小写升序。
func sortFileEntries(files []FileEntry) {
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})
}

func (s *service) deleteFiles(paths []string) error {
	for _, p := range paths {
		if err := s.requireMutable(p); err != nil {
			return err
		}
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("%s: %w", i18n.Get().ErrDeletePartial, err)
		}
	}
	return nil
}

func (s *service) openFile(ctx context.Context, path string) error {
	if err := s.requireAllowed(path); err != nil {
		return err
	}
	return openTarget(ctx, path)
}

func (s *service) deleteItems(paths []string) ([]selector.DeleteItem, error) {
	items := make([]selector.DeleteItem, 0, len(paths))
	for _, p := range paths {
		if err := s.requireMutable(p); err != nil {
			return nil, err
		}
		items = append(items, selector.DeleteItem{
			Path:     p,
			Basename: filepath.Base(p),
		})
	}
	return items, nil
}

func (s *service) requireAllowed(path string) error {
	if IsAllowedPath(path, s.allowedRoots()) {
		return nil
	}
	return errors.New(i18n.Get().ErrPathDenied)
}

func (s *service) requireMutable(path string) error {
	if IsAllowedTarget(path, s.allowedRoots()) {
		return nil
	}
	return errors.New(i18n.Get().ErrPathDenied)
}

func (s *service) allowedRoots() []string {
	roots := make([]string, 0, 1+len(s.shipPaths))
	roots = append(roots, s.triesPath)
	roots = append(roots, s.shipPaths...)
	return roots
}

func normalizeName(name string) string {
	return strings.ReplaceAll(strings.TrimSpace(name), " ", "-")
}

func normalizeSource(source string) string {
	if source == "all" {
		return ""
	}
	return source
}
