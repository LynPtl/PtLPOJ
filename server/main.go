package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pt_lpoj/api"
	"pt_lpoj/models"
	"pt_lpoj/sandbox"
	"pt_lpoj/scheduler"
	"pt_lpoj/storage"
)

func main() {
	log.Println("Initializing PtLPOJ Server (Phase 3)...")

	// 1. Initialize SQLite Database
	dbFile := "ptlpoj_dev.db"
	if err := storage.InitDB(dbFile); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	log.Println("Database connection established")

	// 2. Load Problem Repository
	if err := storage.InitProblemRepo("../data"); err != nil {
		log.Printf("Warning: Failed to load problems repository: %v", err)
		log.Println("Ensure to generate problems.json via tools/parse_questions.py")
	}

	// 3. Initialize Docker Client
	if err := sandbox.InitDockerClient(); err != nil {
		log.Fatalf("Docker initialization failed: %v", err)
	}

	// 4. Seed Init Admin User Whitelist
	seedErr := seedInitAdmin()
	if seedErr != nil {
		log.Printf("Note: Initial Admin user seeding info: %v", seedErr)
	}

	// 5. Start Worker Pool (Scheduler)
	jobQueue := scheduler.NewJobQueue(100)
	scheduler.SetGlobalQueue(jobQueue)
	scheduler.StartWorkerPool(2, jobQueue)

	// 6. Mount Router & Start Server
	router := api.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Printf("Server starting on http://localhost%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server crashed: %v", err)
		}
	}()

	// 7. Graceful Shutdown on SIGINT / SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutdown signal received, draining...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	jobQueue.Close()

	log.Println("Server shutdown complete")
}

// seedInitAdmin injects the admin email from PTLPOJ_INIT_ADMIN_EMAIL into the whitelist if set
func seedInitAdmin() error {
	adminEmail := os.Getenv("PTLPOJ_INIT_ADMIN_EMAIL")
	if adminEmail == "" {
		log.Println("PTLPOJ_INIT_ADMIN_EMAIL not set, skipping initial admin seeding.")
		return nil
	}

	user, err := storage.GetUserByEmail(adminEmail)
	if err == nil && user != nil {
		return nil
	}
	_, err = storage.CreateUser(adminEmail, models.RoleAdmin)
	if err != nil {
		return err
	}
	log.Printf("Initial admin account %s seeded into whitelist.", adminEmail)
	return nil
}
