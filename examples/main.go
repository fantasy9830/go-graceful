package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/fantasy9830/go-graceful"
)

func main() {
	options := []graceful.OptionFunc{
		graceful.WithContext(context.Background()),
	}

	m := graceful.NewManager(options...)

	go func() {
		for err := range m.Errors() {
			log.Printf("error: %s\n", err.Error())
		}
	}()

	m.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				log.Print("graceful shutdown 01")
				return nil
			default:
				log.Print("goroutine 01")
				time.Sleep(1 * time.Second)
			}
		}
	})

	m.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				log.Print("graceful shutdown 02")
				return nil
			default:
				log.Print("goroutine 02")
				time.Sleep(2 * time.Second)
				return errors.New("graceful shutdown 02")
			}
		}
	})

	m.RegisterOnShutdown(func() error {
		log.Print("graceful shutdown 03")
		time.Sleep(1 * time.Second)
		return nil
	})

	m.RegisterOnShutdown(func() error {
		log.Print("graceful shutdown 04")
		time.Sleep(2 * time.Second)
		return nil
	})

	<-m.Done()
}
