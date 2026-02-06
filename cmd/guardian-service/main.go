package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/chuxorg/chux-agent-mesh/internal/guardianservice"
)

func main() {
	port := getenvInt("GUARDIAN_PORT", 8788)
	baseURL := os.Getenv("CODEX_APP_SERVER_URL")
	endpoint := os.Getenv("CODEX_APP_SERVER_ENDPOINT")
	token := os.Getenv("CODEX_APP_SERVER_TOKEN")

	var engine guardianservice.ProposalEngine
	if baseURL == "" {
		engine = guardianservice.UnavailableEngine{Reason: "CODEX_APP_SERVER_URL not set"}
	} else {
		engine = guardianservice.NewAppServerEngine(baseURL, token, endpoint)
	}

	service := guardianservice.NewService(engine)
	addr := ":" + strconv.Itoa(port)
	server := &http.Server{
		Addr:         addr,
		Handler:      service.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("guardian_service_starting addr=%s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("guardian_service_error err=%v", err)
	}
}

func getenvInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
