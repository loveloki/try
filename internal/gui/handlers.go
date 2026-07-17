package gui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/loveloki/try/internal/git"
	"github.com/loveloki/try/internal/i18n"
	"github.com/loveloki/try/internal/script"
	"github.com/loveloki/try/internal/selector"
)

func (s *server) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, BootstrapDTO{
		Locale:   s.locale,
		Theme:    s.theme,
		Messages: bootstrapMessages(i18n.Get()),
		Paths: PathsDTO{
			Tries: s.triesPath,
			Ships: s.shipPaths,
		},
	})
}

func (s *server) handleEntries(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	source := r.URL.Query().Get("source")
	if source == "all" {
		source = ""
	}
	entries := selector.LoadAllEntries(s.triesPath, s.shipPaths)
	opts := selector.SourceOptions(s.shipPaths)
	matched := selector.MatchEntries(entries, q, source, 0)
	dtos := make([]EntryDTO, len(matched))
	for i, m := range matched {
		dtos[i] = entryToDTO(m)
	}
	writeJSON(w, http.StatusOK, EntriesResponse{
		Entries: dtos,
		Counts:  selector.SourceCounts(entries, opts),
	})
}

func (s *server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, i18n.Get().ErrBadRequest)
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" || strings.Contains(name, "/") {
		writeErr(w, http.StatusBadRequest, i18n.Get().ErrBadRequest)
		return
	}
	name = strings.ReplaceAll(name, " ", "-")
	dateSuffix := time.Now().Format("2006-01-02")
	name = git.ResolveUniqueName(s.triesPath, name, dateSuffix)
	dirName := name + "-" + dateSuffix
	path := filepath.Join(s.triesPath, dirName)
	if err := s.requireMutable(path); err != nil {
		writeErr(w, http.StatusForbidden, err.Error())
		return
	}
	if err := script.ExecuteSideEffect(&selector.SelectionResult{
		Type: selector.SelectMkdir,
		Path: path,
	}); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"path": path})
}

func (s *server) handleDeleteEntries(w http.ResponseWriter, r *http.Request) {
	var req pathsReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, i18n.Get().ErrBadRequest)
		return
	}
	items, err := s.deleteItems(req.Paths)
	if err != nil {
		writeErr(w, http.StatusForbidden, err.Error())
		return
	}
	if err := script.ExecuteSideEffect(&selector.SelectionResult{
		Type:     selector.SelectDelete,
		Paths:    items,
		BasePath: s.triesPath,
	}); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) handleRename(w http.ResponseWriter, r *http.Request) {
	var req renameReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, i18n.Get().ErrBadRequest)
		return
	}
	if err := s.requireMutable(req.Path); err != nil {
		writeErr(w, http.StatusForbidden, err.Error())
		return
	}
	newName := strings.TrimSpace(req.NewName)
	if newName == "" || strings.Contains(newName, "/") {
		writeErr(w, http.StatusBadRequest, i18n.Get().RenameSlash)
		return
	}
	basePath := filepath.Dir(req.Path)
	oldName := filepath.Base(req.Path)
	if err := script.ExecuteSideEffect(&selector.SelectionResult{
		Type:     selector.SelectRename,
		Old:      oldName,
		New:      newName,
		BasePath: basePath,
	}); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"path": filepath.Join(basePath, newName),
	})
}

func (s *server) handleShip(w http.ResponseWriter, r *http.Request) {
	var req shipReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, i18n.Get().ErrBadRequest)
		return
	}
	if err := s.requireMutable(req.Path); err != nil {
		writeErr(w, http.StatusForbidden, err.Error())
		return
	}
	if req.DestIndex < 0 || req.DestIndex >= len(s.shipPaths) {
		writeErr(w, http.StatusBadRequest, i18n.Get().ErrBadRequest)
		return
	}
	basename := filepath.Base(req.Path)
	dest := filepath.Join(s.shipPaths[req.DestIndex], basename)
	if err := s.requireMutable(dest); err != nil {
		writeErr(w, http.StatusForbidden, err.Error())
		return
	}
	if err := script.ExecuteSideEffect(&selector.SelectionResult{
		Type:     selector.SelectShip,
		Source:   req.Path,
		Dest:     dest,
		Basename: basename,
		BasePath: s.triesPath,
	}); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"path": dest})
}

func (s *server) handleFiles(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if err := s.requireAllowed(path); err != nil {
		writeErr(w, http.StatusForbidden, err.Error())
		return
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, i18n.Get().ErrReadDir+err.Error())
		return
	}
	files := make([]FileDTO, 0, len(entries))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		full := filepath.Join(path, e.Name())
		sizeKB := float64(info.Size()) / 1024
		files = append(files, FileDTO{
			ID:       full,
			Name:     e.Name(),
			Type:     fileTypeOf(e.Name(), e.IsDir()),
			SizeKB:   sizeKB,
			Modified: info.ModTime().UTC().Format(time.RFC3339),
			IsDir:    e.IsDir(),
			Path:     full,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"files": files})
}

func (s *server) handleDeleteFiles(w http.ResponseWriter, r *http.Request) {
	var req pathsReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, i18n.Get().ErrBadRequest)
		return
	}
	for _, p := range req.Paths {
		if err := s.requireMutable(p); err != nil {
			writeErr(w, http.StatusForbidden, err.Error())
			return
		}
		if err := os.RemoveAll(p); err != nil {
			writeErr(w, http.StatusInternalServerError, fmt.Sprintf("%s: %v", i18n.Get().ErrDeletePartial, err))
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) handleOpenFile(w http.ResponseWriter, r *http.Request) {
	var req pathReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, i18n.Get().ErrBadRequest)
		return
	}
	if err := s.requireAllowed(req.Path); err != nil {
		writeErr(w, http.StatusForbidden, err.Error())
		return
	}
	ctx := r.Context()
	if err := openURL(ctx, req.Path); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) deleteItems(paths []string) ([]selector.DeleteItem, error) {
	items := make([]selector.DeleteItem, 0, len(paths))
	for _, p := range paths {
		if err := s.requireMutable(p); err != nil {
			return nil, err
		}
		items = append(items, selector.DeleteItem{Path: p, Basename: filepath.Base(p)})
	}
	return items, nil
}

func (s *server) requireAllowed(path string) error {
	if !IsAllowedPath(path, s.roots) {
		return fmt.Errorf("%s", i18n.Get().ErrPathDenied)
	}
	return nil
}

// requireMutable 校验路径可被修改：必须严格位于根目录内，拒绝根目录自身。
func (s *server) requireMutable(path string) error {
	if !IsAllowedTarget(path, s.roots) {
		return fmt.Errorf("%s", i18n.Get().ErrPathDenied)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResp{Error: msg})
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	return dec.Decode(dst)
}
