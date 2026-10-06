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

func getGoogleMapsKey() string {
	return getenv("GOOGLE_MAPS_API_KEY", "DEMO")
}

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

// device info for every device the user hasn't hidden
func getDeviceInfo(w http.ResponseWriter, r *http.Request, user_id int64) error {
	(w).Header().Set("Content-Type", "application/json")
	// so long
	url := "https://track.onestepgps.com/v3/api/public/device-info?lat_lng=True&device_id=True&license_plate=True&drive_status=True&drive_status_begin_time=True&drive_status_distance_mi=True&fuel_percent=True&honored_gps_time=True&speed_mph=True&ev_charge_percent=True"
	store, err := NewItemStore()
	if err != nil {
		return err
	}

	blockedDevices, err := store.GetHiddenDeviceIDs(r.Context(), user_id)
	if err != nil {
		return err
	}
	url = url + "&device_id_not_match=" + makeMatchOrRegex(blockedDevices)
	return getJson(w, url)
}

// device_id and display_name for each device the user has hidden
func getHiddenDeviceInfo(w http.ResponseWriter, r *http.Request, user_id int64) error {
	(w).Header().Set("Content-Type", "application/json")
	// so short
	url := "https://track.onestepgps.com/v3/api/public/device-info?device_id=True"
	store, err := NewItemStore()
	if err != nil {
		return err
	}

	blockedDevices, err := store.GetHiddenDeviceIDs(r.Context(), user_id)
	if err != nil {
		return err
	}
	// an empty match pattern could match every device, so skip the call when nothing is hidden
	if len(blockedDevices) == 0 {
		_, err = w.Write([]byte("[]"))
		return err
	}
	url = url + "&device_id_match=" + makeMatchOrRegex(blockedDevices)
	return getJson(w, url)
}
