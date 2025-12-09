package urlutil

import (
	"testing"
)

func TestFilter_IsAllowed(t *testing.T) {
	// Create filter with root URL and whitelist
	filter := NewFilter(
		"https://docs.unrealengine.com/udk/Two/SiteMap.html",
		[]string{"cdn.example.com", "static.unrealengine.com"},
	)

	tests := []struct {
		name    string
		url     string
		want    bool
		wantErr bool
	}{
		{
			name: "same domain and path prefix",
			url:  "https://docs.unrealengine.com/udk/Two/WebHome.html",
			want: true,
		},
		{
			name: "same domain but different path prefix",
			url:  "https://docs.unrealengine.com/udk/Three/index.html",
			want: false,
		},
		{
			name: "whitelisted CDN domain",
			url:  "https://cdn.example.com/assets/image.png",
			want: true,
		},
		{
			name: "whitelisted static domain",
			url:  "https://static.unrealengine.com/script.js",
			want: true,
		},
		{
			name: "non-whitelisted external domain",
			url:  "https://external.com/page.html",
			want: false,
		},
		{
			name: "subdirectory of allowed path",
			url:  "https://docs.unrealengine.com/udk/Two/API/Core/index.html",
			want: true,
		},
		{
			name: "case insensitive domain check",
			url:  "https://DOCS.UNREALENGINE.COM/udk/Two/page.html",
			want: true,
		},
		{
			name: "http vs https same domain",
			url:  "http://docs.unrealengine.com/udk/Two/page.html",
			want: true,
		},
		{
			name:    "invalid URL",
			url:     "://invalid",
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filter.IsAllowed(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Filter.IsAllowed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Filter.IsAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_GetResourceType(t *testing.T) {
	filter := NewFilter("https://example.com/", nil)

	tests := []struct {
		name        string
		url         string
		contentType string
		want        ResourceType
	}{
		{
			name: "HTML by extension",
			url:  "https://example.com/page.html",
			want: ResourceHTML,
		},
		{
			name: "HTML by extension (htm)",
			url:  "https://example.com/page.htm",
			want: ResourceHTML,
		},
		{
			name: "CSS by extension",
			url:  "https://example.com/style.css",
			want: ResourceCSS,
		},
		{
			name: "JavaScript by extension",
			url:  "https://example.com/script.js",
			want: ResourceJS,
		},
		{
			name: "PNG image by extension",
			url:  "https://example.com/image.png",
			want: ResourceImage,
		},
		{
			name: "JPEG image by extension",
			url:  "https://example.com/photo.jpg",
			want: ResourceImage,
		},
		{
			name: "GIF image by extension",
			url:  "https://example.com/anim.gif",
			want: ResourceImage,
		},
		{
			name: "SVG image by extension",
			url:  "https://example.com/icon.svg",
			want: ResourceImage,
		},
		{
			name: "WEBP image by extension",
			url:  "https://example.com/image.webp",
			want: ResourceImage,
		},
		{
			name: "Font file (woff)",
			url:  "https://example.com/font.woff",
			want: ResourceOther,
		},
		{
			name: "Font file (woff2)",
			url:  "https://example.com/font.woff2",
			want: ResourceOther,
		},
		{
			name: "Font file (ttf)",
			url:  "https://example.com/font.ttf",
			want: ResourceOther,
		},
		{
			name: "PDF file",
			url:  "https://example.com/doc.pdf",
			want: ResourceOther,
		},
		{
			name:        "HTML by content type",
			url:         "https://example.com/page",
			contentType: "text/html",
			want:        ResourceHTML,
		},
		{
			name:        "CSS by content type",
			url:         "https://example.com/styles",
			contentType: "text/css",
			want:        ResourceCSS,
		},
		{
			name:        "JavaScript by content type",
			url:         "https://example.com/app",
			contentType: "application/javascript",
			want:        ResourceJS,
		},
		{
			name:        "JavaScript by content type (text/javascript)",
			url:         "https://example.com/app",
			contentType: "text/javascript",
			want:        ResourceJS,
		},
		{
			name:        "PNG by content type",
			url:         "https://example.com/img",
			contentType: "image/png",
			want:        ResourceImage,
		},
		{
			name:        "JPEG by content type",
			url:         "https://example.com/img",
			contentType: "image/jpeg",
			want:        ResourceImage,
		},
		{
			name: "unknown extension defaults to Other",
			url:  "https://example.com/file.xyz",
			want: ResourceOther,
		},
		{
			name: "no extension defaults to HTML",
			url:  "https://example.com/page",
			want: ResourceHTML,
		},
		{
			name: "extension with query string",
			url:  "https://example.com/page.html?v=1",
			want: ResourceHTML,
		},
		{
			name: "extension with fragment",
			url:  "https://example.com/page.html#section",
			want: ResourceHTML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter.GetResourceType(tt.url, tt.contentType)
			if got != tt.want {
				t.Errorf("Filter.GetResourceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewFilter(t *testing.T) {
	tests := []struct {
		name      string
		rootURL   string
		whitelist []string
		wantErr   bool
	}{
		{
			name:      "valid root URL",
			rootURL:   "https://example.com/docs/index.html",
			whitelist: nil,
			wantErr:   false,
		},
		{
			name:      "valid root URL with whitelist",
			rootURL:   "https://example.com/",
			whitelist: []string{"cdn.example.com", "static.example.com"},
			wantErr:   false,
		},
		{
			name:      "invalid root URL",
			rootURL:   "://invalid",
			whitelist: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFilter(tt.rootURL, tt.whitelist)
			// If we expect an error, the filter operations should handle it
			if !tt.wantErr && filter == nil {
				t.Errorf("NewFilter() returned nil for valid input")
			}
		})
	}
}
