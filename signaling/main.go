package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	addrDefault := envOr("SIGNALING_ADDR", "127.0.0.1:8081")
	dbDefault := envOr("SIGNALING_DB_PATH", "./signaling.db")
	addr := flag.String("addr", addrDefault, "listen address")
	dbPath := flag.String("db", dbDefault, "SQLite database path")
	flag.Parse()

	server, err := newServer(*dbPath, serverOptions{
		AllowedOrigins:   splitCSV(envOr("SIGNALING_ALLOWED_ORIGINS", "http://localhost:5200")),
		CreatorVerifyKey: os.Getenv("ROOMWORKS_CREATOR_VERIFY_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	log.Printf("signaling API listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, server.Handler()))
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
