build:
	go build

run:
	go run yandex-weather.go terminal_unix.go util.go

update-from-github:
	go get -u github.com/msoap/yandex-weather-cli

test:
	go test -v -cover -race ./...

lint:
	golint ./...
	go vet ./...
	errcheck ./...
