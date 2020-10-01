package web

import (
	"net/http"

	db "../db"

	log "github.com/Sirupsen/logrus"
)

// https://gist.github.com/mschoebel/9398202

func getUserID(r *http.Request) (userID string) {
	if cookie, err := r.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userID = cookieValue["userID"]
		}
	}
	return userID
}

func setSession(w http.ResponseWriter, userID string) {
	value := map[string]string{
		"userID": userID,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
	}
}

func clearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

func loginInternal(w http.ResponseWriter, r *http.Request) {
	log.Info("Received internal login request from " + r.RemoteAddr)
	r.ParseForm()
	userID := r.FormValue("userID")
	password := r.FormValue("password")
	if userID == "admin" && password == Config.AdminPassword {
		log.Info("Successful internal login from " + r.RemoteAddr)
		setSession(w, userID)
		// Create a cart if none is present
		db.CreateCart(userID)
		http.Redirect(w, r, "/shop", 302)
	} else {
		log.Error("Failed internal login from " + r.RemoteAddr)
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
	http.Redirect(w, r, "/login", 302)
}

func securePageHandler(w http.ResponseWriter, r *http.Request) bool {
	userID := getUserID(r)
	if userID == "" {
		log.Info("Client with IP ", r.RemoteAddr, " tried to access ", r.URL.Path, " without active session")
		http.Redirect(w, r, "/login", 302)
		return false
	}
	return true
}
