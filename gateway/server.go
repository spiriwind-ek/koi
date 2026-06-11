package gateway

import (
	"log"
	"net/http"
	"time"

	"koi/config"
	"koi/lua"
	"koi/storage"
)

type Server struct {
	DB     *storage.DB
	VMPool *lua.VMPool
	Cfg    *config.Config
	WebDir string
	APIKey string
}

func NewServer(db *storage.DB, cfg *config.Config, webDir string, apiKey string) *Server {
	return &Server{
		DB:     db,
		VMPool: lua.NewVMPool(db, cfg),
		Cfg:    cfg,
		WebDir: webDir,
		APIKey: apiKey,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/lua/execute", s.requireAuth(s.handleLuaExecute))
	mux.HandleFunc("/api/filesystem/", s.requireAuth(s.handleFilesystem))
	mux.HandleFunc("/api/settings", s.requireAuth(s.handleSettings))
	mux.HandleFunc("/api/settings/update", s.requireAuth(s.handleSettingsUpdate))
	mux.HandleFunc("/api/health", s.handleHealth)

	if s.WebDir != "" {
		mux.Handle("/", http.FileServer(http.Dir(s.WebDir)))
	}

	return corsMiddleware(logMiddleware(mux))
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
