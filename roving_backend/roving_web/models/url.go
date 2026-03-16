package models

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

// The 2000-character guideline for URLs, especially in the context of maximum length,
// predominantly stems from the limitations imposed by Internet Explorer (IE).
// Historically, Internet Explorer had a URL length limit of 2,083 characters.
const MaximumUrlLength = 2000

const ignoreDataScheme = "data"

type Url struct {
	Scheme      string
	Host        string
	Path        string
	Hash        string
	QueryParams map[string][]string
}

func ParseURI(rawUrl string) (*Url, error) {
	len := len(rawUrl)

	if len > MaximumUrlLength {
		return nil, errors.New("URL exceeds the maximum allowed length")
	}

	// Parse the URL
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	// Ignore data scheme
	if parsedUrl.Scheme == ignoreDataScheme {
		return nil, errors.New("data scheme is not supported")
	}

	url := &Url{
		Scheme:      parsedUrl.Scheme,
		Host:        parsedUrl.Host,
		Path:        parsedUrl.Path,
		Hash:        parsedUrl.Fragment,
		QueryParams: parsedUrl.Query(),
	}

	return url, nil
}

// Parsing URLs like https://www.example.com:443/path/to/resource?query=value#fragment
// will result in host = www.example.com:443 and we'd like to remove the port number.
func SanitizeHostname(rawHostname string) string {
	// Remove any leading or trailing whitespace
	rawHostname = strings.TrimSpace(rawHostname)

	// Remove port number if it exists
	hostname, _, err := net.SplitHostPort(rawHostname)

	if err != nil {
		// if no port found, reassign raw hostname to hostname
		hostname = rawHostname
	}

	hostname = strings.Trim(hostname, "[]")
	hostname = strings.TrimPrefix(hostname, "www.")

	return hostname
}

// GetPathname reconstructs the pathname from the given URL structure.
// It appends the hash fragment to the pathname.
//
// Example:
// Given a URL like "https://example.com/some/path?query=value#fragment"
// the function will return "/some/path#fragment"
//
// Parameters:
// - parsedUrl: the already parsed URL structure
//
// Returns:
// - the reconstructed pathname
func GetPathname(parsedUrl *Url) string {
	pathname := parsedUrl.Path
	if pathname == "" {
		pathname = "/"
	}

	// Decode and trim the pathname
	decodedPathname, err := url.PathUnescape(pathname)
	if err != nil {
		decodedPathname = pathname // If decoding fails, use the original
	}
	decodedPathname = strings.TrimSpace(decodedPathname)

	// Append hash only if it exists
	if parsedUrl.Hash != "" {
		decodedPathname += "#" + parsedUrl.Hash
	}

	return decodedPathname
}
