Command line interface for Yandex weather service
=================================================

Install
-------------------

MacOS via homebrew:

    brew tap msoap/tools
    brew install yandex-weather-cli
    # update:
    brew update; brew upgrade yandex-weather-cli

Or download binaries from: [releases](https://github.com/msoap/yandex-weather-cli/releases) (OS X/Linux/Windows/RaspberryPi)

Or build from source:

    go get -u github.com/msoap/yandex-weather-cli
    ln -s $GOPATH/bin/yandex-weather-cli ~/bin/

Usage
-----

    # weather client by default use your current location
    yandex-weather-cli [options] [city]

    # options
        -json     : JSON out
        -no-color : no coloring
        -no-today : skip today forecast
        -help

    # in another city
    yandex-weather-cli kiev
    yandex-weather-cli london

    # JSON out
    yandex-weather-cli -json london

Screenshot
----------
<img src="https://raw.githubusercontent.com/msoap/yandex-weather-cli/misc/img/yandex-weather.go.2015-12-28.0.screenshot.png" align="center" alt="Screenshot" height="539" width="764">

See also
--------

  * [pogoda.yandex.ru](https://pogoda.yandex.ru/)
  * [github.com/schachmat/wego](https://github.com/schachmat/wego) - another weather command line client (Go)
  * [github.com/jfrazelle/weather](https://github.com/jfrazelle/weather) - another weather command line client (Go)
  * [github.com/sramsay/wu](https://github.com/sramsay/wu) - another weather command line client (Go)
  * [github.com/brianriley/weather-cli](https://github.com/brianriley/weather-cli) - another weather command line client (Python)
  * [github.com/JackWink/Weather](https://github.com/JackWink/Weather) - another weather command line client (Python)
