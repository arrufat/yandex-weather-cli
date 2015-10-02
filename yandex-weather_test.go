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
