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
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/msoap/html2data"
)

// config - application config
type config struct {
	baseURL     string
	baseURLMini string
	city        string
	getJSON     bool
	noColor     bool
	noToday     bool
	daysLimit   int
}

// hourTemp - one hour temperature
type hourTemp struct {
	Hour int    `json:"hour"`
	Temp int    `json:"temp"`
	Icon string `json:"icon"`
}

// dayForecast - one day forecast
type dayForecast struct {
	DateHuman string `json:"-"`
	Date      string `json:"date"`
	Desc      string `json:"desc"`
	Temp      int    `json:"temp"`
	TempNight int    `json:"temp_night"`
}

var (
	version   = "1.15"
	userAgent = "yandex-weather-cli/" + version
)

const (
	// envBaseURLName - environment variable for setup base URL
	envBaseURLName = "Y_WEATHER_URL"
	// envBaseURLMiniName - environment variable for setup base URL (for days forecast)
	envBaseURLMiniName = "Y_WEATHER_MINI_URL"
	// baseURLDefault - yandex pogoda service url (testing: "http://localhost:8080/get?url=https://yandex.ru/pogoda/")
	baseURLDefault = "https://yandex.ru/pogoda/"
	// baseURLMiniDefault - url for forecast by hours (testing: "http://localhost:8080/get?url=https://p.ya.ru/")
	baseURLMiniDefault = "https://p.ya.ru/"
	// todayForecastTableWidth - today forecast table width for align tables
	todayForecastTableWidth = 14*4 - 27
)

// selectors - css selectors for forecast today
var selectors = map[string]string{
	"city":     "title",
	"term_now": "div.fact div.fact__temp",
	"desc_now": "div.fact div.link__condition",
	"wind":     "div.fact div.fact__props div.fact__wind-speed",
	"humidity": "div.fact div.fact__props div.fact__humidity",
	"pressure": "div.fact div.fact__props div.fact__pressure",
}

// selectorsNextDays - css selectors for forecast next days
var selectorsNextDays = map[string]string{
	"date":       "div.forecast-briefly__days time.time:attr(datetime)",
	"desc":       "div.forecast-briefly__days div.forecast-briefly__condition",
	"temp":       "div.forecast-briefly__days div.forecast-briefly__temp_day span.temp__value",
	"temp_night": "div.forecast-briefly__days div.forecast-briefly__temp_night span.temp__value",
}

// selectorByHoursRoot - Root element for forecast data
var selectorByHoursRoot = "div.temp-chart__wrap"

// selectorByHours - get forecast by hours
var selectorByHours = map[string]string{
	"hour": "p.temp-chart__hour",
	"temp": "div.temp-chart__temp",
	"icon": "i.icon:attr(class)",
}

// icons - unicode symbols for icon names
var icons = map[string]string{
	"icon_snow": "✻",
	"icon_rain": "☂",
}

//-----------------------------------------------------------------------------
// check if program's output used in *nix pipe
func outputIsPiped() bool {
	stdoutStat, err := os.Stdout.Stat()
	return err != nil || (stdoutStat.Mode()&os.ModeCharDevice) == 0
}

//-----------------------------------------------------------------------------
// get command line parameters
func getParams() (cfg config) {
	flag.BoolVar(&cfg.getJSON, "json", false, "get JSON")
	flag.BoolVar(&cfg.noColor, "no-color", false, "disable colored output")
	flag.BoolVar(&cfg.noToday, "no-today", false, "disable today forecast")
	flag.IntVar(&cfg.daysLimit, "days", 10, "maximum days to show")
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] [city]\noptions:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Printf("\nexamples:\n  %s kyiv\n  %s -json london\n", os.Args[0], os.Args[0])
	}
	getVersion := flag.Bool("version", false, "get version")
	flag.Parse()

	if *getVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	cfg.city = ""
	if flag.NArg() >= 1 {
		cfg.city = flag.Args()[0]
	}

	if runtime.GOOS == "windows" {
		// broken unicode symbols in cmd.exe and don't detect pipe
		cfg.noToday = true
	} else if outputIsPiped() {
		cfg.noColor = true
	}

	if baseURL := os.Getenv(envBaseURLName); len(baseURL) > 0 {
		cfg.baseURL = baseURL
	} else {
		cfg.baseURL = baseURLDefault
	}
	if baseURLMini := os.Getenv(envBaseURLMiniName); len(baseURLMini) > 0 {
		cfg.baseURLMini = baseURLMini
	} else {
		cfg.baseURLMini = baseURLMiniDefault
	}

	return cfg
}

