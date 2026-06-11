package lua

import (
	"sync"

	lua "github.com/yuin/gopher-lua"
	"koi/config"
	"koi/storage"
)

type VMPool struct {
	db   *storage.DB
	cfg  *config.Config
	pool chan *lua.LState
	mu   sync.Mutex
	size int
}

func NewVMPool(db *storage.DB, cfg *config.Config) *VMPool {
	return &VMPool{
		db:   db,
		cfg:  cfg,
		pool: make(chan *lua.LState, 4),
		size: 4,
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

func (p *VMPool) Discard(L *lua.LState) {
	L.Close()
}

func (p *VMPool) create() *lua.LState {
	L := lua.NewState()
	SetupSandbox(L, p.db, p.cfg)
	return L
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
