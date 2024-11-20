package main

import (
	"fmt"
	"net/http"
	"os"
)

//Not sure what to try yet, but currently it just serves a page that says i am running.
// Now it just checks if the video folder exists which it should in the docker container

func healthcheck(w http.ResponseWriter, r *http.Request) {
	dir, err := os.Open("./Videos")
	if err != nil {
		fmt.Fprint(w, "Failed to read Video directory")
	} else {
		fmt.Fprint(w, "Check passed")
	}
	defer dir.Close()

}
