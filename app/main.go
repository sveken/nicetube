package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Set the Logger to be global
var logger *slog.Logger

// Set global buffer reader
var stdoutBuf, stderrBuf bytes.Buffer

// Set the Mutex map up for global use
var mm = NewMutexMap()

// Set some Global Variables from the environment / default
var maxDuration int = 120
var maxvideoage int = 24

func main() {
	fmt.Println("You are running version 0.7.2")
	// Reading any command line flags and adjust the config
	//When we go to docker the start up bach script should do this passing the envoirmetnal variables to the flag
	//Not used yet
	addr := flag.String("addr", ":8085", "HTTP Server address")
	flag.IntVar(&maxDuration, "maxDuration", 120, "Max Video Duration in minutes")
	flag.IntVar(&maxvideoage, "max-video-age", 24, "The max age of a video before it is deleted by the cleaner in hours. Set to 0 to disable cleaner")
	flag.Parse()
	fmt.Printf("Max Duration of videos is set too %v minutes", maxDuration)
	fmt.Println()
	fmt.Printf("Max Video age has been set to %v hours", maxvideoage)
	//fmt.Printf("Max video age flag is %v", maxvideoage)
	fmt.Println()

	//fmt.Printf("Running with options %s", *addr)
	mux := http.NewServeMux()
	//Change this to /Video for container use...actually Just bind the Video folder under this in docker compose
	fileserver := http.FileServer(http.Dir("./Videos/"))

	//URL Time
	mux.HandleFunc("GET /{$}", homepage)
	mux.Handle("GET /Videos/", http.StripPrefix("/Videos", disablelisting(fileserver)))
	mux.HandleFunc("/reso/", GetResoVideos)
	mux.HandleFunc("GET /health", healthcheck)

	// Logging stuffs
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Starting the webserver on", "addr", *addr)
	//Allow users to disable the cleaner.
	if maxvideoage > 0 {
		go cleaner()
	} else {
		fmt.Println("The cleaner has been disabled. Please monitor your disk usage.")
	}
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

func urlhelper(path, QualitySelector string) (string, string, error) {
	QualityName := fmt.Sprintf("%s/", QualitySelector)
	// We split out the URL so we are just left with the URL at the end
	pathParts := strings.Split(path, QualityName)
	if len(pathParts) < 2 || len(pathParts[1]) < 3 {
		return "", "", errors.New("error: missing or invalid video URL")
	}
	// fixing the https:/ thingy with https:// and getting the final video URL
	VideoURL := strings.Replace(pathParts[1], "https:/", "https://", 1)
	// I use this line for troubleshooting cuz i dont string
	fmt.Printf("The URL we are attempting to download is %s", VideoURL)

	if strings.Contains(VideoURL, "watch?v=") {
		// Parse the URL to safely extract the query parameter for the v= links
		parsedURL, err := url.Parse(VideoURL)
		if err != nil {
			return "", "", errors.New("error: failed to parse video URL")
		}
		// Extract the "v" query parameter as the VideoID
		VideoID := parsedURL.Query().Get("v")
		if VideoID == "" {
			return "", "", errors.New("error: missing video ID in URL")
		}
		//When we are using a V= type URL, just replace the VideoURL with the VideoID
		VideoURL = VideoID
		logger.Info("Getting Youtube video", "URL", VideoURL, "VideoID", VideoID)
		return VideoURL, VideoID, nil
	} else if strings.Contains(VideoURL, "youtu.be") {
		parsedURL, err := url.Parse(VideoURL)
		if err != nil {
			return "", "", errors.New("error: failed to parse video URL")
		}
		// extract only the ID before the ?
		VideoID := strings.TrimPrefix(parsedURL.Path, "/")
		if VideoID == "" {
			return "", "", errors.New("error: missing video ID in URL")
		}
		//This should only return the code before the ? as its its own part?
		VideoURL = VideoID
		logger.Info("Getting Youtube video", "URL", VideoURL, "VideoID", VideoID)
		return VideoURL, VideoID, nil

	} else {
		// Just take the entire string after the last /
		lastSlashPosition := strings.LastIndex(pathParts[1], "/")
		VideoID := pathParts[1][lastSlashPosition+1:]
		logger.Info("Getting Youtube video", "URL", VideoURL, "VideoID", VideoID)
		return VideoURL, VideoID, nil
	}

}

// Makes the folder the video lives under using the last bit of the string from the video.
// This way the video could have the the same name as another video but its still treated as different
// Added Quality in too so people can request the same video at different levels incase the cache isn't cleared in time
func foldergen(VideoID string, savedir string, QualitySelector string) (string, error) {
	savedir = fmt.Sprintf("./Videos/%s/%s", VideoID, QualitySelector)
	err := os.MkdirAll(savedir, 0755) // Permissions: rwxr-xr-x
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return savedir, err
	}
	return savedir, err
}
