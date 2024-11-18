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

var (
	Q1080Ph264forced = "299+ba/137+ba/216+ba/298+ba/136+ba/135+ba/134+ba/133+ba/160+ba"
)

func SetQuality(QualitySelector string) string {
	switch QualitySelector {
	case "Q1080Ph264forced":
		return Q1080Ph264forced
	default:
		return ""
	}
}

func doWeNeedDashf(QualityValue string) string {
	if QualityValue != "" {
		return "-f"
	}
	return ""

}
