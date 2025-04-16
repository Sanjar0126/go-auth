
package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	
	var (
		dbURL = flag.String("db-url", getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/go_firebase?sslmode=disable"), "Database connection URL")
		workerCount = flag.Int("workers", 2, "Number of worker goroutines")
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

	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	
	log.Printf("Starting %d background workers", *workerCount)
	for i := 0; i < *workerCount; i++ {
		go runWorker(ctx, i, db)
	}

	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	
	<-quit
	log.Println("Shutting down workers...")
	cancel()
	
	
	time.Sleep(2 * time.Second)
	log.Println("Workers exited")
}


func runWorker(ctx context.Context, id int, db *sql.DB) {
	log.Printf("Worker %d started", id)
	defer log.Printf("Worker %d stopped", id)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			
			cleanupExpiredSessions(ctx, db)
		}
	}
}


func cleanupExpiredSessions(ctx context.Context, db *sql.DB) {
	result, err := db.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at < NOW()")
	if err != nil {
		log.Printf("Error cleaning up expired sessions: %v", err)
		return
	}

	count, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return
	}

	if count > 0 {
		log.Printf("Cleaned up %d expired sessions", count)
	}
}


func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}