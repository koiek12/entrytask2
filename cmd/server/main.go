package main

import (
	"database/sql"
	"os"

	"git.garena.com/youngiek.song/entry_task/internal/logger"
	"git.garena.com/youngiek.song/entry_task/pkg/message"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

const (
	ConnHost = "localhost"
	ConnPort = "3233"
	ConnType = "tcp"
)

func initLogger() {
	logger.Init("server.log")
}

func initDB() {
	var err error
	db, err = sql.Open("mysql", "song:abcd@/entry_task")
	if err != nil {
		logger.Instance.Fatal("Cannot connect to DB " + err.Error())
		os.Exit(1)
	}
	db.SetMaxIdleConns(20)
	err = db.Ping()
	if err != nil {
		logger.Instance.Fatal("Cannot connect to DB " + err.Error())
		logger.Instance.Fatal("Cannot connect to DB")
		os.Exit(1)
	}
}

func main() {
	initLogger()
	initDB()

	server := message.NewServer("localhost", "3233")
	server.Run()
}