//-----------------------------------------------------------------------------
// parse html via goquery, find DOM-nodes with weather forecast data
func getWeather(cfg config) (map[string]interface{}, []hourTemp, []dayForecast) {
	forecastNow := map[string]interface{}{}
	forecastNext := []dayForecast{}
	forecastByHours := []hourTemp{}

	reRemoveDesc := regexp.MustCompile(`^.+\s*:\s*`)
	reRemoveMultiline := regexp.MustCompile(`\n.+$`)
	reDate := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`)

	var extractNowForecast = func(doc html2data.Doc) {
		data, err := doc.GetDataFirst(selectors)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		for name := range selectors {
			forecastNow[name] = clearNonprintInString(data[name])
			switch name {
			case "city":
				forecastNow[name] = reRemoveMultiline.ReplaceAllString(forecastNow[name].(string), "")
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
	}

	var extractNextForecast = func(doc html2data.Doc) {
		dataNextDays, err := doc.GetData(selectorsNextDays)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if dateColumn, ok := dataNextDays["date"]; ok {
			now := time.Now()
		daysLoop:
			for i, dateStr := range dateColumn {
				if len(forecastNext) >= cfg.daysLimit {
					break daysLoop
				}

				if dateStr == "" {
					continue
				}

				currentDay := dayForecast{}
				for name := range selectorsNextDays {
					text := ""
					if _, ok := dataNextDays[name]; ok && len(dataNextDays[name]) >= i+1 {
						text = dataNextDays[name][i]
					} else {
						continue
					}
					text = clearNonprintInString(text)

					switch name {
					case "date":
						datesRaw := reDate.FindAllString(text, 1)
						if len(datesRaw) == 1 {
							curDate, err := time.Parse("2006-01-02", datesRaw[0])
							if err != nil || !curDate.Truncate(time.Hour*24).After(now.Truncate(time.Hour*24)) {
								continue daysLoop
							}
							currentDay.DateHuman, currentDay.Date = formatDates(curDate)
						}
					case "desc":
						currentDay.Desc = strings.ToLower(text)
					case "temp":
						currentDay.Temp = convertStrToInt(text)
					case "temp_night":
						currentDay.TempNight = convertStrToInt(text)
					}
				}

				if currentDay.Date != "" {
					forecastNext = append(forecastNext, currentDay)
				}
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		doc := html2data.FromURL(cfg.baseURL+cfg.city, html2data.URLCfg{UA: userAgent})
		extractNowForecast(doc)
		extractNextForecast(doc)
		wg.Done()
	}()

	go func() {
		// forecast by hours block
		if !cfg.noToday {
			docMini := html2data.FromURL(cfg.baseURLMini+cfg.city, html2data.URLCfg{UA: userAgent})
			dataHours, err := docMini.GetDataNestedFirst(selectorByHoursRoot, selectorByHours)
			if err == nil {
				for _, row := range dataHours {
					hour := convertStrToInt(row["hour"])
					temp := convertStrToInt(row["temp"])
					forecastByHours = append(forecastByHours, hourTemp{Hour: hour, Temp: temp, Icon: parseIcon(row["icon"])})
				}
			}
		}

		wg.Done()
	}()

	wg.Wait()
	return forecastNow, forecastByHours, forecastNext
}

//-----------------------------------------------------------------------------
// get icon name from css class attribut
func parseIcon(cssClass string) string {
	allAttributes := regexp.MustCompile(`\s+`).Split(cssClass, -1)
	for _, attr := range allAttributes {
		if _, ok := icons[attr]; ok {
			return attr
		}
	}
	return ""
}

//-----------------------------------------------------------------------------
// render data as text or JSON
func render(forecastNow map[string]interface{}, forecastByHours []hourTemp, forecastNext []dayForecast, cfg config) {
	cityFromPage, ok := forecastNow["city"]
	if !ok || cityFromPage == "" {
		fmt.Fprintf(os.Stderr, "City %q not found\n", cfg.city)
		os.Exit(1)
	}
	outWriter := getColorWriter(cfg.noColor)

	if cfg.getJSON {
		if !cfg.noToday && len(forecastByHours) > 0 {
			forecastNow["by_hours"] = forecastByHours
		}

		if len(forecastNext) > 0 {
			forecastNow["next_days"] = forecastNext
		}

		jsonBytes, _ := json.Marshal(forecastNow)
		fmt.Println(string(jsonBytes))
		return
	}

	outWriter.Printf(cfg.ansiColourString("%s (<yellow>%s</>)\n"), cityFromPage, cfg.baseURL+cfg.city)
	outWriter.Printf(
		cfg.ansiColourString("Сейчас: <green>%d °C</> - <green>%s</>\n"),
		forecastNow["term_now"],
		forecastNow["desc_now"],
	)

	outWriter.Printf(cfg.ansiColourString("Давление: <green>%s</>\n"), forecastNow["pressure"])
	outWriter.Printf(cfg.ansiColourString("Влажность: <green>%s</>\n"), forecastNow["humidity"])
	outWriter.Printf(cfg.ansiColourString("Ветер: <green>%s</>\n"), forecastNow["wind"])

	if !cfg.noToday && len(forecastByHours) > 0 {
		textByHour := [4]string{}
		for _, item := range forecastByHours {
			textByHour[0] += fmt.Sprintf("%3d ", item.Hour)
			textByHour[2] += fmt.Sprintf("%3d°", item.Temp)
			icon, exists := icons[item.Icon]
			if !exists {
				icon = " "
			}
			textByHour[3] += fmt.Sprintf(cfg.ansiColourString("<blue>%3s</blue> "), icon)
		}
		textByHour[1] = cfg.ansiColourString("<grey+h>" + renderHisto(forecastByHours) + "</>")

		outWriter.Println(strings.Repeat("─", len(forecastByHours)*4))
		outWriter.Printf("%s\n%s\n%s\n%s\n",
			cfg.ansiColourString("<grey+h>"+textByHour[0]+"</>"),
			textByHour[1],
			textByHour[2],
			textByHour[3],
		)
	}

	if len(forecastNext) > 0 {
		descLength := getMaxLengthDesc(forecastNext)
		if descLength < todayForecastTableWidth {
			// align with today forecast
			descLength = todayForecastTableWidth
		}

		outWriter.Println(strings.Repeat("─", 27+descLength))
		outWriter.Printf(
			cfg.ansiColourString("<blue+h> %-10s %4s %-*s %8s</>\n"),
			"дата",
			"°C",
			descLength, "погода",
			"°C ночью",
		)
		outWriter.Println(strings.Repeat("─", 27+descLength))

		weekendRe := regexp.MustCompile(`(сб|вс)`)
		for _, row := range forecastNext {
			date := weekendRe.ReplaceAllString(row.DateHuman, cfg.ansiColourString("<red+h>$1</>"))
			outWriter.Printf(
				" %10s %3d° %-*s %7d°\n",
				date,
				row.Temp,
				descLength,
				row.Desc,
				row.TempNight,
			)
		}
	}
}

//-----------------------------------------------------------------------------
func main() {
	cfg := getParams()
	forecastNow, forecastByHours, forecastNext := getWeather(cfg)
	render(forecastNow, forecastByHours, forecastNext, cfg)
}
