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
	//Not used yet
	addr := flag.String("addr", ":8085", "HTTP Server address")
	flag.Parse()
	fmt.Printf("Running with options %s", *addr)
	mux := http.NewServeMux()
	//Change this to /Video for container use...actually Just bind the Video folder under this in docker compose
	fileserver := http.FileServer(http.Dir("./Videos/"))

	//URL Time
	mux.HandleFunc("GET /{$}", homepage)
	mux.Handle("GET /Videos/", http.StripPrefix("/Videos", disablelisting(fileserver)))
	mux.HandleFunc("/reso/", GetResoVideos)

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

func urlhelper(path, QualitySelector string) (string, string, error) {
	QualityName := fmt.Sprintf("%s/", QualitySelector)
	pathParts := strings.Split(path, QualityName)
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
func foldergen(VideoID string, savedir string) (string, error) {
	savedir = fmt.Sprintf("./Videos/%s", VideoID)
	err := os.MkdirAll(savedir, 0755) // Permissions: rwxr-xr-x
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return savedir, err
	}
	return savedir, err
}

func ReturnDownloadURL(savedir string, Domain string) (string, error) {
	//fmt.Printf("Save directory is %s and the URL from Path is %s", savedir, Domain)
	mp4File, err := GetFileName(savedir)
	if err != nil {
		return "", fmt.Errorf("no .mp4 file found in the directory")
	}
	//Remove first dot from save directory to make URL from
	URLFriendlyDirIndex := strings.Index(savedir, ".")
	URLFriendlyDir := savedir[:URLFriendlyDirIndex] + savedir[URLFriendlyDirIndex+1:]
	//fmt.Printf("And the Mp4 name is %s", mp4File)
	//fmt.Println()
	TheDownloadURL := fmt.Sprintf("http://%s%s/%s", Domain, URLFriendlyDir, mp4File)
	return TheDownloadURL, err
}

func GetFileName(savedir string) (string, error) {
	files, err := os.ReadDir(savedir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %v", err)
	}

	// Find the .mp4 file in the directory and extract its name
	var mp4File string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".mp4") {
			mp4File = file.Name()
			break
		}
	}

	// If no .mp4 file is found
	if mp4File == "" {
		return "", fmt.Errorf("no .mp4 file found in the directory")
	}

	return mp4File, err
}
