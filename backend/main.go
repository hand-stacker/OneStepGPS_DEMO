package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

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

	mux.HandleFunc("GET /api/google-maps-key/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"key": getGoogleMapsKey()})
	})

	// gets specified user
	mux.HandleFunc("GET /api/user/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		user_id, err := getInt64(r.PathValue("user_id"))
		if err != nil {
			throwBadRequest(w, "invalid user_id")
			return
		}

		user, err := store.GetUser(r.Context(), user_id)
		if errors.Is(err, sql.ErrNoRows) {
			throwStatusNotFound(w, "user not found")
			return
		}
		if err != nil {
			throwInternalServerError(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})

	// creates a new user with only email as input
	// body needs too be in format {'email' :  bob@snailmail.com }
	mux.HandleFunc("POST /api/user/", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var user User
		err := decoder.Decode(&user)
		if err != nil {
			throwBadRequest(w, "bad requjest: "+err.Error())
			return
		}
		err = store.CreateUser(r.Context(), &user)
		if errors.Is(err, ErrEmailTaken) {
			throwConflict(w, err.Error())
			return
		}
		if err != nil {
			throwInternalServerError(w, err.Error())
			return
		}
		userSortPreference := UserSortPreferences{
			UserID:    user.UserID,
			SortOrder: "desc",
		}
		err = store.UpsertSortPreference(r.Context(), &userSortPreference)
		if err != nil {
			throwInternalServerError(w, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(user)
	})

	// all users
	mux.HandleFunc("GET /api/all-users", func(w http.ResponseWriter, r *http.Request) {
		users, err := store.GetAllUser(r.Context())
		if err != nil {
			throwInternalServerError(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	})

	// API to get a user preference by ID
	mux.HandleFunc("GET /api/user-preference/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		user_id, err := getInt64(r.PathValue("user_id"))
		if err != nil {
			throwBadRequest(w, "invalid user_id")
			return
		}
		prefs, err := store.GetSortPreferences(r.Context(), user_id)
		if errors.Is(err, sql.ErrNoRows) {
			throwStatusNotFound(w, "user preference not found")
			return
		}
		if err != nil {
			throwInternalServerError(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prefs)
	})

	// logic for both POST and PUT
	helperSortUpsert := func(w http.ResponseWriter, r *http.Request, retCode int) {
		decoder := json.NewDecoder(r.Body)
		var prefs UserSortPreferences
		err := decoder.Decode(&prefs)
		if err != nil {
			throwBadRequest(w, "bad request: "+err.Error())
			return
		}
		if !validSortType(prefs.SortOrder) {
			throwBadRequest(w, "can only accept asc or desc sort types")
			return
		}
		user_id, err := getInt64(r.PathValue("user_id"))
		if err != nil {
			throwBadRequest(w, "invalid user_id")
			return
		}
		prefs.UserID = &user_id
		err = store.UpsertSortPreference(r.Context(), &prefs)
		if err != nil {
			throwInternalServerError(w, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(retCode)
		json.NewEncoder(w).Encode(prefs)
	}

	//API to make a new user preference, if no prefrence data is sent will populate with default values
	// body needs too be in format {'sort_order' : 'asc' or 'desc'}
	mux.HandleFunc("POST /api/user-preference/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		helperSortUpsert(w, r, 201)
	})

	//API to update a user preference, if no prefrence data is sent will populate with default values
	// body needs too be in format {'sort_order' : 'asc' or 'desc'}
	mux.HandleFunc("PUT /api/user-preference/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		helperSortUpsert(w, r, 204)
	})

	// get hidden devices for a user
	mux.HandleFunc("GET /api/hidden-devices/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		user_id, err := getInt64(r.PathValue("user_id"))
		if err != nil {
			throwBadRequest(w, "invalid user_id")
			return
		}
		ids, err := store.GetHiddenDeviceIDs(r.Context(), user_id)
		if err != nil {
			throwInternalServerError(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		json.NewEncoder(w).Encode(ids)
	})

	helperHiddenDeviceUpdate := func(w http.ResponseWriter, r *http.Request, retCode int) {
		decoder := json.NewDecoder(r.Body)
		var device HiddenDevice
		err := decoder.Decode(&device)
		if err != nil {
			throwBadRequest(w, "bad request: "+err.Error())
			return
		}
		user_id, err := getInt64(r.PathValue("user_id"))
		if err != nil {
			throwBadRequest(w, "invalid user_id")
			return
		}
		device.UserID = &user_id
		err = store.UpsertHiddenDevice(r.Context(), &device)
		if err != nil {
			throwInternalServerError(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(retCode)
		json.NewEncoder(w).Encode(device)
	}

	mux.HandleFunc("POST /api/hidden-devices/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		helperHiddenDeviceUpdate(w, r, 201)
	})

	mux.HandleFunc("PUT /api/hidden-devices/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		helperHiddenDeviceUpdate(w, r, 204)
	})

	// get device nicknames for a user, returns {device_id: display_name}
	mux.HandleFunc("GET /api/device-nicknames/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		user_id, err := getInt64(r.PathValue("user_id"))
		if err != nil {
			throwBadRequest(w, "invalid user_id")
			return
		}
		nicknames, err := store.GetDeviceNicknames(r.Context(), user_id)
		if err != nil {
			throwInternalServerError(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nicknames)
	})

	// logic for both POST and PUT
	// body needs too be in format {'device_id' : 'abc', 'display_name' : 'Truck 1', 'ignore' : false}
	helperDeviceNicknameUpsert := func(w http.ResponseWriter, r *http.Request, retCode int) {
		decoder := json.NewDecoder(r.Body)
		var nickname DeviceNickname
		err := decoder.Decode(&nickname)
		if err != nil {
			throwBadRequest(w, "bad request: "+err.Error())
			return
		}
		if nickname.DeviceID == "" {
			throwBadRequest(w, "device_id is required")
			return
		}
		user_id, err := getInt64(r.PathValue("user_id"))
		if err != nil {
			throwBadRequest(w, "invalid user_id")
			return
		}
		nickname.UserID = &user_id
		err = store.UpsertDeviceNickname(r.Context(), &nickname)
		if err != nil {
			throwInternalServerError(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(retCode)
		json.NewEncoder(w).Encode(nickname)
	}

	mux.HandleFunc("POST /api/device-nicknames/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		helperDeviceNicknameUpsert(w, r, 201)
	})

	mux.HandleFunc("PUT /api/device-nicknames/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		helperDeviceNicknameUpsert(w, r, 200)
	})

	// get list of device_ids where user has a custom marker
	mux.HandleFunc("GET /api/device-markers-list/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		user_id, err := getInt64(r.PathValue("user_id"))
		if err != nil {
			throwBadRequest(w, "invalid user_id")
			return
		}
		ids, err := store.getDeviceWithMarkerIDs(r.Context(), user_id)
		if err != nil {
			throwInternalServerError(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		json.NewEncoder(w).Encode(ids)

	})

	// get the custom marker image for a device, returns the raw image bytes
	mux.HandleFunc("GET /api/device-markers/{user_id}/{device_id}", func(w http.ResponseWriter, r *http.Request) {
		user_id, err := getInt64(r.PathValue("user_id"))
		if err != nil {
			throwBadRequest(w, "invalid user_id")
			return
		}
		marker, err := store.GetDeviceMarker(r.Context(), user_id, r.PathValue("device_id"))
		if errors.Is(err, sql.ErrNoRows) {
			throwStatusNotFound(w, "device marker not found")
			return
		}
		if err != nil {
			throwInternalServerError(w, err.Error())
			return
		}
		writeMarkerImage(w, marker)
	})

	// logic for both POST and PUT
	// body needs to be multipart/form-data with the file in the 'image' field
	helperDeviceMarkerUpsert := func(w http.ResponseWriter, r *http.Request, retCode int) {
		user_id, err := getInt64(r.PathValue("user_id"))
		if err != nil {
			throwBadRequest(w, "invalid user_id")
			return
		}
		data, contentType, err := readMarkerImage(w, r)
		if errors.Is(err, errMarkerTooLarge) {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}
		if err != nil {
			throwBadRequest(w, "bad request: "+err.Error())
			return
		}
		marker := DeviceMarker{
			UserID:      &user_id,
			DeviceID:    r.PathValue("device_id"),
			ContentType: contentType,
			Data:        data,
		}
		err = store.UpsertDeviceMarker(r.Context(), &marker)
		if err != nil {
			throwInternalServerError(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(retCode)
		json.NewEncoder(w).Encode(markerInfo(&marker))
	}

	mux.HandleFunc("POST /api/device-markers/{user_id}/{device_id}", func(w http.ResponseWriter, r *http.Request) {
		helperDeviceMarkerUpsert(w, r, 201)
	})

	mux.HandleFunc("PUT /api/device-markers/{user_id}/{device_id}", func(w http.ResponseWriter, r *http.Request) {
		helperDeviceMarkerUpsert(w, r, 200)
	})

	// removes a custom marker by setting it to ignore, uploading a new image brings it back
	mux.HandleFunc("DELETE /api/device-markers/{user_id}/{device_id}", func(w http.ResponseWriter, r *http.Request) {
		user_id, err := getInt64(r.PathValue("user_id"))
		if err != nil {
			throwBadRequest(w, "invalid user_id")
			return
		}
		err = store.IgnoreDeviceMarker(r.Context(), user_id, r.PathValue("device_id"))
		if errors.Is(err, sql.ErrNoRows) {
			throwStatusNotFound(w, "device marker not found")
			return
		}
		if err != nil {
			throwInternalServerError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	/// HERE ARE EXTERNAL API COLLECTIONS

	// gets device data
	mux.HandleFunc("GET /api/gps-bulk", func(w http.ResponseWriter, r *http.Request) {
		if err := getBulkGPSData(w); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	})

	// gets device info for the user's visible devices (see getDeviceInfo)
	mux.HandleFunc("GET /api/device-info/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		user_id, err := getInt64(r.PathValue("user_id"))
		if err != nil {
			throwBadRequest(w, "invalid user_id")
			return
		}
		if err := getDeviceInfo(w, r, user_id); err != nil {
			throwInternalServerError(w, err.Error())
			return
		}
	})

	// gets device_id and display_name for each of the user's hidden devices (see getHiddenDeviceInfo)
	mux.HandleFunc("GET /api/hidden-device-info/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		user_id, err := getInt64(r.PathValue("user_id"))
		if err != nil {
			throwBadRequest(w, "invalid user_id")
			return
		}
		if err := getHiddenDeviceInfo(w, r, user_id); err != nil {
			throwInternalServerError(w, err.Error())
			return
		}
	})

	addr := "127.0.0.1:" + port
	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, mux))

}
