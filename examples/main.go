package main

import (
	"context"
	"log"
	"time"

	"github.com/fantasy9830/go-graceful"
)

func main() {
	options := []graceful.OptionFunc{
		graceful.WithContext(context.Background()),
	}

	m := graceful.NewManager(options...)

	m.Go(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Print("graceful shutdown 01")
				return
			default:
				log.Print("goroutine 01")
				time.Sleep(1 * time.Second)
			}
		}
	})

	m.Go(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Print("graceful shutdown 02")
				return
			default:
				log.Print("goroutine 02")
				time.Sleep(2 * time.Second)
			}
		}
	})

	m.RegisterOnShutdown(func() {
		log.Print("graceful shutdown 03")
		time.Sleep(1 * time.Second)
	})

	m.RegisterOnShutdown(func() {
		log.Print("graceful shutdown 04")
		time.Sleep(2 * time.Second)
	})

	<-m.Done()
}
