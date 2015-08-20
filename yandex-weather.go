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
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Config - application config
type Config struct {
	city     string
	get_json bool
	no_color bool
	no_today bool
}

// HourTemp - one hour temperature
type HourTemp struct {
	Hour int    `json:"hour"`
	Temp int    `json:"temp"`
	Icon string `json:"icon"`
}

const (
	// BASE_URL - yandex pogoda service url (testing: "http://localhost:8080/get?url=https://pogoda.yandex.ru/")
	BASE_URL = "https://pogoda.yandex.ru/"
	// BASE_URL_MINI - url for forecast by hours (testing: "http://localhost:8080/get?url=https://p.ya.ru/")
	BASE_URL_MINI = "https://p.ya.ru/"
	// USER_AGENT - for http.request
	USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit/600.1.25 (KHTML, like Gecko) Version/8.0 Safari/600.1.25"
	// FORECAST_DAYS - parse days in forecast
	FORECAST_DAYS = 10
)

// SELECTORS - css selectors for forecast today
var SELECTORS = map[string]string{
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

// SELECTORS_NEXT_DAYS - css selectors for forecast next days
var SELECTORS_NEXT_DAYS = map[string]string{
	"date":       "div.tabs-panes span.forecast-brief__item-day",
	"desc":       "div.tabs-panes div.forecast-brief__item-comment",
	"term":       "div.tabs-panes div.forecast-brief__item-temp-day",
	"term_night": "div.tabs-panes div.forecast-brief__item-temp-night",
}

// SELECTORS_BY_HOURS - get forecast by hours
var SELECTOR_BY_HOURS = map[string]string{
	"root": "div.temperatures div.chart_wrapper",
	"hour": "p.th",
	"temp": "span.chart_temperature",
	"icon": "span:nth-child(3)",
}

// ICONS - unicode symbols for icon names
var ICONS = map[string]string{
	"fake_icon": " ",
	"icon_snow": "✻",
	"icon_rain": "☂",
}

//-----------------------------------------------------------------------------
// get command line parameters
func get_params() (cfg Config) {
	flag.BoolVar(&cfg.get_json, "json", false, "get JSON")
	flag.BoolVar(&cfg.no_color, "no-color", false, "disable colored output")
	flag.BoolVar(&cfg.no_today, "no-today", false, "disable today forecast")
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

	// detect pipe
	stdout_stat, _ := os.Stdout.Stat()
	if (stdout_stat.Mode() & os.ModeCharDevice) == 0 {
		cfg.no_color = true
	}

	return cfg
}

//-----------------------------------------------------------------------------
// get weather html page as http.Response
func get_weather_page(weather_url string) *http.Response {
	cookie, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookie,
	}

	request, err := http.NewRequest("GET", weather_url, nil)
	if err != nil {
		log.Fatal(err)
	}

	request.Header.Set("User-Agent", USER_AGENT)

	// create request for set cookies only
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	response, err = client.Get(weather_url)

	if err != nil {
		log.Fatal(err)
	}

	return response
}

//-----------------------------------------------------------------------------
// parse html via goquery, find DOM-nodes with weather forecast data
func get_weather(cfg Config) (map[string]interface{}, []HourTemp, []map[string]interface{}) {
	http_response := get_weather_page(BASE_URL + cfg.city)

	doc, err := goquery.NewDocumentFromResponse(http_response)
	if err != nil {
		log.Fatal(err)
	}

	// now block
	forecast_now := map[string]interface{}{}

	re_remove_desc := regexp.MustCompile(`^.+\s*:\s*`)
	for name, selector := range SELECTORS {
		doc.Find(selector).Each(func(i int, selection *goquery.Selection) {
			forecast_now[name] = clear_nonprint_in_string(selection.Text())
			switch name {
			case "humidity", "pressure", "wind":
				forecast_now[name] = re_remove_desc.ReplaceAllString(forecast_now[name].(string), "")
			case "term_now", "term_another_value1", "term_another_value2":
				forecast_now[name] = convert_str_to_int(forecast_now[name].(string))
			}
		})
		if name == "wind" && forecast_now[name] == nil {
			forecast_now[name] = "0 м/с"
		}
	}

	// forecast for next days block
	forecast_next := make([]map[string]interface{}, 0, FORECAST_DAYS)
	for name, selector := range SELECTORS_NEXT_DAYS {
		doc.Find(selector).Each(func(i int, selection *goquery.Selection) {
			if len(forecast_next)-1 < i {
				forecast_next = append(forecast_next, map[string]interface{}{})
			}

			forecast_next[i][name] = clear_nonprint_in_string(selection.Text())
		})
	}

	// suggest dates
	for i := range forecast_next {
		forecast_next[i]["date"], forecast_next[i]["json_date"] = suggest_date(forecast_next[i]["date"].(string), i)
		forecast_next[i]["term"] = convert_str_to_int(forecast_next[i]["term"].(string))
		forecast_next[i]["term_night"] = convert_str_to_int(forecast_next[i]["term_night"].(string))
	}

	// forecast by hours block
	var forecast_by_hours []HourTemp
	if !cfg.no_today {
		http_response = get_weather_page(BASE_URL_MINI + cfg.city)
		doc, err = goquery.NewDocumentFromResponse(http_response)
		if err != nil {
			log.Fatal(err)
		}

		doc.Find(SELECTOR_BY_HOURS["root"]).Each(func(i int, selection *goquery.Selection) {
			hour := convert_str_to_int(selection.Find(SELECTOR_BY_HOURS["hour"]).Text())
			temp := convert_str_to_int(selection.Find(SELECTOR_BY_HOURS["temp"]).Text())
			icon, _ := selection.Find(SELECTOR_BY_HOURS["icon"]).Attr("class")
			forecast_by_hours = append(forecast_by_hours, HourTemp{Hour: hour, Temp: temp, Icon: icon})
		})
	}

	return forecast_now, forecast_by_hours, forecast_next
}

