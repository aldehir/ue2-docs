package urlutil

import (
	"fmt"
	"net/url"
	"strings"
)

// Normalize normalizes a URL by:
// - Resolving relative URLs against a base URL (if provided)
// - Lowercasing the scheme and domain
// - Removing query strings
// - Preserving fragments (#anchors)
// - Removing default ports (80 for http, 443 for https)
// - Removing trailing slashes (except for root paths)
//
// Parameters:
//   - rawURL: The URL to normalize
//   - baseURL: Optional base URL for resolving relative URLs (empty string if not needed)
//
// Returns the normalized URL string or an error if the URL is invalid.
func Normalize(rawURL, baseURL string) (string, error) {
	var u *url.URL
	var err error

	// Parse the input URL
	u, err = url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL %q: %w", rawURL, err)
	}

	// If we have a base URL and the input is relative, resolve it
	if baseURL != "" && !u.IsAbs() {
		base, err := url.Parse(baseURL)
		if err != nil {
			return "", fmt.Errorf("failed to parse base URL %q: %w", baseURL, err)
		}
		u = base.ResolveReference(u)
	}

	// Ensure we have an absolute URL at this point
	if !u.IsAbs() {
		return "", fmt.Errorf("URL %q is relative and no base URL provided", rawURL)
	}

	// Lowercase scheme and host
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// Remove default ports
	if u.Scheme == "http" && strings.HasSuffix(u.Host, ":80") {
		u.Host = strings.TrimSuffix(u.Host, ":80")
	} else if u.Scheme == "https" && strings.HasSuffix(u.Host, ":443") {
		u.Host = strings.TrimSuffix(u.Host, ":443")
	}

	// Remove query string
	u.RawQuery = ""
	u.ForceQuery = false

	// Remove trailing slash from path (but not for root)
	if u.Path != "/" && strings.HasSuffix(u.Path, "/") {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}

	return u.String(), nil
}
