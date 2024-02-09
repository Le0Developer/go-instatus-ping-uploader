package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus-community/pro-bing"
)

func main() {
	pageId, ok := os.LookupEnv("PAGE_ID")
	if !ok {
		log.Fatal("PAGE_ID not set")
	}
	pingMetricId, ok := os.LookupEnv("PING_METRIC_ID")
	if !ok {
		log.Fatal("PING_METRIC_ID not set")
	}
	lossMetricId, ok := os.LookupEnv("LOSS_METRIC_ID")
	if !ok {
		log.Fatal("LOSS_METRIC_ID not set")
	}
	apiToken, ok := os.LookupEnv("API_TOKEN")
	if !ok {
		log.Fatal("API_TOKEN not set")
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", apiToken),
	}

	statistics := make(chan probing.Statistics)
	go takeStatistics(statistics)

	pingApiUrl := fmt.Sprintf("https://api.instatus.com/v1/%s/metrics/%s/data", pageId, pingMetricId)
	lossApiUrl := fmt.Sprintf("https://api.instatus.com/v1/%s/metrics/%s/data", pageId, lossMetricId)

	pingPoints := []DataPoint{}
	lossPoints := []DataPoint{}
	for measurement := range statistics {
		now := time.Now().UnixMilli()
		fmt.Println("Measurement", now, measurement)
		pingDataPoint := DataPoint{Timestamp: now, Value: float64(measurement.MaxRtt)}
		lossDataPoint := DataPoint{Timestamp: now, Value: float64(measurement.PacketLoss) * 100}

		if lossDataPoint.Value == 0 {
			// Instatus doesn't support zero values for some reason
			lossDataPoint.Value = 0.0001
		}

		pingPoints = append(pingPoints, pingDataPoint)
		lossPoints = append(lossPoints, lossDataPoint)

		if len(pingPoints) >= 5 {
			err := postMetric(pingApiUrl, headers, pingPoints)
			if err == nil {
				pingPoints = []DataPoint{}
			} else {
				log.Print(err)
			}
		}
		if len(lossPoints) >= 5 {
			err := postMetric(lossApiUrl, headers, lossPoints)
			if err == nil {
				lossPoints = []DataPoint{}
			} else {
				log.Print(err)
			}
		}
	}
}

type DataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

func postMetric(url string, headers map[string]string, data []DataPoint) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	for key, value := range headers {
		request.Header.Add(key, value)
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	fmt.Println("Posted", url, resp.StatusCode)
	if resp.StatusCode >= 500 || resp.StatusCode == 429 {
		return errors.New(fmt.Sprintf("Server error: %d", resp.StatusCode))
	}
	return nil
}

func takeStatistics(results chan probing.Statistics) {
	for {
		pinger, err := probing.NewPinger("1.1.1.1")
		if err != nil {
			log.Print(err)
			continue
		}

		pinger.Count = 60
		err = pinger.Run()
		if err != nil {
			log.Print(err)
			continue
		}

		stats := pinger.Statistics()
		if stats != nil {
			results <- *stats
		}
	}
}
