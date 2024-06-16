package main

import (
	"fmt"
	"github.com/mdhender/otto/internal/server"
	"log"
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
	log.Printf("listening on %s", app.BaseURL())
	err = app.ListenAndServe()
	if err != nil {
		return fmt.Errorf("cannot start server: %w", err)
	}

	return nil
}
