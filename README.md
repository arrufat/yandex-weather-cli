Command line interface for Yandex weather service
=================================================

[![Build Status](https://travis-ci.org/msoap/yandex-weather-cli.svg?branch=master)](https://travis-ci.org/msoap/yandex-weather-cli)
[![Coverage Status](https://coveralls.io/repos/github/msoap/yandex-weather-cli/badge.svg?branch=master)](https://coveralls.io/github/msoap/yandex-weather-cli?branch=master)
[![Homebrew formula exists](https://img.shields.io/badge/homebrew-🍺-d7af72.svg)](https://github.com/msoap/yandex-weather-cli#install)
[![Snap Status](https://build.snapcraft.io/badge/msoap/yandex-weather-cli.svg)](https://snapcraft.io/yandex-weather-cli)
[![Report Card](https://goreportcard.com/badge/github.com/msoap/yandex-weather-cli)](https://goreportcard.com/report/github.com/msoap/yandex-weather-cli)

Install
-------

MacOS via homebrew:

    brew tap msoap/tools
    brew install yandex-weather-cli
    # update:
    brew upgrade yandex-weather-cli

Or download binaries from: [releases](https://github.com/msoap/yandex-weather-cli/releases) (OS X/Linux/Windows/RaspberryPi)

Or build from source:

    go get -u github.com/msoap/yandex-weather-cli
    ln -s $GOPATH/bin/yandex-weather-cli ~/bin/

Or use snap (Ubuntu or any Linux distribution with snap)

    # install stable version:
    sudo snap install yandex-weather-cli
    
    # install the latest version:
    sudo snap install --edge yandex-weather-cli
    
    # update
    sudo snap refresh yandex-weather-cli

Usage
-----

    # weather client by default use your current location
    yandex-weather-cli [options] [city]

    # options
        -json     : JSON out
        -no-color : no coloring
        -no-today : skip today forecast
        -version  : get version
        -help

    # in another city
    yandex-weather-cli kyiv
    yandex-weather-cli london

    # JSON out
    yandex-weather-cli -json london

### Environment variables

For setup own yandex.pogoda URL, you may set variables:

  * `Y_WEATHER_URL`
  * `Y_WEATHER_MINI_URL`

Screenshot
----------
<img src="https://raw.githubusercontent.com/msoap/yandex-weather-cli/misc/img/yandex-weather.go.2018-08-05.0.screenshot.png" align="center" alt="Screenshot" height="576" width="682">

See also
--------

  * [pogoda.yandex.ru](https://pogoda.yandex.ru/)
  * [github.com/schachmat/wego](https://github.com/schachmat/wego) - another weather command line client (Go)
  * [github.com/jfrazelle/weather](https://github.com/jfrazelle/weather) - another weather command line client (Go)
  * [github.com/sramsay/wu](https://github.com/sramsay/wu) - another weather command line client (Go)
  * [github.com/brianriley/weather-cli](https://github.com/brianriley/weather-cli) - another weather command line client (Python)
  * [github.com/JackWink/Weather](https://github.com/JackWink/Weather) - another weather command line client (Python)
