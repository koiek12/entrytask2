package main

import (
	"database/sql"

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

func main() {
	initLogger()

	server := message.NewServer("localhost", "3233")
	server.Run()
}
