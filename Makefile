.PHONY: transmission-telegram test deps

transmission-telegram: format deps
	go build

deps:
	go get -t

format:
	go fmt ./...

test: format
	go test ./...

check: test

docker:
	docker build -t docker.pkg.github.com/zhulik/transmission-telegram/transmission-telegram:latest .
