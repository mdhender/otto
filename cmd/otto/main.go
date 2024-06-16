package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/mdhender/otto/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	app, err := server.NewServer()
	if err != nil {
		return fmt.Errorf("cannot create server: %w", err)
	}

	// set up stuff so that we can gracefully shut down the server and application
	serverCh := make(chan struct{})
	go func() {
		log.Printf("[server] serving %q\n", app.BaseURL())
		if err := app.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[server] exited with: %v", err)
		}
		close(serverCh)
	}()

	// create a catch for signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt
	<-signalCh

	// use the context to shut down the application
	log.Printf("[server] received interrupt, shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Shutdown(ctx); err != nil {
		log.Fatalf("[server] failed to shutdown server: %s", err)
	}

	// If we got this far, it was an interrupt, so don't exit cleanly
	return fmt.Errorf("interrupted and stopped")
}
