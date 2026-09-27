package logger

import "testing"

func TestParseLevel(t *testing.T) {
	cases := []struct {
		name    string
		want    Level
		wantErr bool
	}{
		{"debug", DebugLevel, false},
		{"INFO", InfoLevel, false},
		{" warn ", WarnLevel, false},
		{"warning", WarnLevel, false},
		{"Error", ErrorLevel, false},
		{"", InfoLevel, true},
		{"verbose", InfoLevel, true},
	}
	for _, tc := range cases {
		got, err := ParseLevel(tc.name)
		if got != tc.want || (err != nil) != tc.wantErr {
			t.Errorf("ParseLevel(%q) = %v, %v; want %v, err=%v", tc.name, got, err, tc.want, tc.wantErr)
		}
	}
}
