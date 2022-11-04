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
	manager *Manager
	once    sync.Once
)

type GracefulManager interface {
	Go(f func(context.Context))
	RegisterOnShutdown(f func())
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
	once.Do(func() {
		// default options
		managerCtx, managerCancel := context.WithCancel(context.Background())
		manager = &Manager{
			ctx:           context.Background(),
			managerCtx:    managerCtx,
			managerCancel: managerCancel,
		}

		for _, applyFunc := range optFuncs {
			applyFunc(manager)
		}

		manager.ctx, manager.cancel = signal.NotifyContext(manager.ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		go manager.gracefulShutdown()
	})

	return manager
}

func GetManager() GracefulManager {
	return NewManager()
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

		m.mu.Lock()
		// graceful shutdown finish
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
