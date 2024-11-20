package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

//This actually works and gets everything in avc1/h264 ONLY
//will Start and fallback in the following order
//1080P60/1080P30/720P60/720P30/480p/360p/240p/144p

func GetResoVideos(w http.ResponseWriter, r *http.Request) {
	Domain := r.Host
	//fmt.Println(r.URL.String()) // For troubleshooting
	QualitySelector := QualityFinder(r.URL.Path)
	QualityValue := SetQuality(QualitySelector)
	//fmt.Printf("Hello, you selected %s", QualitySelector)
	//fmt.Printf("Hello, you selected %s", QualityValue)
	forceformat := doWeNeedDashf(QualityValue)
	VideoURL, VideoID, err := urlhelper(r.URL.String(), QualitySelector)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var savedir string
	savedir, err = foldergen(VideoID, savedir, QualitySelector)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	outputname := fmt.Sprintf("%s/%%(title)s.%%(ext)s", savedir)
	process := exec.Command(
		"./yt-dlp",
		forceformat, QualityValue,
		"--remux-video", "mp4", "--restrict-filenames",
		"--ffmpeg-location", "./",
		"-o", outputname,
		VideoURL,
	)

	process.Stdin = os.Stdin
	process.Stdout = os.Stdout
	process.Stderr = os.Stderr

	if err := process.Run(); err != nil {
		fmt.Printf("Command failed with exit code %d\n", process.ProcessState.ExitCode())
		fmt.Println(err)
	}
	TheDownloadURL, err := ReturnDownloadURL(savedir, Domain)
	//fmt.Println(TheDownloadURL)
	//http.Redirect(w, r, TheDownloadURL, http.StatusSeeOther)
	if err != nil {
		fmt.Fprintf(w, "error: No file was downloaded. Is the URL correct?")
	}
	fmt.Fprintf(w, TheDownloadURL)
}
