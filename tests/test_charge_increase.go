package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {

	charge := []int{10, 100, 500}
	duration := "40s"
	endpoint := "http://localhost:8080/analysis"
	parameters := fmt.Sprintf("?dimension=likes&duration=%v", duration)

	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Head(endpoint)
	if err != nil {
		slog.Error("Error when calling API", "error", err)
	}
	if resp.StatusCode >= 500 {
		slog.Error("API is not reachable")
		os.Exit(0)
	} else {
		slog.Info("API is reachable !")
	}
	resp.Body.Close()

	for _, c := range charge {
		slog.Info(fmt.Sprintf("Starting test for charge %v", c))
		var wg sync.WaitGroup
		var mu sync.Mutex
		error_counter := make([]map[string]interface{}, 0)

		for i := range c {
			wg.Add(1)
			go func(wg *sync.WaitGroup, mu *sync.Mutex, i int) {
				defer wg.Done()
				mu.Lock()
				client := http.Client{}
				mu.Unlock()
				resp, err := client.Get(endpoint + parameters)
				if err != nil {
					mu.Lock()
					error_counter = append(error_counter, map[string]interface{}{"GET_ERROR": err.Error()})
					mu.Unlock()
					return
				}

				if resp.StatusCode != 200 {
					slog.Error("New error !")
					body, _ := io.ReadAll(resp.Body)
					mu.Lock()
					error_counter = append(error_counter, map[string]interface{}{"RESP_ERROR": map[string]interface{}{"CODE": resp.StatusCode, "BODY": string(body)}})
					mu.Unlock()
				}
				resp.Body.Close()

			}(&wg, &mu, i)
		}
		wg.Wait()

		slog.Info(fmt.Sprintf("Results for charge : %v\n", c))
		slog.Info(fmt.Sprintf("Number of errors : %v\nError average : %v", len(error_counter), (len(error_counter) / c)))
		if len(error_counter) > 0 {
			for _, err := range error_counter {
				slog.Info("Error report")
				slog.Info(fmt.Sprintf("%+v", err))
			}
		}
	}
}
