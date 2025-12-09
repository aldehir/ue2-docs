package fetcher

import (
	"path"
	"strings"
)

// DetectResourceType determines the resource type from URL and Content-Type header
func DetectResourceType(url, contentType string) ResourceType {
	// First, try to detect by Content-Type header
	if contentType != "" {
		ct := strings.ToLower(strings.Split(contentType, ";")[0])
		ct = strings.TrimSpace(ct)

		switch {
		case strings.Contains(ct, "text/html"):
			return ResourceTypeHTML
		case strings.Contains(ct, "text/css"):
			return ResourceTypeCSS
		case strings.Contains(ct, "javascript"), strings.Contains(ct, "application/javascript"),
			strings.Contains(ct, "application/x-javascript"), strings.Contains(ct, "text/javascript"):
			return ResourceTypeJavaScript
		case strings.HasPrefix(ct, "image/"):
			return ResourceTypeImage
		case strings.Contains(ct, "font"), strings.Contains(ct, "woff"), strings.Contains(ct, "ttf"):
			return ResourceTypeFont
		}
	}

	// Fallback to URL extension
	ext := strings.ToLower(path.Ext(url))

	switch ext {
	case ".html", ".htm":
		return ResourceTypeHTML
	case ".css":
		return ResourceTypeCSS
	case ".js", ".mjs":
		return ResourceTypeJavaScript
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".ico", ".bmp":
		return ResourceTypeImage
	case ".woff", ".woff2", ".ttf", ".otf", ".eot":
		return ResourceTypeFont
	case ".pdf", ".zip", ".tar", ".gz":
		return ResourceTypeOther
	default:
		return ResourceTypeUnknown
	}
}

// String returns a string representation of the ResourceType
func (rt ResourceType) String() string {
	switch rt {
	case ResourceTypeHTML:
		return "HTML"
	case ResourceTypeCSS:
		return "CSS"
	case ResourceTypeJavaScript:
		return "JavaScript"
	case ResourceTypeImage:
		return "Image"
	case ResourceTypeFont:
		return "Font"
	case ResourceTypeOther:
		return "Other"
	default:
		return "Unknown"
	}
}
