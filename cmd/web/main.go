package main

import (
	"net/http"

	"git.garena.com/youngiek.song/entry_task/internal/cache"
	"git.garena.com/youngiek.song/entry_task/internal/controller"
	"git.garena.com/youngiek.song/entry_task/internal/logger"
	"git.garena.com/youngiek.song/entry_task/pkg/message"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func initRoute(userController *controller.UserController, docRoot string) {
	r := mux.NewRouter()
	r.HandleFunc("/login", userController.Login)
	r.HandleFunc("/main", userController.GetUserInfo)
	r.HandleFunc("/users/{id}", userController.EditUserInfo)
	r.HandleFunc("/users/{id}/profile/picture", userController.UploadPhoto)
	r.HandleFunc("/", userController.LoginPage)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(docRoot+"/static/"))))
	http.Handle("/", r)
}

func main() {
	logLevel := "info"
	logPath := "web.log"
	backendHost, backendPort := "localhost", "3233"
	backendMaxConn := 100
	cacheHost, cachePort := "localhost", "6379"
	docRoot := "../../web/"
	listenHost, listenPort := "localhost", "8080"

	logger.Init(logPath, logLevel)
	client := message.NewClient(backendHost, backendPort, backendMaxConn)
	cache := cache.NewUserCache(cacheHost, cachePort)
	userController := controller.NewUserController(client, cache, logger.Instance, docRoot)
	initRoute(userController, docRoot)

	logger.Instance.Info("Web Server has started, Listening on port " + listenPort + "...")
	err := http.ListenAndServe(listenHost+":"+listenPort, nil)
	if err != nil {
		logger.Instance.Fatal("http server error", zap.String("error", err.Error()))
	}
}
