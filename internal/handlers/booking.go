package handlers

import (
	"door-control/internal/db"
	"encoding/json"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
)

type BookingHandler struct {
	DB        *db.DB
	Store     *sessions.CookieStore
	Templates *template.Template
}

const (
	maxDistanceKm = 0.05
)

func getStudioCoordinates() (float64, float64) {
	lat := 0.0
	lon := 0.0

	if latStr := os.Getenv("STUDIO_LATITUDE"); latStr != "" {
		if parsed, err := strconv.ParseFloat(latStr, 64); err == nil {
			lat = parsed
		}
	}

	if lonStr := os.Getenv("STUDIO_LONGITUDE"); lonStr != "" {
		if parsed, err := strconv.ParseFloat(lonStr, 64); err == nil {
			lon = parsed
		}
	}

	return lat, lon
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	sess, err := h.Store.Get(r, "webauthn-session")
	if err != nil || sess.Values["authenticated"] != true {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, ok := sess.Values["userID"].(int64)
	if !ok {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}

	var requestData struct {
		StartTime int64 `json:"start_time"`
		EndTime   int64 `json:"end_time"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	if requestData.StartTime == 0 || requestData.EndTime == 0 {
		http.Error(w, "Start time and end time required", http.StatusBadRequest)
		return
	}

	conflict, err := h.DB.CheckBookingConflict(userID, requestData.StartTime, requestData.EndTime)
	if err != nil {
		log.Printf("Error checking booking conflict: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if conflict {
		http.Error(w, "Booking conflict - you already have a booking during this time", http.StatusConflict)
		return
	}

	bookingID, err := h.DB.CreateBooking(userID, requestData.StartTime, requestData.EndTime, time.Now().Unix())
	if err != nil {
		log.Printf("Error creating booking: %v", err)
		http.Error(w, "Failed to create booking", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"booking_id": bookingID,
	})
}

func (h *BookingHandler) GetUserBookings(w http.ResponseWriter, r *http.Request) {
	sess, err := h.Store.Get(r, "webauthn-session")
	if err != nil || sess.Values["authenticated"] != true {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, ok := sess.Values["userID"].(int64)
	if !ok {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}

	bookings, err := h.DB.GetUserBookings(userID)
	if err != nil {
		log.Printf("Error getting bookings: %v", err)
		http.Error(w, "Failed to get bookings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}

func (h *BookingHandler) UnlockDoor(w http.ResponseWriter, r *http.Request) {
	sess, err := h.Store.Get(r, "webauthn-session")
	if err != nil || sess.Values["authenticated"] != true {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, ok := sess.Values["userID"].(int64)
	if !ok {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}

	var requestData struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	currentTime := time.Now().Unix()
	booking, err := h.DB.GetActiveBooking(userID, currentTime)
	if err != nil {
		log.Printf("No active booking: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "No active booking found. Please book a time slot first.",
		})
		return
	}

	studioLat, studioLon := getStudioCoordinates()
	distance := haversine(requestData.Latitude, requestData.Longitude, studioLat, studioLon)

	if distance > maxDistanceKm {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":        "error",
			"message":       "Please go to the front door for the door to open.",
			"distance":      distance,
			"show_navigate": true,
			"studio_lat":    studioLat,
			"studio_lon":    studioLon,
		})
		return
	}

	log.Printf("Door unlocked for user %d, booking %v", userID, booking["id"])

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Door unlocked! Welcome to Waterhouse Studios.",
	})
}

func (h *BookingHandler) BookingPage(w http.ResponseWriter, r *http.Request) {
	sess, err := h.Store.Get(r, "webauthn-session")
	if err != nil || sess.Values["authenticated"] != true {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	h.Templates.ExecuteTemplate(w, "booking.html", nil)
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0

	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