//-----------------------------------------------------------------------------
// render data as text or JSON
func render(forecast_now map[string]interface{}, forecast_by_hours []HourTemp, forecast_next []map[string]interface{}, cfg Config) {
	if _, ok := forecast_now["city"]; ok {
		out_writer := get_color_writer(cfg.no_color)

		if cfg.get_json {

			if !cfg.no_today && len(forecast_by_hours) > 0 {
				forecast_now["by_hours"] = forecast_by_hours
			}

			if len(forecast_next) > 0 {
				for _, row := range forecast_next {
					row["date"] = row["json_date"]
					delete(row, "json_date")
				}
				forecast_now["next_days"] = forecast_next
			}

			json, _ := json.Marshal(forecast_now)
			fmt.Println(string(json))

		} else {

			fmt.Fprintf(out_writer, cfg.ansi_colour_string("%s (<yellow>%s</>)\n"), forecast_now["city"], BASE_URL+cfg.city)
			fmt.Fprintf(out_writer,
				cfg.ansi_colour_string("Сейчас: <green>%d °C</>, <green>%s</>, %s: <green>%d °C</>, %s: <green>%d °C</>\n"),
				forecast_now["term_now"],
				forecast_now["desc_now"],
				forecast_now["term_another_name1"],
				forecast_now["term_another_value1"],
				forecast_now["term_another_name2"],
				forecast_now["term_another_value2"],
			)
			fmt.Fprintf(out_writer, cfg.ansi_colour_string("Давление: <green>%s</>\n"), forecast_now["pressure"])
			fmt.Fprintf(out_writer, cfg.ansi_colour_string("Влажность: <green>%s</>\n"), forecast_now["humidity"])
			fmt.Fprintf(out_writer, cfg.ansi_colour_string("Ветер: <green>%s</>\n"), forecast_now["wind"])

			if !cfg.no_today && len(forecast_by_hours) > 0 {
				text_by_hour := [4]string{}
				for _, item := range forecast_by_hours {
					text_by_hour[0] += fmt.Sprintf("%3d ", item.Hour)
					text_by_hour[2] += fmt.Sprintf("%3d°", item.Temp)
					icon, exists := ICONS[item.Icon]
					if !exists {
						icon = " "
					}
					text_by_hour[3] += fmt.Sprintf(cfg.ansi_colour_string("<blue>%3s</blue> "), icon)
				}
				text_by_hour[1] = cfg.ansi_colour_string("<grey+h>" + render_histo(forecast_by_hours) + "</>")

				fmt.Fprintf(out_writer, strings.Repeat("─", len(forecast_by_hours)*4)+"\n")
				fmt.Fprintf(out_writer, "%s\n%s\n%s\n%s\n",
					cfg.ansi_colour_string("<grey+h>"+text_by_hour[0]+"</>"),
					text_by_hour[1],
					text_by_hour[2],
					text_by_hour[3],
				)
			}

			if len(forecast_next) > 0 {
				desc_length := get_max_length_in_slice(forecast_next, "desc")
				fmt.Fprintf(out_writer, "%s\n", strings.Repeat("─", 27+desc_length))
				fmt.Fprintf(out_writer,
					cfg.ansi_colour_string("<blue+h> %-10s %4s %-*s %8s</>\n"),
					"дата",
					"°C",
					desc_length, "погода",
					"°C ночью",
				)
				fmt.Fprintf(out_writer, "%s\n", strings.Repeat("─", 27+desc_length))

				weekend_re := regexp.MustCompile(`(сб|вс)`)
				for _, row := range forecast_next {
					date := weekend_re.ReplaceAllString(row["date"].(string), cfg.ansi_colour_string("<red+h>$1</>"))
					fmt.Fprintf(out_writer,
						" %10s %3d° %-*s %7d°\n",
						date,
						row["term"].(int),
						desc_length,
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
	cfg := get_params()
	forecast_now, forecast_by_hours, forecast_next := get_weather(cfg)
	render(forecast_now, forecast_by_hours, forecast_next, cfg)
}
