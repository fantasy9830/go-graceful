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
	Go(func(context.Context) error)
	RegisterOnShutdown(func() error)
	Done() <-chan struct{}
	Errors() <-chan error
}

type Manager struct {
	mu              sync.Mutex
	wg              sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
	onShutdown      []func() error
	managerCtx      context.Context
	managerCancel   context.CancelFunc
	inShutdown      atomic.Bool
	errChan         chan error
	isErrChanReader atomic.Bool
}

func NewManager(optFuncs ...OptionFunc) GracefulManager {
	// default options
	managerCtx, managerCancel := context.WithCancel(context.Background())
	m := &Manager{
		ctx:           context.Background(),
		managerCtx:    managerCtx,
		managerCancel: managerCancel,
		errChan:       make(chan error, 1),
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

func (m *Manager) Go(f func(context.Context) error) {
	if m.shuttingDown() {
		log.Print("service shutting down ...")
		return
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		err := f(m.ctx)
		if err != nil && m.isErrChanRead() {
			select {
			case m.errChan <- err:
			default:
				log.Print("cannot write error to error channel, it is not read")
			}
		}
	}()
}

func (m *Manager) RegisterOnShutdown(f func() error) {
	if m.shuttingDown() {
		log.Print("service shutting down ...")
		return
	}

	m.mu.Lock()
	m.onShutdown = append(m.onShutdown, f)
	m.mu.Unlock()
}

func (m *Manager) Errors() <-chan error {
	m.setErrChanRead()
	return m.errChan
}

func (m *Manager) Done() <-chan struct{} {
	return m.managerCtx.Done()
}

func (m *Manager) shutdown() {
	for _, shutdownFunc := range m.onShutdown {
		m.wg.Add(1)
		go func(f func() error) {
			defer m.wg.Done()

			err := f()
			if err != nil && m.isErrChanRead() {
				select {
				case m.errChan <- err:
				default:
					log.Print("cannot write error to error channel, it is not read")
				}
			}
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
		close(m.errChan)
		m.errChan = nil
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

func (m *Manager) isErrChanRead() bool {
	return m.isErrChanReader.Load()
}

func (m *Manager) setErrChanRead() {
	m.isErrChanReader.Store(true)
}
