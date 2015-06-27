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
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mattn/go-colorable"
	"github.com/mgutz/ansi"
)

const (
	// BASE_URL - yandex pogoda service url
	BASE_URL = "https://pogoda.yandex.ru/"
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

//-----------------------------------------------------------------------------
// get weather html page as http.Response
func get_weather_page(city string) *http.Response {
	cookie, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookie,
	}

	weather_url := BASE_URL + city
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
// suggest date from one day, returns human date and json date
func suggest_date(date string, order_num int) (string, string) {
	day, err := strconv.Atoi(clear_integer_in_string(date))
	if err != nil {
		return date, date
	}

	from := time.Now().AddDate(0, 0, order_num)

	for i := 0; day != from.Day() && i < 3; i++ {
		from = from.AddDate(0, 0, 1)
	}

	weekdays_ru := [...]string{
		"вс",
		"пн",
		"вт",
		"ср",
		"чт",
		"пт",
		"сб",
	}

	return from.Format("02.01") + " (" + weekdays_ru[from.Weekday()] + ")",
		from.Format("2006-01-02")
}

//-----------------------------------------------------------------------------
// safe convert string to int, return 0 on error
func convert_str_to_int(str string) int {
	number, err := strconv.Atoi(clear_integer_in_string(str))
	if err != nil {
		return 0
	}
	return number
}

//-----------------------------------------------------------------------------
// parse html via goquery, find DOM-nodes with weather forecast data
func get_weather(http_response *http.Response) (map[string]interface{}, []map[string]interface{}) {
	doc, err := goquery.NewDocumentFromResponse(http_response)
	if err != nil {
		log.Fatal(err)
	}

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
	}

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

	return forecast_now, forecast_next
}

//-----------------------------------------------------------------------------
// get command line parameters
func get_params() (string, bool, bool) {
	get_json := false
	no_color := false
	flag.BoolVar(&get_json, "json", false, "get JSON")
	flag.BoolVar(&no_color, "no-color", false, "disable colored output")
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] [city]\noptions:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Printf("\nexamples:\n  %s kiev\n  %s -json london\n", os.Args[0], os.Args[0])
	}
	flag.Parse()

	city := ""
	if flag.NArg() >= 1 {
		city = flag.Args()[0]
	}

	// detect pipe
	stdout_stat, _ := os.Stdout.Stat()
	if (stdout_stat.Mode() & os.ModeCharDevice) == 0 {
		no_color = true
	}

	return city, get_json, no_color
}

//-----------------------------------------------------------------------------
// get max length of string in slice of map of string
func get_max_length_in_slice(list []map[string]interface{}, key string) int {
	max_lengh := 0
	for _, row := range list {
		length := len([]rune(row[key].(string)))
		if max_lengh < length {
			max_lengh = length
		}
	}

	return max_lengh
}

//-----------------------------------------------------------------------------
// clear all non numeric symbols in string
func clear_integer_in_string(in string) (out string) {
	// replace dashes to minus
	out = regexp.MustCompile(string([]byte{0xE2, 0x88, 0x92})).ReplaceAllString(in, "-")

	// clear non numeric symbols
	out = regexp.MustCompile(`[^\d-]+`).ReplaceAllString(out, "")

	return out
}

//-----------------------------------------------------------------------------
// clear all non print symbols in string
func clear_nonprint_in_string(in string) (out string) {
	// replace spaces
	out = regexp.MustCompile(string([]byte{0xE2, 0x80, 0x89})).ReplaceAllString(in, " ")

	return out
}

//-----------------------------------------------------------------------------
// convert "<red>123</> str <green>456</green>" to ansi color string
// color -- color or simple remove color tags
func ansi_colour_string(str string, color bool) string {
	one_color := `(black|red|green|yellow|blue|magenta|cyan|white|\d{1,3})(\+[bBuih]+)?`
	re := regexp.MustCompile(`<(` + one_color + `(:` + one_color + `)?|/\w*)>`)
	result := re.ReplaceAllStringFunc(str, func(in string) string {
		if !color {
			return ""
		}

		out := in
		tag := in[1 : len(in)-1]

		if tag[0] == '/' {
			out = ansi.ColorCode("reset")
		} else {
			out = ansi.ColorCode(tag)
		}

		return out
	})

	return result
}

//-----------------------------------------------------------------------------
// render data as text or JSON
func render(forecast_now map[string]interface{}, forecast_next []map[string]interface{}, city string, get_json, no_color bool) {
	if _, ok := forecast_now["city"]; ok {
		// for windows
		out_writer := (io.Writer)(os.Stdout)
		if !no_color && runtime.GOOS == "windows" {
			out_writer = colorable.NewColorableStdout()
		}

		if !get_json {
			fmt.Fprintf(out_writer, ansi_colour_string("%s (<yellow>%s</>)\n", !no_color), forecast_now["city"], BASE_URL+city)
			fmt.Fprintf(out_writer,
				ansi_colour_string("Сейчас: <green>%d °C</>, <green>%s</>, %s: <green>%d °C</>, %s: <green>%d °C</>\n", !no_color),
				forecast_now["term_now"],
				forecast_now["desc_now"],
				forecast_now["term_another_name1"],
				forecast_now["term_another_value1"],
				forecast_now["term_another_name2"],
				forecast_now["term_another_value2"],
			)
			fmt.Fprintf(out_writer, ansi_colour_string("Давление: <green>%s</>\n", !no_color), forecast_now["pressure"])
			fmt.Fprintf(out_writer, ansi_colour_string("Влажность: <green>%s</>\n", !no_color), forecast_now["humidity"])
			fmt.Fprintf(out_writer, ansi_colour_string("Ветер: <green>%s</>\n", !no_color), forecast_now["wind"])
		}

		if len(forecast_next) > 0 {
			if get_json {
				for _, row := range forecast_next {
					row["date"] = row["json_date"]
					delete(row, "json_date")
				}
				forecast_now["next_days"] = forecast_next
			} else {
				desc_length := get_max_length_in_slice(forecast_next, "desc")
				fmt.Fprintf(out_writer, "%s\n", strings.Repeat("─", 27+desc_length))
				fmt.Fprintf(out_writer,
					ansi_colour_string("<blue+h> %-10s %4s %-*s %8s</>\n", !no_color),
					"дата",
					"°C",
					desc_length, "погода",
					"°C ночью",
				)
				fmt.Fprintf(out_writer, "%s\n", strings.Repeat("─", 27+desc_length))

				weekend_re := regexp.MustCompile(`(сб|вс)`)
				for _, row := range forecast_next {
					date := weekend_re.ReplaceAllString(row["date"].(string), ansi_colour_string("<red+h>$1</>", !no_color))
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

		if get_json {
			json, _ := json.Marshal(forecast_now)
			fmt.Println(string(json))
		}
	} else {
		fmt.Printf("City \"%s\" dont found\n", city)
	}
}

//-----------------------------------------------------------------------------
func main() {
	city, get_json, no_color := get_params()
	forecast_now, forecast_next := get_weather(get_weather_page(city))
	render(forecast_now, forecast_next, city, get_json, no_color)
}
