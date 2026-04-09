package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"hybrid-rtc-platform/server/internal/rooms"
	"hybrid-rtc-platform/server/internal/signaling"
	"hybrid-rtc-platform/server/internal/ws"
)

func main() {
	addr := envOrDefault("SERVER_ADDR", ":8080")

	roomManager := rooms.NewManager()
	router := signaling.NewRouter(roomManager)
	handler := ws.NewHandler(router)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/ws", handler)

	server := &http.Server{
		Addr:              addr,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("signaling server listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server stopped: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s from %s in %s", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}
