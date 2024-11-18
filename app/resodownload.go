package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func urlhelper(path, prefix string) (string, error) {
	pathParts := strings.Split(path, prefix)
	if len(pathParts[1]) < 3 {
		return "", errors.New("missing or invalid video URL")
	}
	// fixing the https:/ thingy with https:// and getting the final video URL
	VideoURL := strings.Replace(pathParts[1], "https:/", "https://", 1)
	//Generate a unique ID to track the video based off string at the end of the video URL
	lastSlashPosition := strings.LastIndex(pathParts[1], "/")
	VideoID := pathParts[1][lastSlashPosition+1:]
	logger.Info("Getting Youtube video", "URL", VideoURL, "VideoID", VideoID)
	return VideoURL, nil

}

//This actually works and gets everything in avc1/h264 ONLY
//will Start and fallback in the following order
//1080P60/1080P30/720P60/720P30/480p/360p/240p/144p

func reso1080Ph264forced(w http.ResponseWriter, r *http.Request) {
	VideoURL, err := urlhelper(r.URL.Path, "reso1080Ph264forced/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	process := exec.Command(
		"./yt-dlp",
		"-f", "299+ba/137+ba/216+ba/298+ba/136+ba/135+ba/134+ba/133+ba/160+ba",
		"--remux-video", "mp4",
		"-o", "./Videos/%(title)s.%(ext)s",
		VideoURL,
	)

	process.Stdin = os.Stdin
	process.Stdout = os.Stdout
	process.Stderr = os.Stderr

	if err := process.Run(); err != nil {
		fmt.Printf("Command failed with exit code %d\n", process.ProcessState.ExitCode())
		fmt.Println(err)
	}
}
