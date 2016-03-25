/*

Command line interface for Yandex weather service (https://pogoda.yandex.ru/)

usage:
	go build yandex-weather.go

	./yandex-weather
	./yandex-weather -no-color
	./yandex-weather kiev

	# JSON out
	./yandex-weather -json london

https://github.com/msoap/yandex-weather-cli

*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/msoap/html2data"
)

// Config - application config
type Config struct {
	city    string
	getJSON bool
	noColor bool
	noToday bool
}

// HourTemp - one hour temperature
type HourTemp struct {
	Hour int    `json:"hour"`
	Temp int    `json:"temp"`
	Icon string `json:"icon"`
}

const (
	// BaseURL - yandex pogoda service url (testing: "http://localhost:8080/get?url=https://pogoda.yandex.ru/")
	BaseURL = "https://pogoda.yandex.ru/"
	// BaseURLMini - url for forecast by hours (testing: "http://localhost:8080/get?url=https://p.ya.ru/")
	BaseURLMini = "https://p.ya.ru/"
	// UserAgent - for http.request
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11) AppleWebKit/601.1.56 (KHTML, like Gecko) Version/9.0 Safari/601.1.56"
	// ForecastDays - parse days in forecast
	ForecastDays = 10
)

// Selectors - css selectors for forecast today
var Selectors = map[string]string{
	"city":                "div.navigation-city h1",
	"term_now":            "div.current-weather div.current-weather__thermometer_type_now",
	"term_another_name1":  "span.current-weather__col:nth-child(3) span.current-weather__thermometer-name",
	"term_another_name2":  "span.current-weather__col:nth-child(4) span.current-weather__thermometer-name",
	"term_another_value1": "span.current-weather__col:nth-child(3) div.current-weather__thermometer",
	"term_another_value2": "span.current-weather__col:nth-child(4) div.current-weather__thermometer",
	"desc_now":            "div.current-weather span.current-weather__comment",
	"wind":                "div.current-weather div.current-weather__info-row:nth-child(2) span.wind-speed",
	"humidity":            "div.current-weather div.current-weather__info-row:nth-child(3)",
	"pressure":            "div.current-weather div.current-weather__info-row:nth-child(4)",
}

// SelectorsNextDays - css selectors for forecast next days
var SelectorsNextDays = map[string]string{
	"date":       "div.tabs-panes span.forecast-brief__item-day",
	"desc":       "div.tabs-panes div.forecast-brief__item-comment",
	"term":       "div.tabs-panes div.forecast-brief__item-temp-day",
	"term_night": "div.tabs-panes div.forecast-brief__item-temp-night",
}

// SelectorByHoursRoot - Root element for forecast data
var SelectorByHoursRoot = "div.temp-chart__wrap"

// SelectorByHours - get forecast by hours
var SelectorByHours = map[string]string{
	"hour": "p.temp-chart__hour",
	"temp": "div.temp-chart__temp",
	"icon": "i.icon:attr(class)",
}

// ICONS - unicode symbols for icon names
var ICONS = map[string]string{
	"icon_snow": "✻",
	"icon_rain": "☂",
}

//-----------------------------------------------------------------------------
// get command line parameters
func getParams() (cfg Config) {
	flag.BoolVar(&cfg.getJSON, "json", false, "get JSON")
	flag.BoolVar(&cfg.noColor, "no-color", false, "disable colored output")
	flag.BoolVar(&cfg.noToday, "no-today", false, "disable today forecast")
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] [city]\noptions:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Printf("\nexamples:\n  %s kiev\n  %s -json london\n", os.Args[0], os.Args[0])
	}
	flag.Parse()

	cfg.city = ""
	if flag.NArg() >= 1 {
		cfg.city = flag.Args()[0]
	}

	if runtime.GOOS == "windows" {
		// broken unicode symbols in cmd.exe and dont detect pipe
		cfg.noToday = true
	} else {
		// detect pipe
		stdoutStat, _ := os.Stdout.Stat()
		if (stdoutStat.Mode() & os.ModeCharDevice) == 0 {
			cfg.noColor = true
		}
	}

	return cfg
}

//-----------------------------------------------------------------------------
// get weather html page as http.Response
func getWeatherPage(weatherURL string) *http.Response {
	cookie, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookie,
	}

	request, err := http.NewRequest("GET", weatherURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	request.Header.Set("User-Agent", UserAgent)

	// create request for set cookies only
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	response, err = client.Get(weatherURL)

	if err != nil {
		log.Fatal(err)
	}

	return response
}

//-----------------------------------------------------------------------------
// parse html via goquery, find DOM-nodes with weather forecast data
func getWeather(cfg Config) (map[string]interface{}, []HourTemp, []map[string]interface{}) {
	doc := html2data.FromURL(BaseURL+cfg.city, html2data.URLCfg{UA: UserAgent})

	// now block
	forecastNow := map[string]interface{}{}
	data, err := doc.GetDataFirst(Selectors)
	if err != nil {
		log.Fatal(err)
	}

	reRemoveDesc := regexp.MustCompile(`^.+\s*:\s*`)
	for name := range Selectors {
		forecastNow[name] = clearNonprintInString(data[name])
		switch name {
		case "humidity", "pressure", "wind":
			forecastNow[name] = reRemoveDesc.ReplaceAllString(forecastNow[name].(string), "")
		case "term_now", "term_another_value1", "term_another_value2":
			forecastNow[name] = convertStrToInt(forecastNow[name].(string))
		}
		if name == "wind" && forecastNow[name] == nil {
			forecastNow[name] = "0 м/с"
		}
	}

	// forecast for next days block
	forecastNext := make([]map[string]interface{}, 0, ForecastDays)
	dataNextDays, err := doc.GetData(SelectorsNextDays)
	if err != nil {
		log.Fatal(err)
	}

	if dateColumn, ok := dataNextDays["date"]; ok {
		for i := range dateColumn {
			forecastNext = append(forecastNext, map[string]interface{}{})

			for name := range SelectorsNextDays {
				text := ""
				if _, ok := dataNextDays[name]; ok && len(dataNextDays[name]) >= i+1 {
					text = dataNextDays[name][i]
				}
				forecastNext[i][name] = clearNonprintInString(text)
			}
		}
	}

	// suggest dates
	for i := range forecastNext {
		forecastNext[i]["date"], forecastNext[i]["json_date"] = suggestDate(time.Now(), forecastNext[i]["date"].(string), i)
		forecastNext[i]["term"] = convertStrToInt(forecastNext[i]["term"].(string))
		forecastNext[i]["term_night"] = convertStrToInt(forecastNext[i]["term_night"].(string))
	}

	// forecast by hours block
	var forecastByHours []HourTemp
	if !cfg.noToday {
		docMini := html2data.FromURL(BaseURLMini+cfg.city, html2data.URLCfg{UA: UserAgent})
		dataHours, err := docMini.GetDataNested(SelectorByHoursRoot, SelectorByHours)
		if err != nil {
			log.Fatal(err)
		}

		for _, row := range dataHours {
			hour := convertStrToInt(row["hour"][0])
			temp := convertStrToInt(row["temp"][0])
			forecastByHours = append(forecastByHours, HourTemp{Hour: hour, Temp: temp, Icon: parseIcon(row["icon"][0])})
		}
	}

	return forecastNow, forecastByHours, forecastNext
}

//-----------------------------------------------------------------------------
// get icon name from css class attribut
func parseIcon(cssClass string) string {
	allAttributes := regexp.MustCompile(`\s+`).Split(cssClass, -1)
	for _, attr := range allAttributes {
		if _, ok := ICONS[attr]; ok {
			return attr
		}
	}
	return ""
}

//-----------------------------------------------------------------------------
// render data as text or JSON
func render(forecastNow map[string]interface{}, forecastByHours []HourTemp, forecastNext []map[string]interface{}, cfg Config) {
	if _, ok := forecastNow["city"]; ok {
		outWriter := getColorWriter(cfg.noColor)

		if cfg.getJSON {

			if !cfg.noToday && len(forecastByHours) > 0 {
				forecastNow["by_hours"] = forecastByHours
			}

			if len(forecastNext) > 0 {
				for _, row := range forecastNext {
					row["date"] = row["json_date"]
					delete(row, "json_date")
				}
				forecastNow["next_days"] = forecastNext
			}

			json, _ := json.Marshal(forecastNow)
			fmt.Println(string(json))

		} else {

			fmt.Fprintf(outWriter, cfg.ansiColourString("%s (<yellow>%s</>)\n"), forecastNow["city"], BaseURL+cfg.city)
			fmt.Fprintf(outWriter,
				cfg.ansiColourString("Сейчас: <green>%d °C</>, <green>%s</>, %s: <green>%d °C</>, %s: <green>%d °C</>\n"),
				forecastNow["term_now"],
				forecastNow["desc_now"],
				forecastNow["term_another_name1"],
				forecastNow["term_another_value1"],
				forecastNow["term_another_name2"],
				forecastNow["term_another_value2"],
			)
			fmt.Fprintf(outWriter, cfg.ansiColourString("Давление: <green>%s</>\n"), forecastNow["pressure"])
			fmt.Fprintf(outWriter, cfg.ansiColourString("Влажность: <green>%s</>\n"), forecastNow["humidity"])
			fmt.Fprintf(outWriter, cfg.ansiColourString("Ветер: <green>%s</>\n"), forecastNow["wind"])

			if !cfg.noToday && len(forecastByHours) > 0 {
				textByHour := [4]string{}
				for _, item := range forecastByHours {
					textByHour[0] += fmt.Sprintf("%3d ", item.Hour)
					textByHour[2] += fmt.Sprintf("%3d°", item.Temp)
					icon, exists := ICONS[item.Icon]
					if !exists {
						icon = " "
					}
					textByHour[3] += fmt.Sprintf(cfg.ansiColourString("<blue>%3s</blue> "), icon)
				}
				textByHour[1] = cfg.ansiColourString("<grey+h>" + renderHisto(forecastByHours) + "</>")

				fmt.Fprintf(outWriter, strings.Repeat("─", len(forecastByHours)*4)+"\n")
				fmt.Fprintf(outWriter, "%s\n%s\n%s\n%s\n",
					cfg.ansiColourString("<grey+h>"+textByHour[0]+"</>"),
					textByHour[1],
					textByHour[2],
					textByHour[3],
				)
			}

			if len(forecastNext) > 0 {
				descLength := getMaxLengthInSlice(forecastNext, "desc")
				fmt.Fprintf(outWriter, "%s\n", strings.Repeat("─", 27+descLength))
				fmt.Fprintf(outWriter,
					cfg.ansiColourString("<blue+h> %-10s %4s %-*s %8s</>\n"),
					"дата",
					"°C",
					descLength, "погода",
					"°C ночью",
				)
				fmt.Fprintf(outWriter, "%s\n", strings.Repeat("─", 27+descLength))

				weekendRe := regexp.MustCompile(`(сб|вс)`)
				for _, row := range forecastNext {
					date := weekendRe.ReplaceAllString(row["date"].(string), cfg.ansiColourString("<red+h>$1</>"))
					fmt.Fprintf(outWriter,
						" %10s %3d° %-*s %7d°\n",
						date,
						row["term"].(int),
						descLength,
						row["desc"],
						row["term_night"].(int),
					)
				}
			}
		}

	} else {
		fmt.Printf("City \"%s\" dont found\n", cfg.city)
	}
}

//-----------------------------------------------------------------------------
func main() {
	cfg := getParams()
	forecastNow, forecastByHours, forecastNext := getWeather(cfg)
	render(forecastNow, forecastByHours, forecastNext, cfg)
}
