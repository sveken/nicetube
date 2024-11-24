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
	QPoggers         = ""
	Q1080Ph264Forced = "299+ba/137+ba/216+ba/298+ba/136+ba/135+ba/134+ba/133+ba/160+ba"
	Q720Ph264Forced  = "298+ba/136+ba/135+ba/134+ba/133+ba/160+ba"
	Q480Ph264Forced  = "135+ba/134+ba/133+ba/160+ba"
	Q1080P           = "303+ba/248+ba/699+ba/399+ba/302+ba/247+ba/244+ba/243+ba/242+ba/278+ba"
	Q720P            = "302+ba/247+ba/244+ba/243+ba/242+ba/278+ba"
	Q480P            = "244+ba/243+ba/242+ba/278+ba"
	Q1440P           = "308+ba/271+ba/700+ba/400+ba/303+ba/248+ba/699+ba/399+ba/302+ba/247+ba/244+ba/243+ba/242+ba/278+ba"
	Q2160P           = "315+ba/313+ba/701+ba/401+ba/308+ba/271+ba/700+ba/400+ba/303+ba/248+ba/699+ba/399+ba/302+ba/247+ba/244+ba/243+ba/242+ba/278+ba"
)

func SetQuality(QualitySelector string) string {
	switch QualitySelector {
	case "Q1080Ph264Forced":
		return Q1080Ph264Forced
	case "Q720Ph264Forced":
		return Q720Ph264Forced
	case "Q480Ph264Forced":
		return Q480Ph264Forced
	case "Q1080P":
		return Q1080P
	case "Q720P":
		return Q720P
	case "Q480P":
		return Q480P
	case "Q1440P":
		return Q1440P
	case "Q2160P":
		return Q2160P
	case "QPoggers":
		return QPoggers
	default:
		return Q720P
	}
}

func doWeNeedDashf(QualityValue string) string {
	if QualityValue != "" {
		return "-f"
	}
	return ""

}
