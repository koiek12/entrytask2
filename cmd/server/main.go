package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"git.garena.com/youngiek.song/entry_task/internal/jwt"
	"git.garena.com/youngiek.song/entry_task/internal/logger"
	"git.garena.com/youngiek.song/entry_task/pkg/message"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sevlyar/go-daemon"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var configPath string

func initCmdLineFlag() {
	flag.StringVar(&configPath, "config", "./configs/server.yaml", "configuration file")
	flag.StringVar(&configPath, "c", "./configs/server.yaml", "configuration file")
	flag.Parse()
}

type config struct {
	Tcp struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"tcp"`
	Database struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		MaxConn  int    `yaml:"max_connection"`
	} `yaml:"database"`
	JWT struct {
		SecretKey  string `yaml:"secret"`
		ExpireTime int64  `yaml:"expire"`
	} `yaml:"jwt"`
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

func initDB(host, port, id, passwd string, maxDBConn int) *sql.DB {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/entry_task", id, passwd, host, port))
	if err != nil {
		logger.Instance.Fatal("Cannot connect to DB", zap.String("error", err.Error()))
		os.Exit(1)
	}
	// set below options to maintain and limit number of connection for assigned time
	db.SetMaxOpenConns(maxDBConn)
	db.SetMaxIdleConns(maxDBConn)
	db.SetConnMaxLifetime(time.Second * 60)
	// check database alive
	err = db.Ping()
	if err != nil {
		logger.Instance.Fatal("Cannot connect to DB", zap.String("error", err.Error()))
		os.Exit(1)
	}
	return db
}

func main() {
	//daemonize
	cntxt := &daemon.Context{
		PidFileName: "backend.pid",
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
	conf, err := getConfig(configPath)
	if err != nil {
		fmt.Println("Cannot open config file ", err.Error)
		return
	}
	logger.Init(conf.Log.Path, conf.Log.Level)
	// initialize database connection
	db := initDB(conf.Database.Host, conf.Database.Port, conf.Database.User, conf.Database.Password, conf.Database.MaxConn)
	tokenIssuer := jwt.NewTokenIssuer(conf.JWT.SecretKey, time.Minute*time.Duration(conf.JWT.ExpireTime))
	server := message.NewServer(conf.Tcp.Host, conf.Tcp.Port, db, tokenIssuer, logger.Instance)
	server.Run()
}
