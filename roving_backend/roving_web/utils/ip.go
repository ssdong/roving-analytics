package utils

import (
	"net"
	"net/http"
	"regexp"
	"strings"
)

// Update to other CDN headers if not using CF in production
const cloudfareIPHeader = "cf-connecting-ip"
const xForwardedForHeader = "x-forwarded-for"
const forwardedHeader = "forwarded"

// GetRemoteIP extracts the IP address from the given http.Request.
func GetRemoteIP(r *http.Request) string {
	cfConnectingIP := getFirstHeader(r, cloudfareIPHeader)
	xForwardedFor := getFirstHeader(r, xForwardedForHeader)
	forwarded := getFirstHeader(r, forwardedHeader)

	if cfConnectingIP != "" {
		return sanitizeIP(cfConnectingIP)
	} else if xForwardedFor != "" {
		return parseForwardedFor(xForwardedFor)
	} else if forwarded != "" {
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Forwarded
		forRegex := regexp.MustCompile(`for=([^;,]+)`)
		matches := forRegex.FindStringSubmatch(forwarded)
		// When we use FindStringSubmatch, it returns a slice where:
		// The first element (matches[0]) is the entire portion of the input string that matches the regex pattern.
		// The subsequent elements (matches[1], matches[2], etc.) correspond to the capturing groups in the regex pattern.

		// In this context, matches[0] would contain the entire matched string, including the "for=" prefix, while matches[1]
		// would only contain the content of the capturing group, which is the actual IP address or proxy identifier you're interested in.

		// So, in this specific scenario:
		// matches[0] would be something like for=192.168.0.1.
		// matches[1] would be 192.168.0.1. (First Submatch in case there are multiple "hops", i.e. multiple proxy servers)
		if len(matches) > 1 {
			// Pv6 address is enclosed in both square brackets (to differentiate the address from the port) and double quotes.
			// The double quotes are used to signify that the entire value, including the square brackets, represents a single entity.

			// Example:
			// Forwarded: For="[2001:db8:cafe::17]:4711"
			return sanitizeIP(strings.Trim(matches[1], "\""))
		}
	}
	// If more specific headers like X-Forwarded-For are missing, the remote address can still be used as a
	// rudimentary way to identify the user.
	lastPossibleIpLookUp := sanitizeIP(r.RemoteAddr)
	return lastPossibleIpLookUp
}

func getFirstHeader(r *http.Request, key string) string {
	headerValue := r.Header.Get(key)
	if headerValue != "" {
		return headerValue
	}
	return ""
}

// sanitizeIP removes the port from both IPv4 and IPv6 addresses, if present.
// It also removes surrounding square brackets from an IPv6 address, if present.
//
// Examples:
//
//	sanitizeIP("192.168.1.1:8080")              => "192.168.1.1"
//	sanitizeIP("[2001:db8::ff00:42:8329]:8080") => "2001:db8::ff00:42:8329"
//	sanitizeIP("192.168.1.1")                   => "192.168.1.1"
//	sanitizeIP("[2001:db8::ff00:42:8329]")      => "2001:db8::ff00:42:8329"
func sanitizeIP(ipAndPort string) string {
	host, _, err := net.SplitHostPort(ipAndPort)

	if err != nil {
		// If it fails, there was no port (or it's a malformed string).
		// Just use the original string, but trim IPv6 brackets just in case.
		host = ipAndPort
	}

	// Clean up any remaining IPv6 brackets
	return strings.Trim(host, "[]")
}

// parseForwardedFor extracts the client's IP address from the X-Forwarded-For header.
// The header can contain a list of proxy addresses; this function returns the first address,
// which is typically the original client's IP. If the header contains multiple addresses,
// they are separated by commas. The function also sanitizes the IP to remove any potential
// port information or unwanted characters.
func parseForwardedFor(header string) string {
	addresses := strings.Split(header, ",")
	if len(addresses) > 0 {
		return sanitizeIP(strings.TrimSpace(addresses[0]))
	}
	return ""
}
