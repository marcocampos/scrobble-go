package lastfm

import "testing"

func TestMD5(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"password", "5f4dcc3b5aa765d61d8327deb882cf99"},
		{"pylast", "3cbe303fc28f649f9ce216919e64a0fd"},
	}
	for _, tt := range tests {
		got := MD5(tt.input)
		if got != tt.want {
			t.Errorf("MD5(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseNumber(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"", 0},
		{"0", 0},
		{"42", 42},
		{"3.14", 3.14},
		{"1234567", 1234567},
		{"notanumber", 0},
	}
	for _, tt := range tests {
		got := parseNumber(tt.input)
		if got != tt.want {
			t.Errorf("parseNumber(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"0", 0},
		{"42", 42},
		{"-7", -7},
		{"notanumber", 0},
	}
	for _, tt := range tests {
		got := parseInt(tt.input)
		if got != tt.want {
			t.Errorf("parseInt(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
