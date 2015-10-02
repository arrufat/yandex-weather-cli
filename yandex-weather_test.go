package main

import "testing"

func Test_clear_integer_in_string(t *testing.T) {
	test_data := []struct {
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

	for _, item := range test_data {
		out := clear_integer_in_string(item.in)
		if out != item.out {
			t.Errorf("expected: %#v, real: %#v", item.out, out)
		}
	}
}

func Test_convert_str_to_int(t *testing.T) {
	test_data := []struct {
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

	for _, item := range test_data {
		out := convert_str_to_int(item.in)
		if out != item.out {
			t.Errorf("expected: %#v, real: %#v", item.out, out)
		}
	}
}

func Test_parse_icon(t *testing.T) {
	test_data := []struct {
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

	for _, item := range test_data {
		out := parse_icon(item.in)
		if out != item.out {
			t.Errorf("expected: %#v, real: %#v", item.out, out)
		}
	}
}

func Test_get_max_length_in_slice(t *testing.T) {
	test_data := []struct {
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

	for _, item := range test_data {
		out := get_max_length_in_slice(item.list, item.key)
		if out != item.out {
			t.Errorf("expected: %#v, real: %#v", item.out, out)
		}
	}
}
