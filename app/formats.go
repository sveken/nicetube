package main

import (
	"strings"
)

func QualityFinder(path string) string {
	pathParts := strings.Index(path, "/reso/")
	if pathParts != 1 {
		// Further split the result on "/" and return the first segment
		substr := path[pathParts+len("/reso/"):]
		endIndex := strings.Index(substr, "/")
		if endIndex != -1 {
			return substr[:endIndex]
		}
		return substr
	}
	return ""
}
