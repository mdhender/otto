package server

import (
	"errors"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/mdhender/otto/internal/database"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Server struct {
	http.Server
	scheme string
	host   string
	port   string
	paths  struct {
		public    string // path to public files
		templates string // path to template files
	}
	db database.Service
}

func NewServer(options ...Option) (*Server, error) {
	s := &Server{
		scheme: "http",
		host:   "localhost",
		port:   "3000",
	}
	s.IdleTimeout, s.ReadTimeout, s.WriteTimeout = 10*time.Second, 5*time.Second, 10*time.Second
	s.MaxHeaderBytes = 1 << 20 // about 1MB

	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}
	s.Addr = net.JoinHostPort(s.host, s.port)

	s.Handler = s.RegisterRoutes()

	return s, nil
}

func (s *Server) BaseURL() string {
	return fmt.Sprintf("%s://%s", s.scheme, s.Addr)
}

type Option func(*Server) error

func WithHost(host string) Option {
	return func(s *Server) (err error) {
		s.host = host
		s.Addr = net.JoinHostPort(s.host, s.port)
		return nil
	}
}

func WithPort(port string) Option {
	return func(s *Server) (err error) {
		s.port = port
		s.Addr = net.JoinHostPort(s.host, s.port)
		return nil
	}
}

func WithPublic(path string) Option {
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
		s.paths.public = path
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
