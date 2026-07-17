package gui

import (
	"net/http"
	"net/url"
)

type serverConfig struct {
	triesPath string
	shipPaths []string
	locale    string
	theme     string
}

type server struct {
	triesPath string
	shipPaths []string
	roots     []string
	locale    string
	theme     string
}

func newServer(cfg serverConfig) *server {
	roots := make([]string, 0, 1+len(cfg.shipPaths))
	roots = append(roots, cfg.triesPath)
	roots = append(roots, cfg.shipPaths...)
	return &server{
		triesPath: cfg.triesPath,
		shipPaths: cfg.shipPaths,
		roots:     roots,
		locale:    cfg.locale,
		theme:     cfg.theme,
	}
}

func (s *server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/bootstrap", s.handleBootstrap)
	mux.HandleFunc("GET /api/entries", s.handleEntries)
	mux.HandleFunc("POST /api/entries/create", s.handleCreate)
	mux.HandleFunc("POST /api/entries/delete", s.handleDeleteEntries)
	mux.HandleFunc("POST /api/entries/rename", s.handleRename)
	mux.HandleFunc("POST /api/entries/ship", s.handleShip)
	mux.HandleFunc("GET /api/files", s.handleFiles)
	mux.HandleFunc("POST /api/files/delete", s.handleDeleteFiles)
	mux.HandleFunc("POST /api/files/open", s.handleOpenFile)

	if assets := WebAssets(); assets != nil {
		fileServer := http.FileServer(http.FS(assets))
		mux.Handle("GET /", fileServer)
		mux.Handle("GET /assets/", http.StripPrefix("/assets/", fileServer))
	}

	return corsLocalhost(mux)
}

func corsLocalhost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" || isLocalOrigin(origin) {
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isLocalOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := u.Hostname()
	return host == "127.0.0.1" || host == "localhost" || host == "::1"
}
