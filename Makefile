build:
	go build yandex-weather.go

run:
	go run yandex-weather.go

VERSION=$$(git tag | head -1)
build-all-platform:
	@for GOOS in linux darwin windows; \
	do \
		for GOARCH in amd64 386; \
		do \
			echo build: $$GOOS/$$GOARCH; \
			GOOS=$$GOOS GOARCH=$$GOARCH go build; \
			if [ $$GOOS == windows ]; \
			then \
				zip -9 yandex-weather-cli_$(VERSION).$$GOARCH.$$GOOS.zip yandex-weather-cli.exe README.md LICENSE; \
				rm yandex-weather-cli.exe; \
			else \
				zip -9 yandex-weather-cli_$(VERSION).$$GOARCH.$$GOOS.zip yandex-weather-cli README.md LICENSE; \
				rm yandex-weather-cli; \
			fi \
		done \
	done
	GOOS=linux GOARCH=arm go build
	@zip -9 yandex-weather-cli_$(VERSION).arm.linux.zip yandex-weather-cli README.md LICENSE
	@rm yandex-weather-cli

zip-all-platform:
	ls yandex-weather-cli.*.{linux,darwin,exe} | xargs -I@ zip -9 @.zip @

update-from-github:
	go get -u github.com/msoap/yandex-weather-cli

sha1-binary:
	@ls yandex-weather-cli*.{linux,darwin}.zip | xargs -I@ sh -c 'echo "@ $$(unzip -p @ yandex-weather-cli | shasum)"'
	@ls yandex-weather-cli*.windows.zip | xargs -I@ sh -c 'echo "@ $$(unzip -p @ yandex-weather-cli.exe | shasum)"'

sha1-zip:
	shasum yandex-weather-cli*.zip

clean:
	rm yandex-weather-cli*.zip
