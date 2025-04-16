package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	httpHandlers "go-fr-project/api/http"
	"go-fr-project/internal/auth"
	"go-fr-project/internal/common/middleware"
)

func main() {
	var (
		port  = flag.Int("port", 8080, "API server port")
		dbURL = flag.String("db-url", getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/go_firebase?sslmode=disable"), "Database connection URL")
	)
	flag.Parse()

	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("Could not ping database: %v", err)
	}
	log.Println("Connected to database successfully")

	authService := auth.NewService(db)

	authMiddleware := middleware.NewAuthMiddleware(authService)

	authHandler := httpHandlers.NewAuthHandler(authService)

	mux := http.NewServeMux()

	authHandler.RegisterPublicRoutes(mux)

	protectedMux := http.NewServeMux()

	authHandler.RegisterProtectedRoutes(protectedMux)

	protectedHandler := authMiddleware.RequireAuth(protectedMux)

	mux.Handle("/api/protected/", http.StripPrefix("/api/protected", protectedHandler))

	handler := middleware.LogRequest(mux)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting API server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
