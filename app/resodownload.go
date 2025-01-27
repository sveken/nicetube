package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

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

	//Lock out the Quality and VideoID combination until it errors or downloads to stop multiple download jobs for the same video running over the top of each other.
	LockKey := fmt.Sprintf("%s_%s", VideoID, QualitySelector)
	mutex := mm.GetMutex(LockKey)
	mutex.Lock()
	defer mm.ReleaseMutex(LockKey)
	defer mutex.Unlock()

	//Precheck if file is downloaded
	var VideoExists bool
	VideoExists, TheDownloadURL := PrecheckVideo(savedir, Domain)
	if VideoExists {
		fmt.Fprint(w, TheDownloadURL)
		return
	}
	// Enable Cookie paramaters if the cookie location is set
	cookieset := enablecookies()

	//Set Duration Limit for the funny people.
	duration, err := getVideoDuration(VideoURL, cookieset)
	if err != nil {
		fmt.Printf("Error fetching video duration: %v\n", err)
	}
	if duration > time.Duration(maxDuration)*time.Minute {
		logger.Error("Error: Video over the set max Duration")
		fmt.Fprintf(w, "error: Video over the set max Duration of %v minutes of this server", maxDuration)
		return
	}

	outputname := fmt.Sprintf("%s/%%(title)s.%%(ext)s", savedir)

	var args []string

	if cookieset != "" {
		args = append(args, "--cookies", cookieset)
	}

	args = append(args,
		forceformat, QualityValue,
		"--restrict-filenames", "--replace-in-metadata", "title", "%", "_",
		"--ffmpeg-location", "./",
		"-o", outputname, "--",
		VideoURL,
	)

	//To test what command is getting passed to ytdlp
	//fmt.Println(args)

	process := exec.Command("./yt-dlp", args...)

	//removed the following to try fix "--remux-video", "mp4",
	// This pipes the output into the buffer for error checking and also the terminal while i build the program.
	process.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	process.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	//process.Stderr = os.Stderr

	// Use these when outputting to terminal can be reduced ie production
	//process.Stdout = &stdoutBuf
	//process.Stderr = &stderrBuf
	err = process.Run()

	stdout := normalizeOutput(stdoutBuf.String())
	stderr := normalizeOutput(stderrBuf.String())

	//debugging
	//fmt.Println("Captured stdout:", stdout)
	//fmt.Println("Captured stderr:", stderr)

	if err != nil {
		fmt.Printf("Command failed with exit code %d\n", process.ProcessState.ExitCode())
		fmt.Println(err)
	}

	// Check if Youtube has blocked the server.
	// and notify the user so they can hopefully let the host know to fix it.
	botblocked := false
	if containsBotCheck(stdout) || containsBotCheck(stderr) {
		//fmt.Println("Error: Bot confirmation required. Please sign in to continue.")
		logger.Error("Error: Bot confirmation required. Please sign in to continue. This IP may be blacklisted by Youtube")
		fmt.Fprintf(w, "error: YouTube has blocked this IP or server. Please swap to another or notify the host.")
		botblocked = true

	}

	//For some reason, sometimes if a video takes awhile to merge the logic continues anyway and pulls a part number as the URL
	//This gives a 1 second delay so the invalid part files can be deleted.
	time.Sleep(1 * time.Second)

	TheDownloadURL, err = ReturnDownloadURL(savedir, Domain)
	//fmt.Println(TheDownloadURL)
	//http.Redirect(w, r, TheDownloadURL, http.StatusSeeOther)
	if err != nil && botblocked != true {
		fmt.Fprintf(w, "error: No file was downloaded. Is the URL correct?")
	}
	fmt.Fprint(w, TheDownloadURL)
}
