package urlutil

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

// ResourceType represents the type of a web resource
type ResourceType int

const (
	ResourceUnknown ResourceType = iota
	ResourceHTML
	ResourceCSS
	ResourceJS
	ResourceImage
	ResourceFont
	ResourceOther
)

// String returns a string representation of the resource type
func (rt ResourceType) String() string {
	switch rt {
	case ResourceHTML:
		return "HTML"
	case ResourceCSS:
		return "CSS"
	case ResourceJS:
		return "JavaScript"
	case ResourceImage:
		return "Image"
	case ResourceFont:
		return "Font"
	case ResourceOther:
		return "Other"
	default:
		return "Unknown"
	}
}

// Filter handles URL filtering and resource type detection
type Filter struct {
	rootDomain string
	rootPath   string
	whitelist  map[string]bool
}

// NewFilter creates a new URL filter with the given root URL and domain whitelist
func NewFilter(rootURL string, whitelistDomains []string) *Filter {
	u, err := url.Parse(rootURL)
	if err != nil {
		// For invalid URLs, create a filter that will reject everything
		return &Filter{
			rootDomain: "",
			rootPath:   "",
			whitelist:  make(map[string]bool),
		}
	}

	// Extract the root path (directory containing the root URL)
	rootPath := path.Dir(u.Path)
	if rootPath == "." {
		rootPath = "/"
	}

	// Create whitelist map
	whitelist := make(map[string]bool)
	for _, domain := range whitelistDomains {
		whitelist[strings.ToLower(domain)] = true
	}

	return &Filter{
		rootDomain: strings.ToLower(u.Host),
		rootPath:   rootPath,
		whitelist:  whitelist,
	}
}

// IsAllowed checks if a URL is allowed to be scraped based on the root domain and whitelist
func (f *Filter) IsAllowed(rawURL string) (bool, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false, fmt.Errorf("failed to parse URL %q: %w", rawURL, err)
	}

	if !u.IsAbs() {
		return false, fmt.Errorf("URL %q is relative", rawURL)
	}

	domain := strings.ToLower(u.Host)

	// Check if it's the root domain
	if domain == f.rootDomain {
		// Check if the path has the same prefix as root path
		return strings.HasPrefix(u.Path, f.rootPath), nil
	}

	// Check if it's in the whitelist
	if f.whitelist[domain] {
		return true, nil
	}

	return false, nil
}

// DetectResourceType determines the resource type based on URL and content type
// This is a standalone function that can be used without a Filter instance
func DetectResourceType(rawURL, contentType string) ResourceType {
	// First try to determine by Content-Type header if provided
	if contentType != "" {
		ct := strings.ToLower(strings.Split(contentType, ";")[0])
		ct = strings.TrimSpace(ct)

		switch {
		case strings.Contains(ct, "text/html"):
			return ResourceHTML
		case strings.Contains(ct, "text/css"):
			return ResourceCSS
		case strings.Contains(ct, "javascript"), strings.Contains(ct, "application/javascript"),
			strings.Contains(ct, "application/x-javascript"), strings.Contains(ct, "text/javascript"):
			return ResourceJS
		case strings.HasPrefix(ct, "image/"):
			return ResourceImage
		case strings.Contains(ct, "font"), strings.Contains(ct, "woff"), strings.Contains(ct, "ttf"):
			return ResourceFont
		}
	}

	// Fall back to extension-based detection
	u, err := url.Parse(rawURL)
	if err != nil {
		return ResourceOther
	}

	// Extract extension from path (ignoring query and fragment)
	ext := strings.ToLower(path.Ext(u.Path))

	switch ext {
	case ".html", ".htm":
		return ResourceHTML
	case ".css":
		return ResourceCSS
	case ".js", ".mjs":
		return ResourceJS
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".ico", ".bmp":
		return ResourceImage
	case ".woff", ".woff2", ".ttf", ".otf", ".eot":
		return ResourceFont
	case ".pdf", ".zip", ".tar", ".gz":
		return ResourceOther
	case "":
		// No extension - assume HTML (common for index pages)
		return ResourceHTML
	default:
		return ResourceUnknown
	}
}

// GetResourceType is a convenience method that calls DetectResourceType
// Kept for backward compatibility
func (f *Filter) GetResourceType(rawURL, contentType string) ResourceType {
	return DetectResourceType(rawURL, contentType)
}

// GetWeight returns the priority weight for a resource type
// Higher weight = higher priority
func (rt ResourceType) GetWeight() int {
	switch rt {
	case ResourceHTML:
		return 100
	case ResourceCSS:
		return 75
	case ResourceJS:
		return 50
	case ResourceImage:
		return 25
	case ResourceFont:
		return 20
	case ResourceOther:
		return 10
	case ResourceUnknown:
		return 5
	default:
		return 0
	}
}
