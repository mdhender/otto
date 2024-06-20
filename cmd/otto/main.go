package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/inconshreveable/mousetrap"
	"github.com/mdhender/otto/internal/server"
	"github.com/mdhender/otto/internal/sqlc"
	"github.com/peterbourgon/ff/v3"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

	if len(os.Args) < 2 {
		fmt.Printf("Otto is a web application for creating TribeNet maps from turn reports.\n")
		fmt.Printf("The turn reports are stored in a local database, so you must specify\n")
		fmt.Printf("the path where we'll create that database.\n")
		fmt.Printf("\n")
		HelpConsole()
		fmt.Printf("\n")
		fmt.Printf("Once Otto is running, it will print a line with the URL. Plug that\n")
		fmt.Printf("into your browser and you're ready to go. Just remember that you can\n")
		fmt.Printf("lose data if you kill the application while it is writing to the local\n")
		fmt.Printf("database. Please use the browser link to shut it down.\n")
		fmt.Printf("\n")
		fmt.Printf("If you already know this and just want to see the command line options,\n")
		fmt.Printf("type `%s help`.\n", os.Args[0])
		fmt.Printf("\n")
		if mousetrap.StartedByExplorer() {
			// explorer launched us, so give the user a chance to see the output before we exit.
			// otherwise, the console window will close before the user sees the output.
			fmt.Printf("\n")
			fmt.Println("Press return to continue...")
			_, _ = fmt.Scanln()
			fmt.Printf("\n")
		} else {
			// wait a few seconds to give the user a chance to see the output before we exit.
			time.Sleep(3 * time.Second)
		}

		os.Exit(1)
	}

	fs := flag.NewFlagSet("otto", flag.ExitOnError)

	var databasePath string
	fs.StringVar(&databasePath, "db", databasePath, "path for database files")

	dev := false
	fs.BoolVar(&dev, "dev", dev, "run in development mode")

	devMagic, devHandle, devPassword := "95660abc-8ea6-4e42-82a3-0e791e2c49cc", "otto", ""
	fs.StringVar(&devPassword, "dev-password", devPassword, "password for development")

	assetsPath := filepath.Join("..", "frontend", "assets")
	fs.StringVar(&assetsPath, "assets", assetsPath, "override assets path")

	templatesPath := filepath.Join("..", "frontend")
	fs.StringVar(&templatesPath, "templates", templatesPath, "override templates path")

	err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("OTTO"), ff.WithConfigFileFlag("config"), ff.WithConfigFileParser(ff.JSONParser))
	if err != nil {
		log.Printf("error: %v\n", err)
		os.Exit(2)
	}

	//log.Printf("assetsPath   : %q\n", assetsPath)
	//log.Printf("databasePath : %q\n", databasePath)
	//log.Printf("templatesPath: %q\n", templatesPath)

	if dev && devPassword == "" {
		log.Fatalf("error: dev-password is required in development mode\n")
	} else if dev && strings.TrimSpace(devPassword) != devPassword {
		log.Fatalf("error: dev-password cannot contain whitespace\n")
	} else if !dev && devPassword != "" {
		log.Fatalf("error: password is not accepted in production mode\n")
	}

	// ugh. setting up the database here feels so wrong.
	if databasePath == "" {
		log.Fatalf("error: database path is required\n")
	} else if strings.TrimSpace(databasePath) != databasePath {
		log.Fatalf("error: database path cannot contain whitespace\n")
	} else if sb, err := os.Stat(databasePath); err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("error: %s: %v\n", databasePath, err)
		}
		log.Fatalf("error: %s: is not a directory\n", databasePath)
	} else if !sb.IsDir() {
		log.Fatalf("error: %s: is not a directory\n", databasePath)
	} else if path, err := filepath.Abs(databasePath); err != nil {
		log.Fatalf("error: %s: cannot determine absolute path: %v\n", databasePath, err)
	} else {
		databasePath = filepath.Join(path, "otto.sqlite")
	}

	if err = sqlc.CreateDatabase(databasePath); err != nil {
		log.Fatalf("error: %v\n", err)
	} else if err = sqlc.MigrateSchema(databasePath); err != nil {
		log.Fatalf("error: %v\n", err)
	}
	if dev {
		if err = sqlc.DevSchema(databasePath, devMagic, "otto", devPassword); err != nil {
			log.Fatalf("error: %v\n", err)
		}
	}

	err = run(databasePath, assetsPath, templatesPath, dev, devMagic, devHandle, devPassword)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
}

func run(databasePath, assetsPath, templatesPath string, dev bool, devMagic, devHandle, devPassword string) error {
	// ugh. setting up the database here feels so wrong.
	if databasePath == "" {
		return fmt.Errorf("database path is required")
	} else if strings.TrimSpace(databasePath) != databasePath {
		return fmt.Errorf("database path cannot contain whitespace")
	} else if path, err := filepath.Abs(databasePath); err != nil {
		return fmt.Errorf("cannot determine absolute path for database path: %w", err)
	} else if path != databasePath {
		return fmt.Errorf("database path cannot contain relative paths")
	} else if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("cannot stat database path: %w", err)
		}
	}

	var options []server.Option
	if dev {
		options = append(options, server.WithDevMode(devMagic, devHandle, devPassword))
	}
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
		log.Printf("[server] serving %s\n", app.BaseURL())
		log.Printf("[server] log in at %s/login\n", app.BaseURL())
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
