package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
)

func main() {
	port := getenv("PORT", "8080")

	// connect to the ItemStore and ensure the database exists
	store, err := NewItemStore()
	if err != nil {
		return
	}

	if err = store.CreateDB(); err != nil {
		return
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/google-maps-key/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"key": getGoogleMapsKey()})
	})

	// API to get a user preference by ID
	mux.HandleFunc("GET /api/user-preference/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		user_id, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}
		prefs, err := store.GetPreference(r.Context(), user_id)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "user preference not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prefs)
	})

	//API to make a new user preference, if no prefrence data is sent will populate with default values
	// body needs too be in format {'sort_order' : 'asc' or 'desc'}
	mux.HandleFunc("POST /api/user-preference/", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var prefs UserPreferences
		err := decoder.Decode(&prefs)
		if err != nil {
			http.Error(w, "bad request: "+err.Error(), 400)
			return
		}
		err = store.UpsertPreference(r.Context(), &prefs)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(prefs)
	})

	mux.HandleFunc("PUT /api/user-preference/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var prefs UserPreferences
		err := decoder.Decode(&prefs)
		if err != nil {
			http.Error(w, "bad request: "+err.Error(), 400)
			return
		}
		user_id, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}
		prefs.UserID = &user_id
		err = store.UpsertPreference(r.Context(), &prefs)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(204)
		json.NewEncoder(w).Encode(prefs)
	})

	mux.HandleFunc("GET /api/all-preferences", func(w http.ResponseWriter, r *http.Request) {
		prefs, err := store.GetAllPreferences(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prefs)
	})

	mux.HandleFunc("GET /api/gps-bulk", func(w http.ResponseWriter, r *http.Request) {
		if err := getBulkGPSData(w); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	})

	mux.HandleFunc("GET /api/device-info", func(w http.ResponseWriter, r *http.Request) {
		if err := getDeviceInfo(w, r); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	})

	addr := "127.0.0.1:" + port
	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, mux))

}
