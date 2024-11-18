package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

// Set the Logger to be global
var logger *slog.Logger

func main() {
	fmt.Println("Hello")
	// Reading any command line flags and adjust the config
	//When we go to docker the start up bach script should do this passing the envoirmetnal variables to the flag
	addr := flag.String("addr", ":8085", "HTTP Server address")
	flag.Parse()
	fmt.Printf("Running with options %s", *addr)
	mux := http.NewServeMux()
	//Change this to /Video for container use...actually Just bind the Video folder under this in docker compose
	fileserver := http.FileServer(http.Dir("./Videos/"))

	//URL Time
	mux.HandleFunc("GET /{$}", homepage)
	mux.Handle("GET /Videos/", http.StripPrefix("/Videos", disablelisting(fileserver)))
	mux.HandleFunc("/reso1080Ph264forced/", reso1080Ph264forced)

	// Logging stuffs
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Starting the webserver on", "addr", *addr)
	err := http.ListenAndServe(*addr, mux)
	logger.Error(err.Error())
	os.Exit(1)

}

func homepage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Place Holder Homepage")
}

// Disables directory listing
func disablelisting(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
