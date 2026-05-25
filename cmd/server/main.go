package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"opencodepod/frontend"
	"opencodepod/internal/config"
	"opencodepod/internal/docker"
	"opencodepod/internal/handlers"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	dm, err := docker.NewDockerManager(cfg)
	if err != nil {
		log.Fatalf("docker init: %v", err)
	}
	defer dm.Close()

	server := handlers.NewServer(cfg, dm)

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	// Serve frontend static files
	static, err := fs.Sub(frontend.FS, "dist")
	if err != nil {
		log.Fatalf("frontend fs: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(static)))

	srv := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: cors(mux),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	fmt.Printf("OpenCodePod listening on %s\n", cfg.ListenAddr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down...")
	if err := srv.Close(); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}

func cors(next http.Handler) http.Handler {
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
