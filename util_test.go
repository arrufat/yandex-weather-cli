package main

import (
	"testing"
	"time"
)

func TestSuggestDate(t *testing.T) {
	now, _ := time.Parse("2006-01-02", "2016-03-17")
	tests := []struct {
		name     string
		now      time.Time
		date     string
		orderNum int
		want     string
		want1    string
	}{
		{
			name:     "today",
			now:      now,
			date:     "17",
			orderNum: 0,
			want:     "17.03 (чт)",
			want1:    "2016-03-17",
		},
		{
			name:     "next day",
			now:      now,
			date:     "18",
			orderNum: 1,
			want:     "18.03 (пт)",
			want1:    "2016-03-18",
		},
		{
			name:     "last day",
			now:      now,
			date:     "26",
			orderNum: 9,
			want:     "26.03 (сб)",
			want1:    "2016-03-26",
		},
		{
			name:     "error day string",
			now:      now,
			date:     "26-03",
			orderNum: 0,
			want:     "26-03",
			want1:    "26-03",
		},
		{
			name:     "suggest by step next day",
			now:      now,
			date:     "18",
			orderNum: 0,
			want:     "18.03 (пт)",
			want1:    "2016-03-18",
		},
	}
	for _, tt := range tests {
		got, got1 := suggestDate(tt.now, tt.date, tt.orderNum)
		if got != tt.want {
			t.Errorf("%q. suggestDate() got = %v, want %v", tt.name, got, tt.want)
		}
		if got1 != tt.want1 {
			t.Errorf("%q. suggestDate() got1 = %v, want %v", tt.name, got1, tt.want1)
		}
	}
}
