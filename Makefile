.PHONY: build run

build:
	go build -o bin/zbittorrent ./main.go

run: build
	./bin/zbittorrent --config configs/config.yaml
