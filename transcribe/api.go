package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var logger *log.Logger

func init() {
	logFile, err := os.Create("logs.txt")
	if err != nil {
		fmt.Println("Error creating log file:", err)
		os.Exit(1)
	}
	logger = log.New(logFile, "", log.LstdFlags)
}

type YouTubeTranscriptAPI struct{}

type TranscriptEntry struct {
	Text     string  `xml:",chardata"`
	Start    float64 `xml:"start,attr"`
	Duration float64 `xml:"dur,attr"`
}

type TranscriptParser struct {
	PlainData string
}

func (api *YouTubeTranscriptAPI) Get(videoIds ...string) map[string][]TranscriptEntry {
	data := make(map[string][]TranscriptEntry)

	for _, videoId := range videoIds {
		logger.Printf("Fetching transcript for video ID: %s\n", videoId)
		transcript, err := fetchTranscript(videoId)
		if err != nil {
			logger.Printf("Could not get the transcript for the video %s: %v\n", videoId, err)
			continue
		}

		logger.Printf("Transcript fetched successfully for video ID: %s\n", videoId)
		entries := parseTranscript(transcript)
		data[videoId] = entries

		fmt.Println(entries)

		// Transform entries to JSON
		jsonData, _ := json.Marshal(entries)
		fmt.Printf("JSON for video ID %s:\n%s\n", videoId, string(jsonData))

		// Transform entries to dictionary
		dict := make(map[float64]TranscriptEntry)
		for _, entry := range entries {
			dict[entry.Start] = entry
		}
		fmt.Printf("Dictionary for video ID %s:\n%+v\n", videoId, dict)
	}

	return data
}

func fetchTranscript(videoId string) (string, error) {
	logger.Printf("Fetching YouTube page for video ID: %s\n", videoId)
	resp, err := http.Get(fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoId))
	if err != nil {
		return "", fmt.Errorf("error fetching video: %v", err)
	}
	defer resp.Body.Close()

	logger.Printf("Reading response body for video ID: %s\n", videoId)
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	fetchedSite := string(body)

	re := regexp.MustCompile(`"captions":\{"playerCaptionsTracklistRenderer":\{"captionTracks":\[{"baseUrl":"(.*?)"`)

	match := re.FindStringSubmatch(fetchedSite)
	if len(match) < 2 {
		return "", fmt.Errorf("timedtext URL not found")
	}

	apiURL := strings.ReplaceAll(match[1], `\u0026`, "&")
	apiURL = strings.ReplaceAll(apiURL, `\\`, "")

	logger.Printf("Fetching transcript from API URL: %s\n", apiURL)
	transcriptResponse, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("error fetching transcript: %v", err)
	}
	defer transcriptResponse.Body.Close()

	logger.Printf("Reading transcript body for video ID: %s\n", videoId)
	transcriptBody, err := io.ReadAll(transcriptResponse.Body)
	if err != nil {
		return "", fmt.Errorf("error reading transcript body: %v", err)
	}

	return string(transcriptBody), nil
}

func parseTranscript(plainData string) []TranscriptEntry {
	logger.Printf("Parsing transcript data")
	// logger.Printf(plainData)
	var parser TranscriptParser
	parser.PlainData = plainData
	return parser.parse()
}

func (parser *TranscriptParser) parse() []TranscriptEntry {
	logger.Printf("Unmarshalling XML data")
	var entries []TranscriptEntry
	err := xml.Unmarshal([]byte(parser.PlainData), &entries)
	if err != nil {
		logger.Println("Error parsing transcript:", err)
		return nil
	}

	return entries
}
