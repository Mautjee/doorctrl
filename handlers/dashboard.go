package handlers

import (
	"door-control/db"
	"html/template"
	"net/http"

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

	_, displayName, err := h.DB.GetUserByUsername("")
	if err != nil {
		displayName = "User"
	}

	data := map[string]interface{}{
		"UserID":      userID,
		"DisplayName": displayName,
	}

	h.Templates.ExecuteTemplate(w, "dashboard.html", data)
}
