package main

import (
	"errors"
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

func urlhelper(path, prefix string) (string, string, error) {
	pathParts := strings.Split(path, prefix)
	if len(pathParts[1]) < 3 {
		return "", "", errors.New("missing or invalid video URL")
	}
	// fixing the https:/ thingy with https:// and getting the final video URL
	VideoURL := strings.Replace(pathParts[1], "https:/", "https://", 1)
	//Generate a unique ID to track the video based off string at the end of the video URL
	lastSlashPosition := strings.LastIndex(pathParts[1], "/")
	VideoID := pathParts[1][lastSlashPosition+1:]
	logger.Info("Getting Youtube video", "URL", VideoURL, "VideoID", VideoID)
	return VideoURL, VideoID, nil

}

// Makes the folder the video lives under using the last bit of the string from the video.
// This way the video could have the the same name as another video but its still treated as different
func foldergen(VideoID string) error {
	err := os.MkdirAll(fmt.Sprintf("./Videos/%s", VideoID), 0755) // Permissions: rwxr-xr-x
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return err
	}
	return err
}
