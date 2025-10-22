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
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	authenticated, ok := sess.Values["authenticated"].(bool)
	if !ok || !authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, ok := sess.Values["userID"].(int64)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username, ok := sess.Values["username"].(string)
	if !ok {
		username = ""
	}

	_, displayName, err := h.DB.GetUserByUsername(username)
	if err != nil {
		displayName = "User"
	}

	bookings, err := h.DB.GetUserBookings(userID)
	if err != nil {
		log.Printf("Error getting bookings: %v", err)
		bookings = []map[string]interface{}{}
	}

	currentTime := time.Now().Format("2006-01-02 15:04:05")
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
