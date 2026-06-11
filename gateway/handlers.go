package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func writeJSON(w http.ResponseWriter, data string) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))
}

func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(fmt.Sprintf(`{"error":%q}`, msg)))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, `{"status":"ok","version":"0.1.0-mvp"}`)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	safe := map[string]interface{}{
		"security": map[string]interface{}{
			"max_timeout":     s.Cfg.Security.MaxTimeout,
			"max_memory":      s.Cfg.Security.MaxMemory,
			"max_matrix_size": s.Cfg.Security.MaxMatrixSize,
			"max_tensor_ndim": s.Cfg.Security.MaxTensorNdim,
		},
		"engine": map[string]interface{}{"edition": s.Cfg.Engine.Edition},
		"ui": map[string]interface{}{
			"theme":     s.Cfg.UI.Theme,
			"font_size": s.Cfg.UI.FontSize,
			"tab_size":  s.Cfg.UI.TabSize,
		},
	}

	data, _ := json.Marshal(safe)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) handleSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, "read body", http.StatusBadRequest)
		return
	}

	var update struct {
		Security *struct {
			MaxTimeout     int `json:"max_timeout"`
			MaxMemory      int `json:"max_memory"`
			MaxMatrixSize  int `json:"max_matrix_size"`
			MaxTensorNdim  int `json:"max_tensor_ndim"`
		} `json:"security,omitempty"`
		UI *struct {
			Theme    string `json:"theme"`
			FontSize int    `json:"font_size"`
			TabSize  int    `json:"tab_size"`
		} `json:"ui,omitempty"`
	}

	if err := json.Unmarshal(body, &update); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if update.Security != nil {
		s.Cfg.Security.MaxTimeout = update.Security.MaxTimeout
		s.Cfg.Security.MaxMemory = update.Security.MaxMemory
		s.Cfg.Security.MaxMatrixSize = update.Security.MaxMatrixSize
		s.Cfg.Security.MaxTensorNdim = update.Security.MaxTensorNdim
	}
	if update.UI != nil {
		s.Cfg.UI.Theme = update.UI.Theme
		s.Cfg.UI.FontSize = update.UI.FontSize
		s.Cfg.UI.TabSize = update.UI.TabSize
	}

	writeJSON(w, `{"status":"ok"}`)
}

func (s *Server) handleLuaExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, "read body", http.StatusBadRequest)
		return
	}

	code := strings.TrimSpace(string(body))
	if code == "" {
		writeError(w, "empty code", http.StatusBadRequest)
		return
	}

	L := s.VMPool.Get()

	// Capture io.print output
	var output []string
	if ioTbl, ok := L.GetGlobal("io").(*lua.LTable); ok {
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
	go func() { done <- L.DoString(code) }()

	select {
	case <-time.After(s.Cfg.GetTimeout()):
		s.VMPool.Discard(L)
		writeJSON(w, `{"error":"execution timeout"}`)
		return
	case err := <-done:
		if err != nil {
			s.VMPool.Discard(L)
			writeJSON(w, fmt.Sprintf(`{"error":%q}`, err.Error()))
			return
		}
		s.VMPool.Put(L)
	}

	result := strings.Join(output, "\n")
	writeJSON(w, fmt.Sprintf(`{"output":%q}`, result))
}

func (s *Server) handleFilesystem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/filesystem")
	if path == "" {
		path = "/"
	}

	node, err := s.DB.GetNode(path)
	if err != nil {
		writeError(w, "not found", http.StatusNotFound)
		return
	}

	if node.ObjType == "dir" {
		children, err := s.DB.ListChildren(path)
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
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
		data, _ := json.Marshal(map[string]interface{}{"path": path, "children": items})
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	} else {
		resp := map[string]interface{}{"key": node.Key, "name": node.Name, "type": node.ObjType}
		if node.Value != nil {
			resp["value"] = *node.Value
		}
		data, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}
