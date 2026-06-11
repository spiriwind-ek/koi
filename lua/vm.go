package lua

import (
	"fmt"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
	"koi/storage"
)

type VMPool struct {
	db   *storage.DB
	pool chan *lua.LState
	mu   sync.Mutex
}

func NewVMPool(db *storage.DB) *VMPool {
	return &VMPool{
		db:   db,
		pool: make(chan *lua.LState, 4),
	}
}

func (p *VMPool) Get() *lua.LState {
	select {
	case L := <-p.pool:
		return L
	default:
		return p.create()
	}
}

func (p *VMPool) Put(L *lua.LState) {
	select {
	case p.pool <- L:
	default:
		L.Close()
	}
}

func (p *VMPool) create() *lua.LState {
	L := lua.NewState()
	SetupSandbox(L, p.db)
	return L
}

func (p *VMPool) Execute(code string, timeout time.Duration) (string, error) {
	L := p.Get()
	defer p.Put(L)

	var output string

	done := make(chan error, 1)
	go func() {
		done <- L.DoString(code)
	}()

	select {
	case <-time.After(timeout):
		return "", fmt.Errorf("execution timeout (%v)", timeout)
	case err := <-done:
		if err != nil {
			return "", err
		}
		return output, nil
	}
}

func (p *VMPool) Close() {
	for {
		select {
		case L := <-p.pool:
			L.Close()
		default:
			return
		}
	}
}
