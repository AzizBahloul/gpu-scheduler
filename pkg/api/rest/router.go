package rest

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates and configures the HTTP router
func NewRouter(handlers *Handlers) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(corsMiddleware)

	// Health check
	r.Get("/health", handlers.HealthCheckHandler)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Jobs
		r.Post("/jobs", handlers.SubmitJobHandler)
		r.Get("/jobs", handlers.ListJobsHandler)
		r.Get("/jobs/{jobID}", handlers.GetJobStatusHandler)
		r.Delete("/jobs/{jobID}", handlers.CancelJobHandler)

		// Tenants
		r.Post("/tenants", handlers.CreateTenantHandler)

		// Cluster
		r.Get("/cluster/status", handlers.GetClusterStatusHandler)
	})

	return r
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
