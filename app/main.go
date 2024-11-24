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

func main() {
	fmt.Println("Hello")
	// Reading any command line flags and adjust the config
	//When we go to docker the start up bach script should do this passing the envoirmetnal variables to the flag
	//Not used yet
	addr := flag.String("addr", ":8085", "HTTP Server address")
	flag.Parse()
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
	go cleaner()
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

func ReturnDownloadURL(savedir string, Domain string) (string, error) {
	//fmt.Printf("Save directory is %s and the URL from Path is %s", savedir, Domain)
	Videofile, err := GetFileName(savedir)
	if err != nil {
		return "", fmt.Errorf("no .mp4 or .webm file found in the directory")
	}
	//Remove first dot from save directory to make URL from
	URLFriendlyDirIndex := strings.Index(savedir, ".")
	URLFriendlyDir := savedir[:URLFriendlyDirIndex] + savedir[URLFriendlyDirIndex+1:]
	//fmt.Printf("And the Mp4 name is %s", mp4File)
	//fmt.Println()
	TheDownloadURL := fmt.Sprintf("http://%s%s/%s", Domain, URLFriendlyDir, Videofile)
	return TheDownloadURL, err
}

func GetFileName(savedir string) (string, error) {
	files, err := os.ReadDir(savedir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %v", err)
	}

	// Find the .mp4 file in the directory and extract its name
	var Videofile string
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".mp4") || strings.HasSuffix(file.Name(), ".webm")) {
			Videofile = file.Name()
			break
		}
	}

	// If no .mp4 file is found
	if Videofile == "" {
		return "", fmt.Errorf("no .mp4 or .webm file found in the directory")
	}

	return Videofile, err
}

// The Bot check
func containsBotCheck(output string) bool {
	return strings.Contains(strings.ToLower(output), "sign in to confirm")
}

func normalizeOutput(output string) string {
	// Replace any problematic characters (like �) with a space
	return strings.ReplaceAll(output, "�", "")
}

func PrecheckVideo(savedir string, Domain string) (bool, string) {
	TheDownloadURL, err := ReturnDownloadURL(savedir, Domain)
	if err != nil {
		return false, ""
	}
	return true, TheDownloadURL

}
