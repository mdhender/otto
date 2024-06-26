package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"github.com/mdhender/otto/frontend/admin"
	"github.com/mdhender/otto/frontend/authn"
	"github.com/mdhender/otto/frontend/hero"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Server) RegisterRoutes() (http.Handler, error) {
	// default mux, no routes
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", getHeroPage(s.paths.templates))
	mux.HandleFunc("GET /admin", getAdminShell(s.paths.templates))
	mux.HandleFunc("GET /admin/settings", getAdminSettings(s.paths.templates))
	mux.HandleFunc("GET /features", getFeaturesPage(s.paths.templates))
	mux.HandleFunc("GET /health", s.healthHandler)
	mux.HandleFunc("GET /login", getLoginPage(s.paths.templates, s.dev.mode, "otto", s.dev.password))
	mux.HandleFunc("POST /login", postLoginPage())
	mux.HandleFunc("GET /sign-up/{inviteId}", getInvitePage(s.paths.templates, s.magic.handle, s.magic.link))
	mux.HandleFunc("POST /sign-up/{inviteId}", postInvitePage(s.magic.handle, s.magic.link))

	// walk the frontend assets directory and add routes to serve static files
	validExtensions := map[string]bool{
		".css":    true,
		".exe":    true,
		".html":   true,
		".ico":    true,
		".jpg":    true,
		".js":     true,
		".png":    true,
		".robots": true,
		".svg":    true,
	}
	if err := filepath.WalkDir(s.paths.assets, func(path string, d os.DirEntry, err error) error {
		// don't serve unknown file types or dotfiles
		if err != nil || d.IsDir() || !validExtensions[filepath.Ext(path)] || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		route := "GET " + strings.TrimPrefix(path, s.paths.assets)
		//log.Printf("server: assets: path  %q\n", path)
		//log.Printf("server: assets: route %q\n", route)
		mux.Handle(route, getAsset("", s.paths.assets, s.debug.traceAssets))
		return nil
	}); err != nil {
		return nil, err
	}

	return mux, nil
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	stats := make(map[string]string)
	if s == nil {
		stats["status"] = "down"
		stats["error"] = "server is down"
	} else if s.db == nil {
		stats["status"] = "down"
		stats["error"] = "database is down"
	} else {
		stats = s.db.Health()
	}

	jsonResp, err := json.Marshal(stats)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonResp)
}

