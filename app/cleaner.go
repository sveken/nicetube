package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

const VideoDir = "./Videos"

//var CleanerMaxAge = time.Duration(maxvideoage) * time.Hour

func cleaner() {
	runcleaner()
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	for {
		<-ticker.C
		runcleaner()
	}
}

func runcleaner() {
	logger.Info("The Cleaner is now running")
	// Iterate over all folders in the root video directory.
	err := filepath.Walk(VideoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			//logger.Error("Error accessing %s: %v", path, err)
			return nil // Continue walking even if there's an error.
		}

		// Check if this is a folder and not the root directory.
		if info.IsDir() && path != VideoDir {
			// Check the folder's last modification time.
			if time.Since(info.ModTime()) > time.Duration(maxvideoage)*time.Hour {
				log.Printf("Deleting folder and its contents: %s", path)
				err := os.RemoveAll(path)
				if err != nil {
					logger.Error("Failed to delete folder %s: %v", path, err)
				}
			}
		}

		return nil
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Error walking the file tree: %v", err))
	}
}
