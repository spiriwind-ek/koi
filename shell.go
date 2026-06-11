package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
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

	// Tab completions
	completes := []string{
		"fs.mkdir(", "fs.write(", "fs.read(", "fs.ls(", "fs.rm(", "fs.exists(",
		"math.mat_new(", "math.mat_mul(", "math.mat_transpose(", "math.mat_determinant(",
		"math.mat_inverse(", "math.mat_print(", "math.mat_shape(",
		"math.tensor_new(", "math.tensor_print(", "math.tensor_shape(",
		"math.dot(", "math.norm(", "math.cross(",
		"io.print(", "os.time()", "os.clock()", "os.version()", "os.edition()",
		"serialize(", "help", "exit", "quit",
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "\033[31m>\033[0m ",
		HistoryFile:  "/tmp/.koi_history",
		AutoComplete: readline.NewPrefixCompleter(
			readline.PcItem("fs.mkdir("),
			readline.PcItem("fs.write("),
			readline.PcItem("fs.read("),
			readline.PcItem("fs.ls("),
			readline.PcItem("fs.rm("),
			readline.PcItem("fs.exists("),
			readline.PcItem("math.mat_new("),
			readline.PcItem("math.mat_mul("),
			readline.PcItem("math.mat_transpose("),
			readline.PcItem("math.mat_determinant("),
			readline.PcItem("math.mat_inverse("),
			readline.PcItem("math.mat_print("),
			readline.PcItem("math.mat_shape("),
			readline.PcItem("math.tensor_new("),
			readline.PcItem("math.tensor_print("),
			readline.PcItem("math.tensor_shape("),
			readline.PcItem("math.dot("),
			readline.PcItem("math.norm("),
			readline.PcItem("math.cross("),
			readline.PcItem("io.print("),
			readline.PcItem("os.time()"),
			readline.PcItem("os.clock()"),
			readline.PcItem("os.version()"),
			readline.PcItem("os.edition()"),
			readline.PcItem("serialize("),
			readline.PcItem("help"),
			readline.PcItem("exit"),
			readline.PcItem("quit"),
		),
	})
	if err != nil {
		fmt.Printf("Failed to init readline: %v\n", err)
		return
	}
	defer rl.Close()

	fmt.Println("Koi 0.0.1-Alpha — Lua Shell")
	fmt.Println("Type 'help' for commands, 'exit' to quit. Tab for completion.")
	fmt.Println()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt || err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		line = strings.TrimSpace(line)
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

		// Try as expression first (auto-print), then as statement
		err = L.DoString("local __r = (" + line + "); if __r ~= nil then print(__r) end")
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
  help                  Show this help
  exit / quit           Exit the shell

Filesystem:
  fs.mkdir("/path")           Create directory
  fs.write("/path", val)      Write value
  fs.read("/path")            Read value
  fs.ls("/path")              List contents
  fs.rm("/path")              Delete
  fs.exists("/path")          Check existence

Matrix:
  math.mat_new("/data/A", 2, 2, {1,2,3,4})
  math.mat_mul("/data/A", "/data/B", "/data/C")
  math.mat_transpose("/data/A", "/data/AT")
  math.mat_determinant("/data/A")
  math.mat_inverse("/data/A", "/data/AI")
  math.mat_print("/data/A")
  math.mat_shape("/data/A")

Tensor:
  math.tensor_new("/data/T", {2,3,4}, {1,2,3,4,5,6,...})
  math.tensor_print("/data/T")
  math.tensor_shape("/data/T")

Vector:
  math.dot({1,2,3}, {4,5,6})     Dot product
  math.norm({1,2,3})             Norm
  math.cross({1,0,0}, {0,1,0})   Cross product (3D)

System:
  os.time()           Unix timestamp
  os.clock()          CPU time
  os.version()        Koi version
  os.edition()        Edition (full/lite)
  serialize(value)    Serialize any value to string

Shortcuts:
  Tab                 Auto-complete
  ↑/↓                 Browse history
  Ctrl+C              Cancel / Exit
  Ctrl+L              Clear screen

Note: Paths must be quoted!
  ✗ fs.mkdir(test)
  ✓ fs.mkdir("/test")`)
}
