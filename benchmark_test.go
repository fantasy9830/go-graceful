package graceful_test

import (
	"context"
	"testing"

	"github.com/fantasy9830/go-graceful"
)

func BenchmarkGracefulGo(b *testing.B) {
	b.ReportAllocs()
	m := graceful.NewManager()
	for n := 0; n < b.N; n++ {
		m.Go(func(ctx context.Context) error {
			return nil
		})
	}
}

func BenchmarkGracefulShutdown(b *testing.B) {
	b.ReportAllocs()
	m := graceful.NewManager()
	for n := 0; n < b.N; n++ {
		m.RegisterOnShutdown(func() error {
			return nil
		})
	}
}
