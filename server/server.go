package server

import (
	"database/sql"
	"log"
	"net/http"

	"gitea.com/ypatios/forum/handlers"
)

// Server represents the HTTP server for the forum application
type Server struct {
	DB     *sql.DB
	Router *http.ServeMux
}

// NewServer creates a new server instance with the provided database connection
func NewServer(db *sql.DB) *Server {
	s := &Server{
		DB:     db,
		Router: http.NewServeMux(),
	}

	// Register routes
	s.registerRoutes()

	return s
}

// registerRoutes sets up all the routes for the server
func (s *Server) registerRoutes() {
	// Home route
	s.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		handlers.HomeHandler(w, r, s.DB)
	})

	// Category filter route - compatible with all Go versions
	s.Router.HandleFunc("/categories/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		handlers.CategoryFilterHandler(w, r, s.DB)
	})
	s.Router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		handlers.Register(w, r)
	})
	s.Router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		handlers.Login(w, r)
	})
}

// Start begins listening for HTTP requests on the specified address
func (s *Server) Start(addr string) error {
	log.Printf("Server starting on %s", addr)
	return http.ListenAndServe(addr, s.Router)
}
