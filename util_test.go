package main

import (
	"testing"

	"github.com/mgutz/ansi"
)

func Test_clearNonprintInString(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantOut string
	}{
		{"simple string", "str", "str"},
		{"string with unprinted", string([]byte{0xE2, 0x80, 0x89}) + "str", " str"},
	}

	for _, tt := range tests {
		if gotOut := clearNonprintInString(tt.in); gotOut != tt.wantOut {
			t.Errorf("%q. clearNonprintInString() = '%v', want '%v'", tt.name, gotOut, tt.wantOut)
		}
	}
}

func Test_ansiColourString(t *testing.T) {
	tests := []struct {
		name    string
		city    string
		getJSON bool
		noColor bool
		noToday bool
		str     string
		want    string
	}{
		{
			name:    "simple",
			noColor: true,
			str:     "string",
			want:    "string",
		},
		{
			name:    "simple color",
			noColor: false,
			str:     "string",
			want:    "string",
		},
		{
			name:    "with noColor, with tag",
			noColor: true,
			str:     "string <green>green</>",
			want:    "string green",
		},
		{
			name:    "with color, with tag",
			noColor: false,
			str:     "string <green>green</>",
			want:    "string " + ansi.ColorCode("green") + "green" + ansi.ColorCode("reset"),
		},
		{
			name:    "with color, with tag",
			noColor: false,
			str:     "string <green>green</green>",
			want:    "string " + ansi.ColorCode("green") + "green" + ansi.ColorCode("reset"),
		},
		{
			name:    "with color, with unclosed tag",
			noColor: false,
			str:     "string <green>green",
			want:    "string " + ansi.ColorCode("green") + "green",
		},
	}

	for _, tt := range tests {
		cfg := config{
			city:    tt.city,
			getJSON: tt.getJSON,
			noColor: tt.noColor,
			noToday: tt.noToday,
		}
		if got := cfg.ansiColourString(tt.str); got != tt.want {
			t.Errorf("%q. Config.ansiColourString() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func Test_renderHisto(t *testing.T) {
	tests := []struct {
		name            string
		forecastByHours []hourTemp
		want            string
	}{
		{
			name: "1",
			forecastByHours: []hourTemp{
				{
					Hour: 17,
					Temp: -3,
					Icon: "icon_rain",
				},
				{
					Hour: 18,
					Temp: 1,
					Icon: "icon_snow",
				},
				{
					Hour: 19,
					Temp: 1,
					Icon: "icon_snow",
				},
				{
					Hour: 20,
					Temp: 0,
					Icon: "icon_snow",
				},
				{
					Hour: 21,
					Temp: 0,
					Icon: "",
				},
				{
					Hour: 22,
					Temp: -2,
					Icon: "",
				},
				{
					Hour: 23,
					Temp: -5,
					Icon: "",
				},
				{
					Hour: 0,
					Temp: -3,
					Icon: "icon_snow",
				},
				{
					Hour: 1,
					Temp: -1,
					Icon: "icon_snow",
				},
				{
					Hour: 2,
					Temp: -1,
					Icon: "icon_snow",
				},
				{
					Hour: 3,
					Temp: -1,
					Icon: "icon_snow",
				},
				{
					Hour: 4,
					Temp: -1,
					Icon: "icon_snow",
				},
				{
					Hour: 5,
					Temp: -2,
					Icon: "icon_snow",
				},
				{
					Hour: 6,
					Temp: -1,
					Icon: "icon_snow",
				},
			},
			want: "▃▄▅▆█████▇▇▇▆▆▆▆▆▆▅▅▄▃▂▁▁▁▂▂▃▃▄▅▅▅▅▅▅▅▅▅▅▅▅▅▅▅▅▄▄▄▅▅▅▅▅▅",
		},
		{
			name: "all same temperature",
			forecastByHours: []hourTemp{
				{
					Hour: 17,
					Temp: 1,
					Icon: "icon_rain",
				},
				{
					Hour: 18,
					Temp: 1,
					Icon: "icon_snow",
				},
				{
					Hour: 19,
					Temp: 1,
					Icon: "icon_snow",
				},
			},
			want: "▁▁▁▁▁▁▁▁▁▁▁▁",
		},
		{
			name: "all same negative temperature",
			forecastByHours: []hourTemp{
				{
					Hour: 17,
					Temp: -10,
					Icon: "icon_rain",
				},
				{
					Hour: 18,
					Temp: -10,
					Icon: "icon_snow",
				},
				{
					Hour: 19,
					Temp: -10,
					Icon: "icon_snow",
				},
			},
			want: "▁▁▁▁▁▁▁▁▁▁▁▁",
		},
	}

	for _, tt := range tests {
		if got := renderHisto(tt.forecastByHours); got != tt.want {
			t.Errorf("%q. renderHisto() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func Test_getColorWriter(t *testing.T) {
	getColorWriter(true)
}
