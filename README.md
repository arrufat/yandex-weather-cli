Command line interface for Yandex weather service
=================================================

Usage:
------

    # install
    go get github.com/msoap/yandex-weather-cli
    ln -s $GOPATH/bin/yandex-weather-cli ~/bin/yandex-weather

    # weather in current location
    yandex-weather

    # options
    yandex-weather -help
    yandex-weather -no-color

    # in another city
    yandex-weather kiev
    yandex-weather london

    # JSON out
    yandex-weather -json london

Screenshot:
-----------
<img src="https://raw.githubusercontent.com/msoap/msoap.github.com/master/img/yandex-weather.go.2014-12-10.0.screenshot.png" align="center" alt="Screenshot" height="439" width="604">

Links:
------

  * [pogoda.yandex.ru](https://pogoda.yandex.ru/)
  * [github.com/jfrazelle/weather](https://github.com/jfrazelle/weather) - another weather command line client (Go)
  * [github.com/sramsay/wu](https://github.com/sramsay/wu) - another weather command line client (Go)
  * [github.com/brianriley/weather-cli](https://github.com/brianriley/weather-cli) - another weather command line client (Python)
  * [github.com/JackWink/Weather](https://github.com/JackWink/Weather) - another weather command line client (Python)
