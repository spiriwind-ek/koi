package lua

import (
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
	"koi/config"
	"koi/storage"
)

func SetupSandbox(L *lua.LState, db *storage.DB, cfg *config.Config) {
	// Remove dangerous functions
	L.SetGlobal("dofile", lua.LNil)
	L.SetGlobal("loadfile", lua.LNil)
	L.SetGlobal("require", lua.LNil)
	L.SetGlobal("module", lua.LNil)
	L.SetGlobal("package", lua.LNil)

	// Safe os table
	osTable := L.NewTable()
	L.SetField(osTable, "time", L.NewFunction(osTime))
	L.SetField(osTable, "clock", L.NewFunction(osClock))
	L.SetField(osTable, "version", L.NewFunction(osVersion))
	L.SetField(osTable, "edition", L.NewFunction(osEdition(cfg)))
	L.SetGlobal("os", osTable)

	// Safe io table
	ioTable := L.NewTable()
	L.SetField(ioTable, "print", L.NewFunction(ioPrint))
	L.SetGlobal("io", ioTable)

	// fs API
	fsTable := L.NewTable()
	L.SetField(fsTable, "mkdir", L.NewFunction(fsMkdir(db)))
	L.SetField(fsTable, "write", L.NewFunction(fsWrite(db)))
	L.SetField(fsTable, "read", L.NewFunction(fsRead(db)))
	L.SetField(fsTable, "ls", L.NewFunction(fsLs(db)))
	L.SetField(fsTable, "rm", L.NewFunction(fsRm(db)))
	L.SetField(fsTable, "exists", L.NewFunction(fsExists(db)))
	L.SetGlobal("fs", fsTable)

	// math API (gonum-powered, with security limits)
	mathTable := L.NewTable()
	L.SetField(mathTable, "mat_new", L.NewFunction(mathMatNew(db, cfg)))
	L.SetField(mathTable, "mat_mul", L.NewFunction(mathMatMul(db)))
	L.SetField(mathTable, "mat_transpose", L.NewFunction(mathMatTranspose(db)))
	L.SetField(mathTable, "mat_determinant", L.NewFunction(mathMatDeterminant(db)))
	L.SetField(mathTable, "mat_inverse", L.NewFunction(mathMatInverse(db)))
	L.SetField(mathTable, "mat_print", L.NewFunction(mathMatPrint(db)))
	L.SetField(mathTable, "mat_shape", L.NewFunction(mathMatShape(db)))
	L.SetField(mathTable, "tensor_new", L.NewFunction(mathTensorNew(db, cfg)))
	L.SetField(mathTable, "tensor_print", L.NewFunction(mathTensorPrint(db)))
	L.SetField(mathTable, "tensor_shape", L.NewFunction(mathTensorShape(db)))
	L.SetField(mathTable, "dot", L.NewFunction(mathDot))
	L.SetField(mathTable, "norm", L.NewFunction(mathNorm))
	L.SetField(mathTable, "cross", L.NewFunction(mathCross))
	L.SetGlobal("math", mathTable)
}

// os API
func osTime(L *lua.LState) int {
	L.Push(lua.LNumber(float64(time.Now().Unix())))
	return 1
}

func osClock(L *lua.LState) int {
	L.Push(lua.LNumber(float64(time.Now().UnixNano()) / 1e9))
	return 1
}

func osVersion(L *lua.LState) int {
	L.Push(lua.LString("koi 0.0.1-Alpha"))
	return 1
}

func osEdition(cfg *config.Config) lua.LGFunction {
	return func(L *lua.LState) int {
		L.Push(lua.LString(cfg.Engine.Edition))
		return 1
	}
}

func ioPrint(L *lua.LState) int {
	top := L.GetTop()
	args := make([]interface{}, top)
	for i := 1; i <= top; i++ {
		args[i-1] = L.Get(i).String()
	}
	L.Push(lua.LString(fmt.Sprint(args...)))
	return 1
}
