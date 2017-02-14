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

create-debian-amd64-package:
	GOOS=linux GOARCH=amd64 go build -ldflags="-w" -o yandex-weather-cli
	set -e ;\
	TAG_NAME=$$(git tag 2>/dev/null | grep -E '^[0-9]+' | tail -1) ;\
	docker run -it --rm -v $$PWD:/app -w /app -e TAG_NAME=$$TAG_NAME ruby-fpm sh -c 'fpm -s dir -t deb --name yandex-weather-cli -v $$TAG_NAME ./yandex-weather-cli=/usr/bin/ ./yandex-weather-cli.1=/usr/share/man/man1/ LICENSE=/usr/share/doc/yandex-weather-cli/copyright README.md=/usr/share/doc/yandex-weather-cli/'
	rm yandex-weather-cli
