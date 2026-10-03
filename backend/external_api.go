package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Source - https://stackoverflow.com/a/31129967 (modified a little to authenticate)
// gets json...
var myClient = &http.Client{}
var API_KEY = getenv("API_KEY", "DEMO")

func getJson(w http.ResponseWriter, url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+API_KEY)
	req.Header.Set("Content-Type", "application/json")
	r, err := myClient.Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d", r.StatusCode)
	}
	_, err = w.Write(body)
	return err
}

func getBulkGPSData(w http.ResponseWriter) error {
	(w).Header().Set("Content-Type", "application/json")
	return getJson(w, "https://track.onestepgps.com/v3/api/public/device")

}
