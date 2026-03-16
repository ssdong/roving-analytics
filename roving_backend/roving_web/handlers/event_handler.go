package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"roving_web/db"
	"roving_web/models"
)

// Consider using https://github.com/nikepan/clickhouse-bulk in the future
// https://clickhouse.com/docs/en/concepts/why-clickhouse-is-so-fast#performance-when-inserting-data
const clickhouseInsertBufferSize = 1500
const flushThreshold = int(0.75 * float64(clickhouseInsertBufferSize))

const requestBufferSize = 1000000

var requestChannel = make(chan *http.Request, requestBufferSize)

const numWorkersPerEventChannel = 10
const numWorkersPerRequestChannel = 10

var channels = make(map[rune]chan models.Event)

func init() {
	// Create channels for alphabetic characters
	for i := 'a'; i <= 'z'; i++ {
		ch := make(chan models.Event, 10000)
		channels[i] = ch
		for j := 0; j < numWorkersPerEventChannel; j++ {
			go handleEventChannel(ch)
		}
	}

	// Create a general channel for non-alphabetic starting characters
	generalChannel := make(chan models.Event, 10000)
	channels['*'] = generalChannel
	for j := 0; j < numWorkersPerEventChannel; j++ {
		go handleEventChannel(generalChannel)
	}

	for i := 0; i < numWorkersPerRequestChannel; i++ {
		go processRequests()
	}
}

func processRequests() {
	for r := range requestChannel {
		processRequest(r)
	}
}

func processRequest(r *http.Request) {
	sanitizedRequest, err := models.SanitizeRequest(r)

	if err != nil {
        // Check if the error is an "Expected Drop" (Bot, Spam, Invalid)
        if _, ok := err.(models.DropReason); ok {
            return
        }

        // Other "Real" System Error (e.g., JSON malformed, unexpected panic, etc.) that isn't a DropReason
        fmt.Printf("System error sanitizing request: %v\n", err)
        return
    }

	event := models.Event{}

	event.
		PopulateDomainInfo(sanitizedRequest).
		PopulateUserAgent(sanitizedRequest).
		PopulateReferrer(sanitizedRequest).
		PopulateJourneyId(sanitizedRequest).
		PopulateUtmParams(sanitizedRequest).
		PopulateLocation(sanitizedRequest).
		ValidateClickHouseEvent()

	firstLetter := rune(event.Domain[0])
	ch, ok := channels[firstLetter]
	if !ok {
		fmt.Printf("No channel found for domain starting with %s\n. Using default one", string(firstLetter))
		ch = channels['*'] // use the general channel
	}

	select {
	case ch <- event:
	default:
		// TODO: Instead of dropping, consider sending it to Kafka and deal later?
		fmt.Printf("Drop event for domain starting with %c - buffer overflow", firstLetter)
	}
}

func handleEventChannel(ch chan models.Event) {
	var buffer []models.Event

	// Create a ticker that triggers every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		// The channel ch for incoming events, which get appended to the buffer.
		// If the buffer reaches the threshold, it's flushed.
		case e := <-ch:
			buffer = append(buffer, e)
			if len(buffer) >= flushThreshold {
				flushToClickHouse(buffer)
				buffer = []models.Event{}
			}
			// The ticker, which sends a signal every 10 seconds. When this signal is received,
			// the buffer is flushed regardless of its fill level, but only if it's not empty.
		case <-ticker.C:
			if len(buffer) > 0 {
				flushToClickHouse(buffer)
				buffer = []models.Event{}
			}
		}
	}
}

func flushToClickHouse(events []models.Event) error {
	clickhouseConn, err := db.GetConnection()

	if clickhouseConn == nil && err != nil {
		// TODO: Dont panic here but start writing to Kafka instead to preserve events? 
		log.Default().Fatal("No ClickHouse connection!")
	}
	ctx := context.Background()

	// Preparing the batch insert for the roving.web_traffic_event table
	batch, err := clickhouseConn.PrepareBatch(ctx, "INSERT INTO roving.web_traffic_event (EventName, SiteId, HostName, PathName, JourneyId, Timestamp, Referrer, ReferrerSource, CountryCode, Subdivision1Code, Subdivision2Code, CityGeonameId, OperatingSystem, OperatingSystemVersion, Browser, BrowserVersion, UtmSource, UtmMedium, UtmCampaign, UtmContent, UtmTerm)")
	if err != nil {
		return err
	}

	// Looping through the events to append them to the batch
	for _, event := range events {
		chEvent := event.ClickhouseEvent // making the code more readable by creating an alias

		seconds := int64(chEvent.Timestamp / 1000)
		nanoseconds := int64((chEvent.Timestamp % 1000) * 1e6) // converting milliseconds to nanoseconds
		timestamp := time.Unix(seconds, nanoseconds)

		err := batch.Append(
			chEvent.EventName,
			chEvent.SiteId,
			chEvent.Hostname,
			chEvent.Pathname,
			chEvent.JourneyId,
			timestamp,
			chEvent.Referrer,
			chEvent.ReferrerSource,
			chEvent.CountryCode,
			chEvent.Subdivision1Code,
			chEvent.Subdivision2Code,
			chEvent.CityGeonameId,
			chEvent.OperatingSystem,
			chEvent.OperatingSystemVersion,
			chEvent.Browser,
			chEvent.BrowserVersion,
			chEvent.UtmSource,
			chEvent.UtmMedium,
			chEvent.UtmCampaign,
			chEvent.UtmContent,
			chEvent.UtmTerm,
		)

		if err != nil {
			return err
		}
	}

	fmt.Printf("Ready to flush %d events\n", len(events))

	// Finally, send the batch to ClickHouse
	return batch.Send()
}

func EventHandler(w http.ResponseWriter, r *http.Request) {
	// Clone the request
	reqClone := r.Clone(context.Background())

	// Read the entire request body into a buffer first from the cloned request
	bodyBytes, err := io.ReadAll(reqClone.Body)
	if err != nil {
		// handle the error
		fmt.Println("Failed to read request body: ", err)
		return
	}

	// Replace the request body with a new ReadCloser that reads from the buffer
	reqClone.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	select {
	case requestChannel <- reqClone:
	default:
		fmt.Println("Server too busy", http.StatusServiceUnavailable)
	}

	// Immediately respond with 200 OK
	w.WriteHeader(http.StatusOK)
}
