package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-session/session"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MyHandlers struct {
	db *gorm.DB
}

func NewMyHandlers(db *gorm.DB) *MyHandlers {
	return &MyHandlers{db: db}
}

// ClientFormHandler get client data from form
func ClientFormHandler(r *http.Request) (string, string, error) {
	clientID := r.Form.Get("client_id")
	if clientID == "" {
		return "", "", errors.ErrInvalidClient
	}
	clientSecret := r.Form.Get("client_secret")
	return clientID, clientSecret, nil
}

func (h *MyHandlers) UserAuthorizeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) (userID string, err error) {

	if dumpvar {
		err = dumpRequest(os.Stdout, "userAuthorizeHandler", r)
		if err != nil {
			log.Println("dumpErr", err)
		}
	}
	store, err := session.Start(ctx, w, r)
	if err != nil {
		logrus.Errorf("Failed Session start err:", err)
		return userID, err
	}

	uid, ok := store.Get(store.Get())
	if !ok {
		if r.Form == nil {
			r.ParseForm()
		}

		store.Set("ReturnUri", r.Form)
		store.Save()

		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	userID = uid.(string)
	store.Delete("LoggedInUserID")
	store.Save()
	return
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if dumpvar {
		_ = dumpRequest(os.Stdout, "login", r) // Ignore the error
	}
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		logrus.Errorf("Failed Session.Start %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		if r.Form == nil {
			if err := r.ParseForm(); err != nil {
				logrus.Errorf("Parse from err %s", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		store.Set("LoggedInUserID", r.Form.Get("username"))
		store.Save()

		w.Header().Set("Location", "/auth")
		w.WriteHeader(http.StatusFound)
		return
	}
	outputHTML(w, r, "static/login.html")
}

func outputHTML(w http.ResponseWriter, req *http.Request, filename string) {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		logrus.Errorf("Failed to parse template: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type pageData struct {
		Roles []string
	}
	data := pageData{
		Roles: AllRoles,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logrus.Errorf("Failed to execute template: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	if dumpvar {
		_ = dumpRequest(os.Stdout, "auth", r) // Ignore the error
	}
	store, err := session.Start(nil, w, r)
	if err != nil {
		logrus.Errorf("failed Session.Start  %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, ok := store.Get("LoggedInUserID"); !ok {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	outputHTML(w, r, "static/auth.html")
}
