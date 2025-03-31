package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// Telling Go what our Map will be, need to actually make it after this.
type MutexMap struct {
	mu      sync.Mutex
	mutexes map[string]*mutexTracker
}

// This counts how many downloads are waiting or using a download mutex, if 0 we can cleanup.
type mutexTracker struct {
	mutex *sync.Mutex
	count int
}

// This holds and keeps track of all the Mutexes we have in play.
func NewMutexMap() *MutexMap {
	return &MutexMap{
		mutexes: make(map[string]*mutexTracker),
	}
}

// This Gets the mutex by its name, or creates one if it does not exist.
// We do this cuz no dynamic variable names so we use a map.
// The name used is a combination of the Quality and VideoID
func (mm *MutexMap) GetMutex(LockKey string) *sync.Mutex {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	tracker, exists := mm.mutexes[LockKey]
	if !exists {
		tracker = &mutexTracker{mutex: &sync.Mutex{}, count: 0}
		mm.mutexes[LockKey] = tracker
		//fmt.Println("Mutex didnt exist so i made it")
	}
	tracker.count++
	//fmt.Println("Increased Mutex")
	return tracker.mutex
}

// Mutex tracker cleanup, As not removing them when not required will slowly increase memory usage.
func (mm *MutexMap) ReleaseMutex(LockKey string) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	if tracker, exists := mm.mutexes[LockKey]; exists {
		tracker.count--
		if tracker.count <= 0 {
			delete(mm.mutexes, LockKey)
			//fmt.Println("Deleted Last Mutex Ref")
		}
	}
}

// Get the Duration of the video
func getVideoDuration(VideoURL string, cookieset string) (time.Duration, error) {
	var args []string
	if cookieset != "" {
		args = append(args, "--cookies", cookieset)
	}
	args = append(args, "--dump-json", "--no-warnings", VideoURL)
	process := exec.Command("./yt-dlp", args...)

	var output bytes.Buffer
	process.Stdout = &output
	process.Stderr = os.Stderr
	err := process.Run()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch metadata: %v", err)
	}
	var metadata struct {
		Duration float64 `json:"duration"` // Duration in seconds
	}
	if err := json.Unmarshal(output.Bytes(), &metadata); err != nil {
		return 0, fmt.Errorf("failed to parse metadata: %v", err)
	}
	return time.Duration(metadata.Duration) * time.Second, nil

}

// The Bot check
func containsBotCheck(output string) bool {
	return strings.Contains(strings.ToLower(output), "sign in to confirm")
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
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".mp4") || strings.HasSuffix(file.Name(), ".webm") || strings.HasSuffix(file.Name(), ".mkv")) {
			Videofile = file.Name()
			break
		}
	}

	// If no .mp4 file is found
	if Videofile == "" {
		return "", fmt.Errorf("no .mp4, .mkv or .webm file found in the directory")
	}

	return Videofile, err
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

func enablecookies() string {
	// Check if the cookie value is over 2 characters and if it is pass the cookies command
	if cookieLocation != "" && utf8.RuneCountInString(cookieLocation) > 2 {
		return cookieLocation
	}
	return ""
}

// GetYTDLPVersion executes yt-dlp --version and returns the version string
func GetYTDLPVersion() string {
	// If we already have the version stored, return it
	if currentYTDLPVersion != "" {
		return currentYTDLPVersion
	}

	// Otherwise fetch it and store it
	cmd := exec.Command("./yt-dlp", "--version")
	output, err := cmd.Output()

	if err != nil {
		fmt.Println("Error getting yt-dlp version:", err)
		currentYTDLPVersion = "Unknown"
		return currentYTDLPVersion
	}

	// Trim any whitespace or newlines from the output
	version := strings.TrimSpace(string(output))

	// Store in global variable
	currentYTDLPVersion = version

	// Print for the console log
	fmt.Printf("Running version %s of yt-dlp\n", version)

	return currentYTDLPVersion
}

// UpdateYTDLP updates yt-dlp to the latest version
func UpdateYTDLP() {
	cmd := exec.Command("./yt-dlp", "-U")
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("Error updating yt-dlp:", err)
		return
	}

	// Trim any whitespace or newlines from the output
	updateOutput := strings.TrimSpace(string(output))

	// Parse the output to provide prettier messages
	if strings.Contains(updateOutput, "yt-dlp is up to date") {
		fmt.Println("yt-dlp is up to date")
		return
	}

	// Extract version information when we are updating so it looks neater in the log thingy.
	var newVersion string
	lines := strings.Split(updateOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Updating to") {
			parts := strings.Split(line, "Updating to ")
			if len(parts) > 1 {
				newVersion = strings.Split(parts[1], " ")[0]
				fmt.Printf("Updating to %s\n", newVersion)
			}
		} else if strings.Contains(line, "Updated yt-dlp to") {
			parts := strings.Split(line, "Updated yt-dlp to ")
			if len(parts) > 1 {
				newVersion = strings.Split(parts[1], " ")[0]
				fmt.Printf("Updated yt-dlp to %s\n", newVersion)

				// Update our cached version
				currentYTDLPVersion = newVersion
			}
		}
	}

	// If we couldn't parse the version from the output, refresh it directly
	if newVersion == "" {
		// Reset the cached version so it will be fetched again
		currentYTDLPVersion = ""
		GetYTDLPVersion()
	}
}

// ytdlpUpdater checks for updates to yt-dlp every 24 hours
func ytdlpUpdater() {
	// Initial update is already done in main, so just set up the timer.
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		<-ticker.C
		fmt.Println("Running scheduled update check for yt-dlp...")
		UpdateYTDLP()
	}
}
