package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"git.garena.com/youngiek.song/entry_task/internal/jwt"
	"git.garena.com/youngiek.song/entry_task/internal/logger"
	"git.garena.com/youngiek.song/entry_task/pkg/message"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

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
	logLevel := "info"
	logPath := "server.log"
	dbHost, dbPort := "localhost", "3306"
	dbId, dbPasswd := "song", "abcd"
	dbMaxConn := 100
	listenHost, listenPort := "localhost", "3233"
	jwtSecretKey := "young"
	var tokenExpireTime time.Duration = 30

	logger.Init(logPath, logLevel)
	// initialize database connection
	db := initDB(dbHost, dbPort, dbId, dbPasswd, dbMaxConn)
	tokenIssuer := jwt.NewTokenIssuer(jwtSecretKey, tokenExpireTime)
	server := message.NewServer(listenHost, listenPort, db, tokenIssuer, logger.Instance)
	server.Run()
}
