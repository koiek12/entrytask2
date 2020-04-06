package main

import (
	"net/http"

	"git.garena.com/youngiek.song/entry_task/internal/handlers"
	"git.garena.com/youngiek.song/entry_task/internal/logger"
	"github.com/gorilla/mux"
)

func initLogger() {
	logger.Init("web.log")
}

func initRouter() {
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.LoginPage)
	r.HandleFunc("/login", handlers.Login)
	r.HandleFunc("/main", handlers.GetUserInfo)
	r.HandleFunc("/users/{id}", handlers.EditUserInfo)
	r.HandleFunc("/users/{id}/profile/picture", handlers.UploadPhoto)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("../../web/static/"))))
	http.Handle("/", r)
}

func main() {
	initLogger()
	initRouter()
	logger.Instance.Info("Web Server has started, Listening on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Instance.Fatal(err.Error())
	}
}
