package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/pointlings/backend/internal/handlers"
	"github.com/pointlings/backend/internal/repository/postgres"
	"github.com/pointlings/backend/pkg/config"
)

func main() {
	cfg := config.Load()

	// Connect to Supabase Postgres via pgx driver
	db, err := sql.Open("pgx", cfg.SupabaseURL+"?sslmode=require")
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	pointlingRepo := postgres.NewPointlingRepository(db)
	xpRepo := postgres.NewXPRepository(db)
	itemRepo := postgres.NewItemRepository(db)
	pointlingItemRepo := postgres.NewPointlingItemRepository(db)
	pointSpendRepo := postgres.NewPointSpendRepository(db)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userRepo)
	pointlingHandler := handlers.NewPointlingHandler(pointlingRepo)
	xpHandler := handlers.NewXPHandler(xpRepo, pointlingRepo)
	itemHandler := handlers.NewItemHandler(itemRepo, pointlingRepo, pointlingItemRepo)
	pointHandler := handlers.NewPointHandler(pointSpendRepo, itemRepo, pointlingItemRepo)

	// Setup router with middlewares
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(30 * time.Second))

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Mount handlers
		userHandler.RegisterRoutes(r)
		pointlingHandler.RegisterRoutes(r)
		xpHandler.RegisterRoutes(r)
		itemHandler.RegisterRoutes(r)
		pointHandler.RegisterRoutes(r)
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create server
	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: r,
	}

	// Start server in goroutine for graceful shutdown
	go func() {
		log.Printf("HTTP server listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	// Gracefully shutdown with 10s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}

	log.Println("server stopped gracefully")
}
