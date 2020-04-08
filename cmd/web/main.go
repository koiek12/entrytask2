package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"git.garena.com/youngiek.song/entry_task/internal/cache"
	"git.garena.com/youngiek.song/entry_task/internal/controller"
	"git.garena.com/youngiek.song/entry_task/internal/logger"
	"git.garena.com/youngiek.song/entry_task/pkg/message"
	"gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
	"github.com/sevlyar/go-daemon"
	"go.uber.org/zap"
)

var configPath string

func initCmdLineFlag() {
	flag.StringVar(&configPath, "config", "./configs/web.yaml", "configuration file")
	flag.StringVar(&configPath, "c", "./configs/web.yaml", "configuration file")
	flag.Parse()
}

type config struct {
	HTTP struct {
		Host    string `yaml:"host"`
		Port    string `yaml:"port"`
		DocRoot string `yaml:"document_root"`
	} `yaml:"http"`
	TCP struct {
		Host    string `yaml:"host"`
		Port    string `yaml:"port"`
		MaxConn int    `yaml:"max_connection"`
	} `yaml:"tcp"`
	Redis struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"redis"`
	Log struct {
		Level string `yaml:"level"`
		Path  string `yaml:"path"`
	} `yaml:"log"`
}

func getConfig(path string) (*config, error) {
	var cfg config
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewDecoder(f)
	decoder.Decode(&cfg)
	return &cfg, nil
}

func initRoute(userController *controller.UserController, docRoot string) {
	r := mux.NewRouter()
	r.HandleFunc("/login", userController.Login)
	r.HandleFunc("/main", userController.Main)
	r.HandleFunc("/users/{id}", userController.EditUserInfo)
	r.HandleFunc("/users/{id}/profile/picture", userController.UploadPhoto)
	r.HandleFunc("/", userController.LoginPage)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(docRoot+"/static/"))))
	http.Handle("/", r)
}

func main() {
	//daemonize
	cntxt := &daemon.Context{
		PidFileName: "web.pid",
		PidFilePerm: 0644,
		Umask:       027,
	}
	d, err := cntxt.Reborn()
	if err != nil {
		fmt.Println("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	initCmdLineFlag()
	cfg, err := getConfig(configPath)
	if err != nil {
		fmt.Println("Cannot open config file ", err.Error)
		return
	}
	logger.Init(cfg.Log.Path, cfg.Log.Level)
	client := message.NewClient(cfg.TCP.Host, cfg.TCP.Port, cfg.TCP.MaxConn)
	cache := cache.NewUserCache(cfg.Redis.Host, cfg.Redis.Port)
	userController := controller.NewUserController(client, cache, logger.Instance, cfg.HTTP.DocRoot)
	initRoute(userController, cfg.HTTP.DocRoot)

	logger.Instance.Info("Web Server has started, Listening on port " + cfg.HTTP.Port + "...")
	err = http.ListenAndServe(cfg.HTTP.Host+":"+cfg.HTTP.Port, nil)
	if err != nil {
		logger.Instance.Fatal("http server error", zap.String("error", err.Error()))
	}
}
