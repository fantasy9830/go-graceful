package graceful

import (
	"context"
)

type OptionFunc func(*Manager)

func WithContext(parent context.Context) OptionFunc {
	return func(m *Manager) {
		m.ctx = parent
	}
}
