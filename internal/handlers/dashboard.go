package handlers

import (
	"door-control/internal/db"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

type DashboardHandler struct {
	DB        *db.DB
	Store     *sessions.CookieStore
	Templates *template.Template
}

func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	sess, err := h.Store.Get(r, "webauthn-session")
	if err != nil {
		log.Printf("Dashboard access denied: session error from IP: %s - %v", r.RemoteAddr, err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	authenticated, ok := sess.Values["authenticated"].(bool)
	if !ok || !authenticated {
		log.Printf("Dashboard access denied: not authenticated from IP: %s", r.RemoteAddr)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, ok := sess.Values["userID"].(int64)
	if !ok {
		log.Printf("Dashboard access denied: invalid user ID from IP: %s", r.RemoteAddr)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	log.Printf("Dashboard accessed by user ID: %d from IP: %s", userID, r.RemoteAddr)

	_, displayName, err := h.DB.GetUserByID(userID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		displayName = "User"
	}

	bookings, err := h.DB.GetUserBookings(userID)
	if err != nil {
		log.Printf("Error getting bookings: %v", err)
		bookings = []map[string]interface{}{}
	}

	currentTime := time.Now().Unix()
	activeBooking, err := h.DB.GetActiveBooking(userID, currentTime)
	hasActiveBooking := err == nil && activeBooking != nil

	data := map[string]interface{}{
		"UserID":           userID,
		"DisplayName":      displayName,
		"Bookings":         bookings,
		"HasActiveBooking": hasActiveBooking,
		"ActiveBooking":    activeBooking,
	}

	h.Templates.ExecuteTemplate(w, "dashboard.html", data)
}
