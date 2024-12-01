package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func getVideoDuration(VideoURL string) (time.Duration, error) {
	process := exec.Command(
		"./yt-dlp",
		"--dump-json",
		"--no-warnings", // Suppress warnings
		VideoURL,
	)

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
