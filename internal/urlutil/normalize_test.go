package urlutil

import (
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		base     string
		want     string
		wantErr  bool
	}{
		{
			name:  "absolute URL with query string",
			input: "https://example.com/path?foo=bar&baz=qux",
			want:  "https://example.com/path",
		},
		{
			name:  "absolute URL preserves fragment",
			input: "https://example.com/path#section",
			want:  "https://example.com/path#section",
		},
		{
			name:  "absolute URL with query and fragment",
			input: "https://example.com/path?key=value#section",
			want:  "https://example.com/path#section",
		},
		{
			name:  "lowercase scheme and domain",
			input: "HTTPS://EXAMPLE.COM/Path",
			want:  "https://example.com/Path",
		},
		{
			name:  "remove default port for https",
			input: "https://example.com:443/path",
			want:  "https://example.com/path",
		},
		{
			name:  "remove default port for http",
			input: "http://example.com:80/path",
			want:  "http://example.com/path",
		},
		{
			name:  "keep non-default port",
			input: "https://example.com:8080/path",
			want:  "https://example.com:8080/path",
		},
		{
			name:  "relative URL with base",
			input: "../other/page.html",
			base:  "https://example.com/docs/current/index.html",
			want:  "https://example.com/docs/other/page.html",
		},
		{
			name:  "absolute path with base",
			input: "/absolute/path.html",
			base:  "https://example.com/docs/current/index.html",
			want:  "https://example.com/absolute/path.html",
		},
		{
			name:  "remove trailing slash from path",
			input: "https://example.com/path/",
			want:  "https://example.com/path",
		},
		{
			name:  "keep root trailing slash",
			input: "https://example.com/",
			want:  "https://example.com/",
		},
		{
			name:    "invalid URL",
			input:   "://invalid",
			wantErr: true,
		},
		{
			name:  "empty query parameter",
			input: "https://example.com/path?",
			want:  "https://example.com/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Normalize(tt.input, tt.base)
			if (err != nil) != tt.wantErr {
				t.Errorf("Normalize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got != tt.want {
				t.Errorf("Normalize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeWithoutBase(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "simple absolute URL",
			input: "https://example.com/path",
			want:  "https://example.com/path",
		},
		{
			name:  "URL with query removed",
			input: "https://example.com/path?key=value",
			want:  "https://example.com/path",
		},
		{
			name:    "relative URL without base should error",
			input:   "../relative/path",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Normalize(tt.input, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("Normalize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got != tt.want {
				t.Errorf("Normalize() = %v, want %v", got, tt.want)
			}
		})
	}
}
