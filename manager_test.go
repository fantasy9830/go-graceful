package graceful_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fantasy9830/go-graceful"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Suit struct {
	suite.Suite
}

func TestGracefulManager(t *testing.T) {
	suite.Run(t, new(Suit))
}

func (s *Suit) TestWithContext() {
	var count atomic.Int32

	c, cancel := context.WithCancel(context.Background())

	m := graceful.NewManager(graceful.WithContext(c))
	m.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				count.Add(1)
				return nil
			default:
				time.Sleep(100 * time.Millisecond)
				count.Add(1)
			}
		}
	})

	m.RegisterOnShutdown(func() error {
		count.Add(1)
		return nil
	})

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	<-m.Done()

	assert.Equal(s.T(), int32(3), count.Load())
}

func (s *Suit) TestInShuttingDown() {
	var count atomic.Int32

	c, cancel := context.WithCancel(context.Background())

	m := graceful.NewManager(graceful.WithContext(c))
	m.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				count.Add(1)
				return nil
			default:
				time.Sleep(100 * time.Millisecond)
				count.Add(1)
			}
		}
	})

	m.RegisterOnShutdown(func() error {
		count.Add(1)
		return nil
	})

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		m.Go(func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					count.Add(1)
					return nil
				default:
					time.Sleep(100 * time.Millisecond)
					count.Add(1)
				}
			}
		})
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		m.RegisterOnShutdown(func() error {
			count.Add(1)
			return nil
		})
	}()

	<-m.Done()

	assert.Equal(s.T(), int32(3), count.Load())
}

func (s *Suit) TestErrors() {
	c, cancel := context.WithCancel(context.Background())

	m := graceful.NewManager(graceful.WithContext(c))
	go func() {
		err := <-m.Errors()
		assert.Equal(s.T(), "test", err.Error())
	}()

	m.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				return errors.New("test")
			}
		}
	})

	m.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				time.Sleep(100 * time.Millisecond)
				return errors.New("test")
			}
		}
	})

	m.RegisterOnShutdown(func() error {
		return errors.New("test")
	})

	m.RegisterOnShutdown(func() error {
		return errors.New("test")
	})

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	<-m.Done()
}

func (s *Suit) TestGo() {
	var count atomic.Int32

	c, cancel := context.WithCancel(context.Background())

	m := graceful.NewManager(graceful.WithContext(c))
	m.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				count.Add(1)
				return nil
			default:
				time.Sleep(100 * time.Millisecond)
				count.Add(1)
			}
		}
	})

	m.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				count.Add(1)
				return nil
			default:
				time.Sleep(100 * time.Millisecond)
				count.Add(1)
			}
		}
	})

	m.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				count.Add(1)
				return nil
			default:
				time.Sleep(100 * time.Millisecond)
				count.Add(1)
			}
		}
	})

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	<-m.Done()

	assert.Equal(s.T(), int32(6), count.Load())
}

func (s *Suit) TestRegisterOnShutdown() {
	var count atomic.Int32

	c, cancel := context.WithCancel(context.Background())

	m := graceful.NewManager(graceful.WithContext(c))
	m.RegisterOnShutdown(func() error {
		count.Add(1)
		return nil
	})

	m.RegisterOnShutdown(func() error {
		count.Add(1)
		return nil
	})

	m.RegisterOnShutdown(func() error {
		count.Add(1)
		return nil
	})

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	<-m.Done()

	assert.Equal(s.T(), int32(3), count.Load())
}

func (s *Suit) TestGetManager() {
	m := graceful.GetManager()
	assert.NotNil(s.T(), m)
}
