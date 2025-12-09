package fetcher

import "testing"

func TestDetectResourceType_ByContentType(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		contentType string
		expected    ResourceType
	}{
		{
			name:        "HTML by content-type",
			url:         "https://example.com/page",
			contentType: "text/html; charset=utf-8",
			expected:    ResourceTypeHTML,
		},
		{
			name:        "CSS by content-type",
			url:         "https://example.com/style",
			contentType: "text/css",
			expected:    ResourceTypeCSS,
		},
		{
			name:        "JavaScript by content-type (application/javascript)",
			url:         "https://example.com/script",
			contentType: "application/javascript",
			expected:    ResourceTypeJavaScript,
		},
		{
			name:        "JavaScript by content-type (text/javascript)",
			url:         "https://example.com/script",
			contentType: "text/javascript",
			expected:    ResourceTypeJavaScript,
		},
		{
			name:        "JavaScript by content-type (application/x-javascript)",
			url:         "https://example.com/script",
			contentType: "application/x-javascript",
			expected:    ResourceTypeJavaScript,
		},
		{
			name:        "PNG image by content-type",
			url:         "https://example.com/photo",
			contentType: "image/png",
			expected:    ResourceTypeImage,
		},
		{
			name:        "JPEG image by content-type",
			url:         "https://example.com/photo",
			contentType: "image/jpeg",
			expected:    ResourceTypeImage,
		},
		{
			name:        "SVG image by content-type",
			url:         "https://example.com/icon",
			contentType: "image/svg+xml",
			expected:    ResourceTypeImage,
		},
		{
			name:        "WOFF font by content-type",
			url:         "https://example.com/font",
			contentType: "font/woff2",
			expected:    ResourceTypeFont,
		},
		{
			name:        "TTF font by content-type",
			url:         "https://example.com/font",
			contentType: "font/ttf",
			expected:    ResourceTypeFont,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectResourceType(tt.url, tt.contentType)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDetectResourceType_ByExtension(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected ResourceType
	}{
		{
			name:     "HTML by .html extension",
			url:      "https://example.com/page.html",
			expected: ResourceTypeHTML,
		},
		{
			name:     "HTML by .htm extension",
			url:      "https://example.com/page.htm",
			expected: ResourceTypeHTML,
		},
		{
			name:     "CSS by .css extension",
			url:      "https://example.com/style.css",
			expected: ResourceTypeCSS,
		},
		{
			name:     "JavaScript by .js extension",
			url:      "https://example.com/script.js",
			expected: ResourceTypeJavaScript,
		},
		{
			name:     "JavaScript by .mjs extension",
			url:      "https://example.com/module.mjs",
			expected: ResourceTypeJavaScript,
		},
		{
			name:     "PNG image",
			url:      "https://example.com/photo.png",
			expected: ResourceTypeImage,
		},
		{
			name:     "JPEG image (.jpg)",
			url:      "https://example.com/photo.jpg",
			expected: ResourceTypeImage,
		},
		{
			name:     "JPEG image (.jpeg)",
			url:      "https://example.com/photo.jpeg",
			expected: ResourceTypeImage,
		},
		{
			name:     "GIF image",
			url:      "https://example.com/animation.gif",
			expected: ResourceTypeImage,
		},
		{
			name:     "SVG image",
			url:      "https://example.com/icon.svg",
			expected: ResourceTypeImage,
		},
		{
			name:     "WebP image",
			url:      "https://example.com/photo.webp",
			expected: ResourceTypeImage,
		},
		{
			name:     "ICO image",
			url:      "https://example.com/favicon.ico",
			expected: ResourceTypeImage,
		},
		{
			name:     "WOFF font",
			url:      "https://example.com/font.woff",
			expected: ResourceTypeFont,
		},
		{
			name:     "WOFF2 font",
			url:      "https://example.com/font.woff2",
			expected: ResourceTypeFont,
		},
		{
			name:     "TTF font",
			url:      "https://example.com/font.ttf",
			expected: ResourceTypeFont,
		},
		{
			name:     "OTF font",
			url:      "https://example.com/font.otf",
			expected: ResourceTypeFont,
		},
		{
			name:     "PDF file",
			url:      "https://example.com/document.pdf",
			expected: ResourceTypeOther,
		},
		{
			name:     "ZIP file",
			url:      "https://example.com/archive.zip",
			expected: ResourceTypeOther,
		},
		{
			name:     "Unknown extension",
			url:      "https://example.com/file.xyz",
			expected: ResourceTypeUnknown,
		},
		{
			name:     "No extension",
			url:      "https://example.com/page",
			expected: ResourceTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectResourceType(tt.url, "")
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDetectResourceType_ContentTypeTakesPrecedence(t *testing.T) {
	// Content-Type header should take precedence over file extension
	url := "https://example.com/page.js"
	contentType := "text/html"

	result := DetectResourceType(url, contentType)

	if result != ResourceTypeHTML {
		t.Errorf("expected HTML (from content-type), got %v", result)
	}
}

func TestDetectResourceType_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		contentType string
		expected    ResourceType
	}{
		{
			name:        "Uppercase extension",
			url:         "https://example.com/page.HTML",
			contentType: "",
			expected:    ResourceTypeHTML,
		},
		{
			name:        "Mixed case extension",
			url:         "https://example.com/photo.JpG",
			contentType: "",
			expected:    ResourceTypeImage,
		},
		{
			name:        "Uppercase content-type",
			url:         "https://example.com/page",
			contentType: "TEXT/HTML",
			expected:    ResourceTypeHTML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectResourceType(tt.url, tt.contentType)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestResourceType_String(t *testing.T) {
	tests := []struct {
		rt       ResourceType
		expected string
	}{
		{ResourceTypeHTML, "HTML"},
		{ResourceTypeCSS, "CSS"},
		{ResourceTypeJavaScript, "JavaScript"},
		{ResourceTypeImage, "Image"},
		{ResourceTypeFont, "Font"},
		{ResourceTypeOther, "Other"},
		{ResourceTypeUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.rt.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
