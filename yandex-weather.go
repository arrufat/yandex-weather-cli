/*

Command line interface for Yandex weather service (https://pogoda.yandex.ru/)

usage:
	go build yandex-weather.go

	./yandex-weather
	./yandex-weather -no-color
	./yandex-weather kyiv

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
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/msoap/html2data"
)

// Config - application config
type Config struct {
	baseURL     string
	baseURLMini string
	city        string
	getJSON     bool
	noColor     bool
	noToday     bool
}

// HourTemp - one hour temperature
type HourTemp struct {
	Hour int    `json:"hour"`
	Temp int    `json:"temp"`
	Icon string `json:"icon"`
}

const (
	// EnvBaseURLName - environment variable for setup base URL
	EnvBaseURLName = "Y_WEATHER_URL"
	// EnvBaseURLMiniName - environment variable for setup base URL (for days forecast)
	EnvBaseURLMiniName = "Y_WEATHER_MINI_URL"
	// BaseURLDefault - yandex pogoda service url (testing: "http://localhost:8080/get?url=https://yandex.ru/pogoda/")
	BaseURLDefault = "https://yandex.ru/pogoda/"
	// BaseURLMiniDefault - url for forecast by hours (testing: "http://localhost:8080/get?url=https://p.ya.ru/")
	BaseURLMiniDefault = "https://p.ya.ru/"
	// UserAgent - for http.request
	UserAgent = "yandex-weather-cli/1.10"
	// ForecastDays - parse days in forecast
	ForecastDays = 10
	// TodayForecastTableWidth - today forecast table width for align tables
	TodayForecastTableWidth = 14*4 - 27
)

// Selectors - css selectors for forecast today
var Selectors = map[string]string{
	"city":                "div.location h1.title",
	"term_now":            "div.fact div.fact__temp",
	"term_another_name1":  "div.content__brief a.link:nth-child(1) div.day-parts-next__name",
	"term_another_value1": "div.content__brief a.link:nth-child(1) div.day-parts-next__value",
	"term_another_name2":  "div.content__brief a.link:nth-child(2) div.day-parts-next__name",
	"term_another_value2": "div.content__brief a.link:nth-child(2) div.day-parts-next__value",
	"term_another_name3":  "div.content__brief a.link:nth-child(3) div.day-parts-next__name",
	"term_another_value3": "div.content__brief a.link:nth-child(3) div.day-parts-next__value",
	"term_another_name4":  "div.content__brief a.link:nth-child(4) div.day-parts-next__name",
	"term_another_value4": "div.content__brief a.link:nth-child(4) div.day-parts-next__value",
	"desc_now":            "div.fact div.fact__condition",
	"wind":                "div.fact div.fact__props dl.fact__wind-speed dd.term__value",
	"humidity":            "div.fact div.fact__props dl.fact__humidity dd.term__value",
	"pressure":            "div.fact div.fact__props dl.fact__pressure dd.term__value",
}

// SelectorsNextDays - css selectors for forecast next days
var SelectorsNextDays = map[string]string{
	"date":       "div.forecast-briefly__days time.time",
	"desc":       "div.forecast-briefly__days div.forecast-briefly__condition",
	"term":       "div.forecast-briefly__days div.forecast-briefly__temp_day span.temp__value",
	"term_night": "div.forecast-briefly__days div.forecast-briefly__temp_night span.temp__value",
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
		fmt.Printf("\nexamples:\n  %s kyiv\n  %s -json london\n", os.Args[0], os.Args[0])
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
		if stdoutStat, err := os.Stdout.Stat(); err != nil || (stdoutStat.Mode()&os.ModeCharDevice) == 0 {
			cfg.noColor = true
		}
	}

	if baseURL := os.Getenv(EnvBaseURLName); len(baseURL) > 0 {
		cfg.baseURL = baseURL
	} else {
		cfg.baseURL = BaseURLDefault
	}
	if baseURLMini := os.Getenv(EnvBaseURLMiniName); len(baseURLMini) > 0 {
		cfg.baseURLMini = baseURLMini
	} else {
		cfg.baseURLMini = BaseURLMiniDefault
	}

	return cfg
}

//-----------------------------------------------------------------------------
// parse html via goquery, find DOM-nodes with weather forecast data
func getWeather(cfg Config) (map[string]interface{}, []HourTemp, []map[string]interface{}) {
	doc := html2data.FromURL(cfg.baseURL+cfg.city, html2data.URLCfg{UA: UserAgent})

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
		case "term_now", "term_another_value1", "term_another_value2", "term_another_value3", "term_another_value4":
			if value, ok := forecastNow[name]; ok {
				forecastNow[name] = convertStrToInt(value.(string))
			}
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
		for i, dateStr := range dateColumn {
			if dateStr == "" {
				continue
			}

			forecastNext = append(forecastNext, map[string]interface{}{})

			for name := range SelectorsNextDays {
				text := ""
				if _, ok := dataNextDays[name]; ok && len(dataNextDays[name]) >= i+1 {
					text = dataNextDays[name][i]
				}
				forecastNext[i][name] = clearNonprintInString(text)

				if value, ok := forecastNext[i][name].(string); ok && name == "desc" {
					forecastNext[i][name] = strings.ToLower(value)
				}
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
		docMini := html2data.FromURL(cfg.baseURLMini+cfg.city, html2data.URLCfg{UA: UserAgent})
		dataHours, err := docMini.GetDataNestedFirst(SelectorByHoursRoot, SelectorByHours)
		if err == nil {
			for _, row := range dataHours {
				hour := convertStrToInt(row["hour"])
				temp := convertStrToInt(row["temp"])
				forecastByHours = append(forecastByHours, HourTemp{Hour: hour, Temp: temp, Icon: parseIcon(row["icon"])})
			}
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
	if cityFromPage, ok := forecastNow["city"]; ok && cityFromPage != "" {
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

			jsonBytes, _ := json.Marshal(forecastNow)
			fmt.Println(string(jsonBytes))

		} else {

			fmt.Fprintf(outWriter, cfg.ansiColourString("%s (<yellow>%s</>)\n"), cityFromPage, cfg.baseURL+cfg.city)
			fmt.Fprintf(outWriter,
				cfg.ansiColourString("Сейчас: <green>%d °C</>, <green>%s</>\n"),
				forecastNow["term_now"],
				forecastNow["desc_now"],
			)

			if _, ok := forecastNow["term_another_value1"]; ok {
				fmt.Fprint(outWriter, "  ")
				for _, num := range []string{"1", "2", "3", "4"} {
					if value, ok := forecastNow["term_another_value"+num].(int); ok {
						fmt.Fprintf(outWriter,
							cfg.ansiColourString("%s: <green>%d °C</> "),
							forecastNow["term_another_name"+num],
							value,
						)
					}
				}
				fmt.Fprint(outWriter, "\n")
			}

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
				if descLength < TodayForecastTableWidth {
					// align with today forecast
					descLength = TodayForecastTableWidth
				}

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
		fmt.Printf("City \"%s\" not found\n", cfg.city)
	}
}

//-----------------------------------------------------------------------------
func main() {
	cfg := getParams()
	forecastNow, forecastByHours, forecastNext := getWeather(cfg)
	render(forecastNow, forecastByHours, forecastNext, cfg)
}
