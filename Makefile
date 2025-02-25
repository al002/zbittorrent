.PHONY: build run validate clean config

build:
	go build -o bin/zbittorrent ./main.go

run: build
	./bin/zbittorrent --config configs/config.yaml

validate: build
	./bin/zbittorrent validate --config configs/config.yaml

info: build
	./bin/zbittorrent info --config configs/config.yaml $(P)

clean:
	rm -r bin/

config:
	mkdir -p configs
	cp configs/config.example.yaml configs/config.yaml
