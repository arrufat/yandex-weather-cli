package main

import "testing"

func Test_clear_integer_in_string(t *testing.T) {
	testData := []struct {
		in  string
		out string
	}{
		{
			"42",
			"42",
		}, {
			" 42 ",
			"42",
		}, {
			"-42",
			"-42",
		}, {
			" -42 ",
			"-42",
		},
	}

	for _, item := range testData {
		out := clearIntegerInString(item.in)
		if out != item.out {
			t.Errorf("expected: %#v, real: %#v", item.out, out)
		}
	}
}

func Test_convert_str_to_int(t *testing.T) {
	testData := []struct {
		in  string
		out int
	}{
		{
			"42",
			42,
		}, {
			" 42 ",
			42,
		}, {
			"-42",
			-42,
		}, {
			" -42 ",
			-42,
		}, {
			"str 42 ",
			42,
		}, {
			"str",
			0,
		},
	}

	for _, item := range testData {
		out := convertStrToInt(item.in)
		if out != item.out {
			t.Errorf("expected: %#v, real: %#v", item.out, out)
		}
	}
}

func Test_parse_icon(t *testing.T) {
	testData := []struct {
		in  string
		out string
	}{
		{
			"",
			"",
		}, {
			"icon",
			"",
		}, {
			"icon icon_size_24 icon_snow",
			"icon_snow",
		}, {
			"icon icon_size_24 icon_rain",
			"icon_rain",
		},
	}

	for _, item := range testData {
		out := parseIcon(item.in)
		if out != item.out {
			t.Errorf("expected: %#v, real: %#v", item.out, out)
		}
	}
}

func Test_get_max_length_in_slice(t *testing.T) {
	testData := []struct {
		list []map[string]interface{}
		key  string
		out  int
	}{
		{
			[]map[string]interface{}{
				{"k1": "aaa"},
				{"k1": "aaaa"},
			},
			"k1",
			4,
		}, {
			[]map[string]interface{}{
				{"k1": "снег", "k2": "снегопад"},
				{"k1": "дождь"},
			},
			"k1",
			5,
		},
	}

	for _, item := range testData {
		out := getMaxLengthInSlice(item.list, item.key)
		if out != item.out {
			t.Errorf("expected: %#v, real: %#v", item.out, out)
		}
	}
}
