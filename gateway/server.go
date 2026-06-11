package gateway

import (
	"crypto/subtle"
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
	db       *storage.DB
	vmPool   *luavm.VMPool
	cfg      *config.Config
	webDir   string
	apiKey   string
}

func NewServer(db *storage.DB, cfg *config.Config, webDir string, apiKey string) *Server {
	return &Server{
		db:     db,
		vmPool: luavm.NewVMPool(db, cfg),
		cfg:    cfg,
		webDir: webDir,
		apiKey: apiKey,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/lua/execute", s.requireAuth(s.handleLuaExecute))
	mux.HandleFunc("/api/filesystem/", s.requireAuth(s.handleFilesystem))
	mux.HandleFunc("/api/settings", s.requireAuth(s.handleSettings))
	mux.HandleFunc("/api/settings/update", s.requireAuth(s.handleSettingsUpdate))
	mux.HandleFunc("/api/health", s.handleHealth)

	if s.webDir != "" {
		mux.Handle("/", http.FileServer(http.Dir(s.webDir)))
	}

	return corsMiddleware(logMiddleware(mux))
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.apiKey == "" {
			next(w, r)
			return
		}

		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimPrefix(auth, "Bearer ")
			if subtle.ConstantTimeCompare([]byte(token), []byte(s.apiKey)) == 1 {
				next(w, r)
				return
			}
		}

		if r.URL.Query().Get("api_key") == s.apiKey {
			next(w, r)
			return
		}

		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, `{"status":"ok","version":"0.1.0-mvp"}`)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	safe := map[string]interface{}{
		"server": map[string]interface{}{
			"timeout": s.cfg.Server.Timeout,
		},
		"security": map[string]interface{}{
			"max_timeout":     s.cfg.Security.MaxTimeout,
			"max_memory":      s.cfg.Security.MaxMemory,
			"max_matrix_size": s.cfg.Security.MaxMatrixSize,
			"max_tensor_ndim": s.cfg.Security.MaxTensorNdim,
		},
		"engine": map[string]interface{}{
			"edition": s.cfg.Engine.Edition,
		},
		"ui": map[string]interface{}{
			"theme":     s.cfg.UI.Theme,
			"font_size": s.cfg.UI.FontSize,
			"tab_size":  s.cfg.UI.TabSize,
		},
	}

	data, _ := json.Marshal(safe)
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
		Security *config.SecurityConfig `json:"security,omitempty"`
		UI       *config.UIConfig       `json:"ui,omitempty"`
	}

	if err := json.Unmarshal(body, &update); err != nil {
		writeJSON(w, fmt.Sprintf(`{"error":%q}`, err.Error()))
		return
	}

	if update.Security != nil {
		s.cfg.Security = *update.Security
	}
	if update.UI != nil {
		s.cfg.UI = *update.UI
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
		s.vmPool.Discard(L)
		writeJSON(w, `{"error":"execution timeout"}`)
		return
	case err := <-done:
		if err != nil {
			s.vmPool.Discard(L)
			writeJSON(w, fmt.Sprintf(`{"error":%q}`, err.Error()))
			return
		}
		s.vmPool.Put(L)
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
			resp := map[string]interface{}{
				"key":  node.Key,
				"name": node.Name,
				"type": node.ObjType,
			}
			if node.Value != nil {
				resp["value"] = *node.Value
			}
			data, _ := json.Marshal(resp)
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		} else {
			children, err := s.db.ListChildren(path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			type childInfo struct {
				Name string `json:"name"`
				Type string `json:"type"`
				Key  string `json:"key"`
			}
			items := make([]childInfo, len(children))
			for i, c := range children {
				items[i] = childInfo{Name: c.Name, Type: c.ObjType, Key: c.Key}
			}
			resp := map[string]interface{}{
				"path":     path,
				"children": items,
			}
			data, _ := json.Marshal(resp)
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		}

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
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

func FindWebDir() string {
	candidates := []string{"./web", "../web"}
	for _, dir := range candidates {
		if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
			return dir
		}
	}
	return ""
}
