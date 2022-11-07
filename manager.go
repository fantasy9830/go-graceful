package graceful

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
)

var (
	manager GracefulManager
	once    sync.Once
)

type GracefulManager interface {
	Go(func(context.Context))
	RegisterOnShutdown(func())
	Done() <-chan struct{}
}

type Manager struct {
	mu            sync.Mutex
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	onShutdown    []func()
	managerCtx    context.Context
	managerCancel context.CancelFunc
	inShutdown    atomic.Bool
}

func NewManager(optFuncs ...OptionFunc) GracefulManager {
	// default options
	managerCtx, managerCancel := context.WithCancel(context.Background())
	m := &Manager{
		ctx:           context.Background(),
		managerCtx:    managerCtx,
		managerCancel: managerCancel,
	}

	for _, applyFunc := range optFuncs {
		applyFunc(m)
	}

	m.ctx, m.cancel = signal.NotifyContext(
		m.ctx,
		os.Interrupt,    // SIGINT, Ctrl+C
		syscall.SIGTERM, // systemd
	)

	go m.gracefulShutdown()

	return m
}

func GetManager() GracefulManager {
	once.Do(func() {
		manager = NewManager()
	})

	return manager
}

func (m *Manager) Go(f func(context.Context)) {
	if m.shuttingDown() {
		log.Print("service shutting down ...")
		return
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		f(m.ctx)
	}()
}

func (m *Manager) RegisterOnShutdown(f func()) {
	if m.shuttingDown() {
		log.Print("service shutting down ...")
		return
	}

	m.mu.Lock()
	m.onShutdown = append(m.onShutdown, f)
	m.mu.Unlock()
}

func (m *Manager) Done() <-chan struct{} {
	return m.managerCtx.Done()
}

func (m *Manager) shutdown() {
	for _, shutdownFunc := range m.onShutdown {
		m.wg.Add(1)
		go func(f func()) {
			defer m.wg.Done()

			f()
		}(shutdownFunc)
	}
}

func (m *Manager) gracefulShutdown() {
	<-m.ctx.Done()

	// graceful shutdown start
	m.setShuttingDown()

	m.cancel()

	m.shutdown()

	go func() {
		m.wg.Wait()

		// graceful shutdown finish
		m.mu.Lock()
		m.managerCancel()
		m.mu.Unlock()
	}()
}

func (m *Manager) shuttingDown() bool {
	return m.inShutdown.Load()
}

func (m *Manager) setShuttingDown() {
	m.inShutdown.Store(true)
}
