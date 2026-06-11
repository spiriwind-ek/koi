package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"koi/config"
	luavm "koi/lua"
	"koi/storage"
)

func runShell(db *storage.DB, cfg *config.Config) {
	pool := luavm.NewVMPool(db, cfg)
	L := pool.Get()
	defer pool.Discard(L)

	printFn := L.NewFunction(func(L *lua.LState) int {
		top := L.GetTop()
		args := make([]interface{}, top)
		for i := 1; i <= top; i++ {
			args[i-1] = L.Get(i).String()
		}
		fmt.Println(args...)
		return 0
	})
	L.SetGlobal("print", printFn)

	if ioTbl, ok := L.GetGlobal("io").(*lua.LTable); ok {
		L.SetField(ioTbl, "print", printFn)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Koi 0.1.0-mvp — Lua Shell")
	fmt.Println("Type 'help' for available commands, 'exit' to quit.")
	fmt.Println()

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}
		if line == "help" {
			printHelp()
			continue
		}

		err := L.DoString("local _r = " + line + "; if _r ~= nil then print(_r) end")
		if err != nil {
			err = L.DoString(line)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	}

	fmt.Println("Bye!")
}

func printHelp() {
	fmt.Println(`Commands:
  help                Show this help
  exit / quit         Exit the shell

Filesystem:
  fs.mkdir(path)              Create directory
  fs.write(path, value)       Write value to path
  fs.read(path)               Read value from path
  fs.ls(path)                 List directory contents
  fs.rm(path)                 Delete node
  fs.exists(path)             Check if path exists

Matrix:
  math.mat_new(path, r, c, data)    Create matrix
  math.mat_mul(a, b, result)        Matrix multiply
  math.mat_transpose(path, result)  Transpose
  math.mat_determinant(path)        Determinant
  math.mat_inverse(path, result)    Inverse
  math.mat_print(path)              Print matrix
  math.mat_shape(path)              Get shape {rows, cols}

Tensor:
  math.tensor_new(path, shape, data)  Create tensor
  math.tensor_print(path)             Print tensor
  math.tensor_shape(path)             Get shape

Vector:
  math.dot(a, b)      Dot product
  math.norm(a)        L2 norm
  math.cross(a, b)    Cross product (3D)

System:
  os.time()           Unix timestamp
  os.clock()          CPU time
  os.version()        Koi version
  os.edition()        Edition (full/lite)

Examples:
  math.mat_new("/data/A", 2, 2, {1,2,3,4})
  math.mat_print("/data/A")
  math.mat_determinant("/data/A")
  fs.write("/home/user/x", 42)
  fs.read("/home/user/x")`)
}
