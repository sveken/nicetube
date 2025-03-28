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
var cookieLocation string
var downloadCounter int
var botblocked bool
var disableYTDLPUpdate bool

func main() {
	fmt.Println("You are running version 1.0 Beta of NiceTube")
	// Reading any command line flags and adjust the config
	//When we go to docker the start up bach script should do this passing the envoirmetnal variables to the flag
	//Not used yet
	addr := flag.String("addr", ":8085", "HTTP Server address")
	flag.StringVar(&cookieLocation, "cookie", "", "Location of the cookie file")
	flag.IntVar(&maxDuration, "maxDuration", 120, "Max Video Duration in minutes")
	flag.IntVar(&maxvideoage, "max-video-age", 24, "The max age of a video before it is deleted by the cleaner in hours. Set to 0 to disable cleaner")
	enableWebPanel := flag.Bool("web-panel", false, "Enable the web panel interface")
	checkhealth := flag.Bool("checkhealth", false, "Performs a health check, this is used in docker image.")
	flag.BoolVar(&disableYTDLPUpdate, "disable-ytdlp-update", false, "Disable automatic yt-dlp updater")
	flag.Parse()

	//Perform the health check if this is just a health check.
	if *checkhealth {
		healthcheck(*addr)
	}
	// Update yt-dlp to the latest version if it is not already and if updates are not disabled
	if !disableYTDLPUpdate {
		fmt.Println("Checking for updates to yt-dlp. Please wait...")
		UpdateYTDLP()
	} else {
		fmt.Println("yt-dlp auto-update is disabled")
	}
	// Get and display yt-dlp version
	GetYTDLPVersion()

	fmt.Printf("Max Duration of videos is set too %v minutes", maxDuration)
	fmt.Println()
	fmt.Printf("Max Video age has been set to %v hours", maxvideoage)
	//fmt.Printf("Max video age flag is %v", maxvideoage)
	fmt.Println()

	// Enable web panel if flag is set
	if *enableWebPanel {
		SetWebPanelEnabled(true)
		fmt.Println("Web panel interface is enabled at /web")
	}

	//fmt.Printf("Running with options %s", *addr)
	mux := http.NewServeMux()
	//Change this to /Video for container use...actually Just bind the Video folder under this in docker compose
	fileserver := http.FileServer(http.Dir("./Videos/"))

	//URL Time
	mux.HandleFunc("GET /{$}", homepage)
	mux.Handle("GET /Videos/", http.StripPrefix("/Videos", disablelisting(fileserver)))
	mux.HandleFunc("/reso/", GetResoVideos)
	mux.HandleFunc("GET /health", healthservice)

	// Setup web panel handlers if enabled
	if IsWebPanelEnabled() {
		SetupWebHandlers(mux)
	}

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
	// If web panel is enabled, redirect to it
	if IsWebPanelEnabled() {
		http.Redirect(w, r, "/web", http.StatusSeeOther)
		return
	}
	fmt.Fprint(w, "The Web Panel is currently disabled.")
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
