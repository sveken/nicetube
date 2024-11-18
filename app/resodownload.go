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

func reso1080Ph264forced(w http.ResponseWriter, r *http.Request) {
	Domain := r.Host
	VideoURL, VideoID, err := urlhelper(r.URL.Path, "reso1080Ph264forced/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var savedir string
	savedir, err = foldergen(VideoID, savedir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	outputname := fmt.Sprintf("./Videos/%s/%%(title)s.%%(ext)s", VideoID)
	process := exec.Command(
		"./yt-dlp",
		"-f", "299+ba/137+ba/216+ba/298+ba/136+ba/135+ba/134+ba/133+ba/160+ba",
		"--remux-video", "mp4",
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
	ReturnDownloadURL(savedir, Domain)
}