func getAdminShell(templatesPath string) http.HandlerFunc {
	templateFiles := []string{
		abstmpl(templatesPath, "admin", "shell.gohtml"),
		abstmpl(templatesPath, "admin", "desktop_sidebar_static.gohtml"),
		abstmpl(templatesPath, "admin", "mobile_menu_off_canvas.gohtml"),
		abstmpl(templatesPath, "admin", "nav_bar_secondary.gohtml"),
		abstmpl(templatesPath, "admin", "search_header_sticky.gohtml"),
		abstmpl(templatesPath, "admin", "account_settings.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		var payload admin.Shell
		payload.Content = admin.Blank{}

		render(w, r, payload, templateFiles...)
	}
}

func getAdminSettings(templatesPath string) http.HandlerFunc {
	templateFiles := []string{
		abstmpl(templatesPath, "admin", "shell.gohtml"),
		abstmpl(templatesPath, "admin", "desktop_sidebar_static.gohtml"),
		abstmpl(templatesPath, "admin", "mobile_menu_off_canvas.gohtml"),
		abstmpl(templatesPath, "admin", "nav_bar_secondary.gohtml"),
		abstmpl(templatesPath, "admin", "search_header_sticky.gohtml"),
		abstmpl(templatesPath, "admin", "account_settings.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		var payload admin.Shell
		payload.Content = admin.Blank{}

		render(w, r, payload, templateFiles...)
	}
}

// returns a handler that will serve an asset if it exists, otherwise return not found.
func getAsset(prefix, root string, trace bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if trace {
			log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
		}

		file := filepath.Join(root, filepath.Clean(strings.TrimPrefix(r.URL.Path, prefix)))

		stat, err := os.Stat(file)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// only serve regular files, never directories or directory listings.
		if stat.IsDir() || !stat.Mode().IsRegular() {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// pretty sure that we have a regular file at this point.
		rdr, err := os.Open(file)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		defer func(r io.ReadCloser) {
			_ = r.Close()
		}(rdr)

		// let Go serve the file. it does magic things like content-type, etc.
		http.ServeContent(w, r, file, stat.ModTime(), rdr)
	}
}

func getFeaturesPage(templatesPath string) http.HandlerFunc {
	templateFiles := []string{
		abstmpl(templatesPath, "hero", "page.gohtml"),
		abstmpl(templatesPath, "hero", "nav_bar.gohtml"),
		abstmpl(templatesPath, "hero", "features.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		var payload hero.Page
		payload.Title = "Otto * Features"
		payload.NavBar.PageName = "features"

		render(w, r, payload, templateFiles...)
	}
}

func getHeroPage(templatesPath string) http.HandlerFunc {
	templateFiles := []string{
		abstmpl(templatesPath, "hero", "page.gohtml"),
		abstmpl(templatesPath, "hero", "nav_bar.gohtml"),
		abstmpl(templatesPath, "hero", "landing.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" { // Go makes / the catch-all route
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		var payload hero.Page
		payload.Title = "Otto"
		payload.NavBar.PageName = "index"

		render(w, r, payload, templateFiles...)
	}
}

func getLoginPage(templatesPath string, devMode bool, handle, password string) http.HandlerFunc {
	templateFiles := []string{
		abstmpl(templatesPath, "authn", "page.gohtml"),
		abstmpl(templatesPath, "authn", "login.gohtml"),
	}

	var payload authn.Page
	payload.Title = "Otto * Login"
	if devMode {
		log.Printf("warning: getLoginPage: dev mode enabled!\n")
		payload.Content = authn.Login{
			DevMode:  true,
			Handle:   handle,
			Password: password,
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		render(w, r, payload, templateFiles...)
	}
}

func postLoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func getInvitePage(templatesPath string, magicHandle, magicInviteId string) http.HandlerFunc {
	log.Printf("[server] getInvitePage: handle %q: invite %q\n", magicHandle, magicInviteId)
	inviteClosed := []string{
		abstmpl(templatesPath, "authn", "page.gohtml"),
		abstmpl(templatesPath, "authn", "signup_closed.gohtml"),
	}
	h := sha256.New()
	h.Write([]byte(magicInviteId))
	inviteHash := h.Sum(nil)

	inviteOpen := []string{
		abstmpl(templatesPath, "authn", "page.gohtml"),
		abstmpl(templatesPath, "authn", "signup.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
		inviteId, validInvite := r.PathValue("inviteId"), false
		if inviteId != "" {
			h := sha256.New()
			h.Write([]byte(inviteId))
			validInvite = bytes.Equal(h.Sum(nil), inviteHash)
		}
		if !validInvite {
			var payload authn.Page
			payload.Title = "Otto * Sign Up"
			render(w, r, payload, inviteClosed...)
			return
		}

		var payload authn.Page
		payload.Title = "Otto * Sign Up"
		payload.Content = authn.SignUp{
			InviteLink: r.URL.Path,
			Handle:     magicHandle,
			Password:   magicInviteId,
		}

		render(w, r, payload, inviteOpen...)
	}
}

func postInvitePage(magicHandle, magicInviteId string) http.HandlerFunc {
	h := sha256.New()
	h.Write([]byte(magicInviteId))
	inviteHash := h.Sum(nil)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
		inviteId := r.PathValue("inviteId")
		if inviteId == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		} else {
			h := sha256.New()
			h.Write([]byte(inviteId))
			if !bytes.Equal(h.Sum(nil), inviteHash) {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// abstmpl is a helper function to return the absolute path to a template file.
// if the template file doesn't exist, it logs a warning and returns the invalid path.
func abstmpl(path ...string) string {
	tf, err := filepath.Abs(filepath.Join(path...))
	if err != nil {
		log.Printf("warning: template file %q is invalid", tf)
	} else if sb, err := os.Stat(tf); err != nil || sb.IsDir() {
		log.Printf("warning: template file %q does not exist", tf)
	} else if sb.IsDir() {
		log.Printf("warning: template file %q is a directory", tf)
	} else if !sb.Mode().IsRegular() {
		log.Printf("warning: template file %q is not a regular file", tf)
	}
	return tf
}

func render(w http.ResponseWriter, r *http.Request, payload any, templates ...string) {
	// parse the template file, logging any errors
	tmpl, err := template.ParseFiles(templates...)
	if err != nil {
		log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// execute the template with our payload, saving the response to a buffer so that we can capture errors in a nice way.
	buf := &bytes.Buffer{}
	if err = tmpl.Execute(buf, payload); err != nil {
		log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
