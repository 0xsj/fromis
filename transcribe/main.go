package main

import (
	"fmt"
)

func main() {
	yt := &YouTubeTranscriptAPI{}

	// https://www.youtube.com/watch?v=BY81yNttfpg

	data := yt.Get("BY81yNttfpg")
	for videoId, transcript := range data {
		fmt.Printf("Transcript for %s:\n", videoId)
		for _, entry := range transcript {
			fmt.Printf("Start: %f, Duration: %f, Text: %s\n", entry.Start, entry.Duration, entry.Text)
		}
	}
}
