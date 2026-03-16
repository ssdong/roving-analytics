package models

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"roving_web/spamlist"
	"roving_web/utils"
	"strings"

	"time"
)

type DropReason string
type RequestBody struct {
	EventName string `json:"e"` // eventName (e.g. "pageview")
	RawUrl    string `json:"u"` // current page URL
	Referrer  string `json:"r"` // referrer
	Timestamp uint64 `json:"t"` // timestamp when event occurred(Date.now() in JS)
	Salt      string `json:"s"` // browser-generated salt
}

type SanitizedRequest struct {
	EventName string
	Timestamp uint64
	IP        string
	UserAgent string
	Uri       *Url
	Referrer  string
	Hostname  string
	Pathname    string
	QueryParams map[string][]string
	Salt        string
}

const (
	// maximum event name length is 200 bytes
	maxEventNameLength = 200
	// Define a threshold for how much time difference is acceptable.
	// For example, allow a 24-hour difference between server time and the provided timestamp.
	threshold = 24 * time.Hour
)

const (
	Bot          DropReason = "bot"
	SpamReferrer DropReason = "spam_referrer"
	Blocked      DropReason = "blocked" // Blocked by rate limiting
	Invalid      DropReason = "invalid"
)

func (d DropReason) Error() string {
    return string(d)
}

func sanitizeEventName(eventName string) (string, error) {
	if eventName == "" {
		return "", Invalid
	}

	if len(eventName) > maxEventNameLength {
		return "", Invalid
	}

	return eventName, nil
}

// ValidateTimestamp checks if the provided timestamp is within a reasonable range of the current time.
func sanitizeTimestamp(host string, timestamp uint64) error {
	// Convert the current time to milliseconds since the Unix epoch.
	currentTimeMillis := uint64(time.Now().UnixNano() / int64(time.Millisecond))

	// Calculate the difference in milliseconds between the current time and the provided timestamp.
	var diff uint64
	if timestamp > currentTimeMillis {
		diff = timestamp - currentTimeMillis
	} else {
		diff = currentTimeMillis - timestamp
	}

	// Convert the difference to a time.Duration for comparison.
	// TODO: this needs to be tested
	diffDuration := time.Duration(diff) * time.Millisecond

	// Check if the difference is greater than the threshold.
	if diffDuration > threshold {
		log.Println("Provided timestamp is outside the acceptable range from %s", host)
		return Invalid
	}

	return nil
}

func setQueryParams(event *SanitizedRequest) {
	if event.Uri == nil {
		return
	}

	// Making a copy to avoid modifying the original data
	sanitizedQueryParams := make(map[string][]string)
	for key, values := range event.Uri.QueryParams {
		for _, value := range values {
			// converting any '%XX' escape sequences back to the actual characters they represent.
			// For example:
			// Input:  "Hello%20World%21"
			// Output: "Hello World!"
			decodedValue, err := url.QueryUnescape(value)
			if err != nil {
				// Skip this value if it can't be decoded
				continue
			}
			sanitizedQueryParams[key] = append(sanitizedQueryParams[key], decodedValue)
		}
	}

	event.QueryParams = sanitizedQueryParams
}

func sanitizeReferrer(rawReferrer string) (string, error) {
	len := len(rawReferrer)

	if len > MaximumUrlLength {
		rawReferrer = rawReferrer[:MaximumUrlLength]
	}

	parsedReferrer, err := url.Parse(rawReferrer)
	if err != nil {
		return "", Invalid
	}

	if spamlist.GetInstance().IsSpam(SanitizeHostname(parsedReferrer.Host)) {
		return "", SpamReferrer
	}

	// We're only interested in http and https referrers
	// Any other scheme (like "ftp", "mailto", etc.) is not considered a valid referrer
	if parsedReferrer.Scheme != "http" && parsedReferrer.Scheme != "https" {
		return "", Invalid
	}

	return rawReferrer, nil
}

func SanitizeRequest(r *http.Request) (*SanitizedRequest, error) {
	// Initialize an empty SanitizedRequest
	sanitized := &SanitizedRequest{}

	var requestBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&requestBody)

	if err != nil {
		log.Println("Error decoding request body: ", err)
		return nil, Invalid
	}

	// 1. Get event name
	eventName, err := sanitizeEventName(requestBody.EventName)
	if err != nil {
		return nil, err
	}
	sanitized.EventName = eventName

	// 2. Grab IP
	sanitized.IP = utils.GetRemoteIP(r)

	// 3. Get user-agent
	agent := parser.Parse(r.UserAgent())
	if agent.IsBot() {
		return nil, Bot
	}
	sanitized.UserAgent = r.UserAgent()

	// 4. Get URI
	rawURL := strings.TrimSpace(requestBody.RawUrl)
	uri, err := ParseURI(rawURL)
	if err != nil || uri == nil {
		return nil, Invalid
	}
	sanitized.Uri = uri

	// 5. Get timestamp
	err = sanitizeTimestamp(uri.Host, requestBody.Timestamp)
	if err != nil {
		return nil, err
	}
	sanitized.Timestamp = requestBody.Timestamp

	// 6. Get referrer
	referrer, err := sanitizeReferrer(requestBody.Referrer)
	if err != nil {
		return nil, err
	}

	sanitized.Referrer = referrer

	// 7. Get hostname
	sanitized.Hostname = SanitizeHostname(sanitized.Uri.Host)

	// 8. Get pathname
	sanitized.Pathname = GetPathname(sanitized.Uri)

	// 9. Get query params
	setQueryParams(sanitized)

	// 10. Get salt
	sanitized.Salt = requestBody.Salt

	return sanitized, nil
}
