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

	// Override io.print to stdout
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

	// Also override io.print
	if ioTbl, ok := L.GetGlobal("io").(*lua.LTable); ok {
		L.SetField(ioTbl, "print", printFn)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Koi 0.1.0-mvp — Lua Shell")
	fmt.Println("Type Lua code or 'exit' to quit.")
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

		// Try as expression first (print result), then as statement
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
