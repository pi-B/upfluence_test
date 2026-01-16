package services

import (
	"analysis-api/logging"
	"analysis-api/models"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

var logger *slog.Logger

func StartAnalysis(dimension string, c *gin.Context, duration time.Duration) (*models.Watcher, error) {
	logger = logging.Get()
	data_chan := make(chan map[string]models.SocialsData)
	watcher := models.NewWatcher(dimension)

	go watcher.Start(data_chan)

	resp, err := http.Get("https://stream.upfluence.co/stream")
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c, duration) // Remove this from watcher if not used inside it
	defer cancel()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	reader := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return &watcher, nil
		case <-ticker.C:
			readEvent(reader, data_chan)
			continue
		}

	}
}

func readEvent(reader *bufio.Reader, data_chan chan map[string]models.SocialsData) {

	var payload map[string]models.SocialsData
	line, err := reader.ReadBytes('\n')
	if err != nil {
		logger.Error("Error when reading for stream",
			"error", err.Error()) // Add an error counter ?
		return
	}
	if bytes.Equal(line, []byte("\n")) {
		return
	}

	// Parse the received response and extract the data we want to analyse if it's present
	raw_data := retrieveData(line)
	json.Unmarshal(raw_data, &payload)

	data_chan <- payload

}

// Test if the received message respects the expected format and return the content of the data field
func retrieveData(raw []byte) []byte {
	re, err := regexp.Compile(`data: \s*({.*})`)
	if err != nil {
		logger.Error(
			"error while compiling the event parsing regex",
			"error", err.Error(),
		)
		return nil
	}

	ok := re.Match(raw)
	if !ok {
		return nil // Count an error here
	}

	re, _ = regexp.Compile(`({.*})`) // New regex to extract only the part of the event with useful data
	data := re.FindSubmatch(raw)
	return data[0]
}
