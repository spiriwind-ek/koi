package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	luavm "koi/lua"
	"koi/config"
	"koi/storage"

	lua "github.com/yuin/gopher-lua"
)

type Server struct {
	db      *storage.DB
	vmPool  *luavm.VMPool
	cfg     *config.Config
	webDir  string
}

func NewServer(db *storage.DB, cfg *config.Config, webDir string) *Server {
	return &Server{
		db:     db,
		vmPool: luavm.NewVMPool(db),
		cfg:    cfg,
		webDir: webDir,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/lua/execute", s.handleLuaExecute)
	mux.HandleFunc("/api/filesystem/", s.handleFilesystem)
	mux.HandleFunc("/api/settings", s.handleSettings)
	mux.HandleFunc("/api/settings/update", s.handleSettingsUpdate)
	mux.HandleFunc("/api/health", s.handleHealth)

	if s.webDir != "" {
		mux.Handle("/", http.FileServer(http.Dir(s.webDir)))
	}

	return corsMiddleware(logMiddleware(mux))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, `{"status":"ok","version":"0.1.0-mvp"}`)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := json.Marshal(s.cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) handleSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}

	var update struct {
		Server   *config.ServerConfig   `json:"server,omitempty"`
		Database *config.DatabaseConfig `json:"database,omitempty"`
		Security *config.SecurityConfig `json:"security,omitempty"`
		Engine   *config.EngineConfig   `json:"engine,omitempty"`
		UI       *config.UIConfig       `json:"ui,omitempty"`
	}

	if err := json.Unmarshal(body, &update); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if update.Server != nil {
		s.cfg.Server = *update.Server
	}
	if update.Database != nil {
		s.cfg.Database = *update.Database
	}
	if update.Security != nil {
		s.cfg.Security = *update.Security
	}
	if update.Engine != nil {
		s.cfg.Engine = *update.Engine
	}
	if update.UI != nil {
		s.cfg.UI = *update.UI
	}

	configPath := "config/koi.toml"
	if err := s.cfg.Save(configPath); err != nil {
		writeJSON(w, fmt.Sprintf(`{"error":%q}`, err.Error()))
		return
	}

	writeJSON(w, `{"status":"ok"}`)
}

func (s *Server) handleLuaExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}

	code := strings.TrimSpace(string(body))
	if code == "" {
		http.Error(w, "empty code", http.StatusBadRequest)
		return
	}

	L := s.vmPool.Get()
	defer s.vmPool.Put(L)

	var output []string
	ioTbl, ok := L.GetGlobal("io").(*lua.LTable)
	if ok {
		L.SetField(ioTbl, "print", L.NewFunction(func(L *lua.LState) int {
			top := L.GetTop()
			args := make([]interface{}, top)
			for i := 1; i <= top; i++ {
				args[i-1] = L.Get(i).String()
			}
			output = append(output, fmt.Sprint(args...))
			return 0
		}))
	}

	done := make(chan error, 1)
	go func() {
		done <- L.DoString(code)
	}()

	select {
	case <-time.After(s.cfg.GetTimeout()):
		writeJSON(w, `{"error":"execution timeout"}`)
		return
	case err := <-done:
		if err != nil {
			writeJSON(w, fmt.Sprintf(`{"error":%q}`, err.Error()))
			return
		}
	}

	result := strings.Join(output, "\n")
	writeJSON(w, fmt.Sprintf(`{"output":%q}`, result))
}

func (s *Server) handleFilesystem(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/filesystem")
	if path == "" {
		path = "/"
	}

	switch r.Method {
	case http.MethodGet:
		if s.db.NodeExists(path) {
			node, err := s.db.GetNode(path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, fmt.Sprintf(`{"key":%q,"name":%q,"type":%q,"value":%v}`,
				node.Key, node.Name, node.ObjType, deref(node.Value)))
		} else {
			children, err := s.db.ListChildren(path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			var items []string
			for _, c := range children {
				items = append(items, fmt.Sprintf(`{"name":%q,"type":%q,"key":%q}`,
					c.Name, c.ObjType, c.Key))
			}
			writeJSON(w, fmt.Sprintf(`{"path":%q,"children":[%s]}`, path, strings.Join(items, ",")))
		}

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func writeJSON(w http.ResponseWriter, data string) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))
}

func deref(s *string) string {
	if s == nil {
		return "null"
	}
	return `"` + *s + `"`
}

func FindWebDir() string {
	candidates := []string{"./web", "../web", "./gateway/web"}
	for _, dir := range candidates {
		if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
			return dir
		}
	}
	return ""
}
