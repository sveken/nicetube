package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

//This health check now checks if the Video directory can be written too and also if botblocking has been triggered.

func healthcheck(addr string) {
	// Remove leading colon if present
	serverAddr := strings.TrimPrefix(addr, ":")
	if serverAddr == "" {
		serverAddr = "8085" // fallback to default port
	}

	healthURL := fmt.Sprintf("http://localhost:%s/health", serverAddr)
	resp, err := http.Get(healthURL)
	if err != nil {
		fmt.Println("Health check failed:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Health check failed with status: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Println("Health check passed")
	os.Exit(0)
}

func healthservice(w http.ResponseWriter, r *http.Request) {
	dir, err := os.Open("./Videos")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // This sets the error respone to error 500.
		fmt.Fprint(w, "Failed to read Video directory")
	} else if botblocked {
		w.WriteHeader(http.StatusInternalServerError) // This sets the error respone to error 500.
		fmt.Fprint(w, "Youtube has blocked the last attempt to download a video.")
	} else {
		w.WriteHeader(http.StatusOK) // Sets 200 OK status code
		fmt.Fprint(w, "Healthy")
	}
	defer dir.Close()

}
