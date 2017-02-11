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

gometalinter:
	gometalinter --vendor --cyclo-over=25 --line-length=150 --dupl-threshold=150 --min-occurrences=3 --enable=misspell --deadline=10m

generate-manpage:
	docker run -it --rm -v $$PWD:/app -w /app ruby-ronn sh -c 'cat README.md | grep -v "^\[" | grep -v Screenshot > yandex-weather-cli.md; ronn yandex-weather-cli.md; mv ./yandex-weather-cli ./yandex-weather-cli.1; rm ./yandex-weather-cli.html ./yandex-weather-cli.md'
