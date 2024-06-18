package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/inconshreveable/mousetrap"
	"github.com/mdhender/otto/internal/server"
	"github.com/peterbourgon/ff/v3"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"
)

func main() {
	fs := flag.NewFlagSet("otto", flag.ExitOnError)

	assetsPath := filepath.Join("..", "frontend", "assets")
	fs.StringVar(&assetsPath, "assets", assetsPath, "override assets path")

	templatesPath := filepath.Join("..", "frontend")
	fs.StringVar(&templatesPath, "templates", templatesPath, "override templates path")

	err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("FH"), ff.WithConfigFileFlag("config"), ff.WithConfigFileParser(ff.JSONParser))
	if err != nil {
		log.Printf("error: %v\n", err)
		os.Exit(2)
	}

	defer func() {
		exitStatus := 0
		if r := recover(); r != nil {
			log.Printf("%s\n\n%s", r, debug.Stack())
			exitStatus = 1 // so that we can exit with a non-zero exit code
		}
		if mousetrap.StartedByExplorer() {
			fmt.Println("Press return to continue...")
			_, _ = fmt.Scanln()
		}
		os.Exit(exitStatus)
	}()

	err = run(assetsPath, templatesPath)
	if err != nil {
		log.Printf("error: %v\n", err)
	}
}

func run(assetsPath, templatesPath string) error {
	var options []server.Option
	options = append(options, server.WithHost("localhost"))
	options = append(options, server.WithPort("3000"))
	options = append(options, server.WithAssets(assetsPath))
	options = append(options, server.WithTemplates(templatesPath))

	app, err := server.NewServer(options...)
	if err != nil {
		return fmt.Errorf("cannot create server: %w", err)
	}

	// set up stuff so that we can gracefully shut down the server and application
	serverCh := make(chan struct{})
	go func() {
		log.Printf("[server] serving %q\n", app.BaseURL())
		log.Printf("[server] log in at %q\n", app.BaseURL()+"/user/login")
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
		return errors.Join(fmt.Errorf("[server] failed to shutdown server"), err)
	}

	// If we got this far, it was an interrupt, so don't exit cleanly
	return fmt.Errorf("interrupted and stopped")
}
