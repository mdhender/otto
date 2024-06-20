package server

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	"github.com/mdhender/otto/internal/database"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

type Server struct {
	http.Server
	scheme string
	host   string
	port   string
	paths  struct {
		assets    string // path to public assets
		templates string // path to template files
	}
	debug struct {
		traceAssets bool // trace asset requests
	}
	db  database.Service
	dev struct {
		mode     bool
		magic    string
		password string
	}
	fileServer http.Handler
	magic      struct {
		link   string
		handle string
	}
}

func NewServer(options ...Option) (*Server, error) {
	s := &Server{
		scheme: "http",
		host:   "localhost",
		port:   "3000",
	}
	s.IdleTimeout, s.ReadTimeout, s.WriteTimeout = 10*time.Second, 5*time.Second, 10*time.Second
	s.MaxHeaderBytes = 1 << 20 // about 1MB
	s.magic.link, s.magic.handle = uuid.New().String(), "otto"

	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	log.Printf("[server] assets    %q\n", s.paths.assets)
	log.Printf("[server] templates %q\n", s.paths.templates)

	if s.host == "" && s.port == "" {
		return nil, errors.New("host and port cannot both be empty")
	}
	s.Addr = net.JoinHostPort(s.host, s.port)

	mux, err := s.RegisterRoutes()
	if err != nil {
		return nil, err
	} else {
		s.Handler = mux
	}

	return s, nil
}

func (s *Server) BaseURL() string {
	return fmt.Sprintf("%s://%s", s.scheme, s.Addr)
}

func (s *Server) MagicLink() string {
	return fmt.Sprintf("/sign-up/%s", s.magic.link)
}

type Option func(*Server) error

func WithAssets(path string) Option {
	return func(s *Server) (err error) {
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
		if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return errors.New("path is not a directory")
		}
		s.paths.assets = path
		return nil
	}
}

func WithDevMode(magic, handle, password string) Option {
	return func(s *Server) (err error) {
		if magic == "" {
			return fmt.Errorf("magic is required in development mode")
		} else if strings.TrimSpace(magic) != magic {
			return fmt.Errorf("magic cannot contain whitespace")
		}
		if password == "" {
			return fmt.Errorf("password is required in development mode")
		} else if strings.TrimSpace(password) != password {
			return fmt.Errorf("password cannot contain whitespace")
		}
		s.dev.mode = true
		s.dev.magic = magic
		s.dev.password = password
		return nil
	}
}

func WithHost(host string) Option {
	return func(s *Server) (err error) {
		s.host = strings.TrimSpace(host)
		s.Addr = net.JoinHostPort(s.host, s.port)
		return nil
	}
}

func WithPort(port string) Option {
	return func(s *Server) (err error) {
		s.port = strings.TrimSpace(port)
		s.Addr = net.JoinHostPort(s.host, s.port)
		return nil
	}
}

func WithTemplates(path string) Option {
	return func(s *Server) (err error) {
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
		if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return errors.New("path is not a directory")
		}
		s.paths.templates = path
		return nil
	}
}

func optionsCors(next http.Handler) http.Handler {
	log.Printf("[server] adding cors middleware\n")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// inject CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "DELETE, GET, HEAD, OPTIONS, POST, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// handle CORS
		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleBadRunes(next http.Handler) http.Handler {
	log.Printf("[server] adding bad runes middleware\n")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//log.Printf("[server] running bad runes middleware\n")
		// return an error if the URL contains any non-printable runes.
		for _, ch := range r.URL.Path {
			if !unicode.IsPrint(ch) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
