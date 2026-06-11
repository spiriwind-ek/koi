package lua

import (
	"encoding/json"
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
	"koi/storage"
)

func fsMkdir(db *storage.DB) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		if err := db.AutoMkdir(path); err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LTrue)
		return 1
	}
}

func fsWrite(db *storage.DB) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		val := L.Get(2)

		parent := parentPath(path)
		if !db.NodeExists(parent) {
			if err := db.AutoMkdir(parent); err != nil {
				L.Push(lua.LFalse)
				L.Push(lua.LString(err.Error()))
				return 2
			}
		}

		var valueStr *string
		var blobData []byte
		var blobMeta *string
		objType := "string"

		switch v := val.(type) {
		case *lua.LTable:
			data, _ := json.Marshal(luaTableToMap(L, v))
			s := string(data)
			valueStr = &s
			objType = "table"
		case lua.LNumber:
			s := v.String()
			valueStr = &s
			objType = "number"
		case lua.LString:
			s := string(v)
			if len(s) > 1024 {
				blobData = []byte(s)
				meta := `{"format":"raw","compressed":false}`
				blobMeta = &meta
				objType = "blob"
			} else {
				valueStr = &s
			}
		case lua.LBool:
			s := v.String()
			valueStr = &s
			objType = "bool"
		default:
			s := v.String()
			valueStr = &s
		}

		n := &storage.Node{
			Key:      path,
			ParentKey: &parent,
			Name:     baseName(path),
			ObjType:  objType,
			Value:    valueStr,
			BlobData: blobData,
			BlobMeta: blobMeta,
		}

		if db.NodeExists(path) {
			if err := db.UpdateNode(n); err != nil {
				L.Push(lua.LFalse)
				L.Push(lua.LString(err.Error()))
				return 2
			}
		} else {
			if err := db.CreateNode(n); err != nil {
				L.Push(lua.LFalse)
				L.Push(lua.LString(err.Error()))
				return 2
			}
		}

		L.Push(lua.LTrue)
		return 1
	}
}

func fsRead(db *storage.DB) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		n, err := db.GetNode(path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		if n.Value != nil {
			var m map[string]interface{}
			if json.Unmarshal([]byte(*n.Value), &m) == nil {
				L.Push(mapToLuaTable(L, m))
				return 1
			}
			L.Push(lua.LString(*n.Value))
			return 1
		}

		if n.BlobData != nil {
			L.Push(lua.LString(string(n.BlobData)))
			return 1
		}

		L.Push(lua.LNil)
		return 1
	}
}

func fsLs(db *storage.DB) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		children, err := db.ListChildren(path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbl := L.NewTable()
		for _, c := range children {
			entry := L.NewTable()
			L.SetField(entry, "name", lua.LString(c.Name))
			L.SetField(entry, "type", lua.LString(c.ObjType))
			L.SetField(entry, "key", lua.LString(c.Key))
			L.SetField(entry, "updated", lua.LNumber(c.UpdatedAt))
			tbl.Append(entry)
		}

		L.Push(tbl)
		return 1
	}
}

func fsRm(db *storage.DB) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		if err := db.DeleteNode(path); err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LTrue)
		return 1
	}
}

func fsExists(db *storage.DB) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		L.Push(lua.LBool(db.NodeExists(path)))
		return 1
	}
}

// helpers
func parentPath(path string) string {
	for i := len(path) - 1; i > 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "/"
}

func baseName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

func luaTableToMap(L *lua.LState, tbl *lua.LTable) map[string]interface{} {
	m := make(map[string]interface{})
	tbl.ForEach(func(key, value lua.LValue) {
		switch v := value.(type) {
		case *lua.LTable:
			m[key.String()] = luaTableToMap(L, v)
		case lua.LNumber:
			m[key.String()] = float64(v)
		case lua.LBool:
			m[key.String()] = bool(v)
		default:
			m[key.String()] = v.String()
		}
	})
	return m
}

func mapToLuaTable(L *lua.LState, m map[string]interface{}) *lua.LTable {
	tbl := L.NewTable()
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			L.SetField(tbl, k, mapToLuaTable(L, val))
		case float64:
			L.SetField(tbl, k, lua.LNumber(val))
		case bool:
			L.SetField(tbl, k, lua.LBool(val))
		case string:
			L.SetField(tbl, k, lua.LString(val))
		default:
			L.SetField(tbl, k, lua.LString(fmt.Sprintf("%v", val)))
		}
	}
	return tbl
}

// suppress unused import
var _ = time.Now
