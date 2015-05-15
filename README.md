Command line interface for Yandex weather service
=================================================

Install
-------------------

Download binaries from: [releases](https://github.com/msoap/yandex-weather-cli/releases) (OS X/Linux/Windows/RaspberryPi)

From source:

    go get -u github.com/msoap/yandex-weather-cli
    ln -s $GOPATH/bin/yandex-weather-cli ~/bin/

Usage
-----

    # weather in current location
    yandex-weather-cli

    # options
    yandex-weather-cli -help
    yandex-weather-cli -no-color

    # in another city
    yandex-weather-cli kiev
    yandex-weather-cli london

    # JSON out
    yandex-weather-cli -json london

Screenshot
----------
<img src="https://raw.githubusercontent.com/msoap/yandex-weather-cli/misc/img/yandex-weather.go.2015-03-28.0.screenshot.png" align="center" alt="Screenshot" height="439" width="604">

See also
--------

  * [pogoda.yandex.ru](https://pogoda.yandex.ru/)
  * [github.com/schachmat/wego](https://github.com/schachmat/wego) - another weather command line client (Go)
  * [github.com/jfrazelle/weather](https://github.com/jfrazelle/weather) - another weather command line client (Go)
  * [github.com/sramsay/wu](https://github.com/sramsay/wu) - another weather command line client (Go)
  * [github.com/brianriley/weather-cli](https://github.com/brianriley/weather-cli) - another weather command line client (Python)
  * [github.com/JackWink/Weather](https://github.com/JackWink/Weather) - another weather command line client (Python)
