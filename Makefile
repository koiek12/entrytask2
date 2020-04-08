BACKEND_NAME=backend
WEB_NAME=web

BASE=$(shell pwd)
BIN=$(BASE)/bin

install:
	go mod download

build:
	go build -o $(BIN)/backend ./cmd/server/main.go || exit
	go build -o $(BIN)/web ./cmd/web/main.go || exit