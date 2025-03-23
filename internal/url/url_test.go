package url

import (
	"os"
	"testing"
)

func TestParseURL(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		name     string
		url      string
		expected string
		wantErr  bool
	}{
		{
			name:     "Parse URL with absolute path",
			url:      "path/to/file",
			expected: wd + "/path/to/file",
			wantErr:  false,
		},
		{
			name:     "Parse empty URL",
			url:      "",
			expected: wd,
			wantErr:  false,
		},
		{
			name:     "Parse URL with relative path",
			url:      "relative/path",
			expected: wd + "/relative/path",
			wantErr:  false,
		},
		{
			name:     "Parse invalid URL",
			url:      "\asd\asd",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Parse URL with only relative path",
			url:      "/relative/path",
			expected: "/relative/path",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseURL(tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}
