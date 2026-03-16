package models

import (
	"encoding/hex"
	"log"
	"net/netip"
	"roving_web/config"
	"roving_web/models/site"
	"roving_web/repositories"
	"sync"

	"github.com/medama-io/go-useragent"
	"github.com/minio/highwayhash"
	"github.com/oschwald/geoip2-golang/v2"
	"github.com/trojanc/iso3166"
)

type LowCardinalityString string
type SiteIdCache struct {
	mu      sync.RWMutex
	siteIds map[string]uint32
}

var (
	parser            *useragent.Parser
	globalJourneyKey  []byte
	geoliteDBReader   *geoip2.Reader
	modelsOnce        sync.Once
)

type Event struct {
	Domain               string
	Site                 *site.Site             // Event is related to track a specific site
	ClickhouseEventAttrs map[string]interface{} // Assuming that the attrs is a map of string to any type
	ClickhouseEvent      *ClickhouseEvent
	Salt                 string
}

type ClickhouseEvent struct {
	EventName              string // Can be LowCardinalityString in Clickhouse
	SiteId                 uint32
	Hostname               string // Can be LowCardinalityString in Clickhouse
	Pathname               string
	JourneyId              string
	Timestamp              uint64
	Referrer               string
	ReferrerSource         string
	CountryCode            string
	Subdivision1Code       string // Can be LowCardinalityString in Clickhouse
	Subdivision2Code       string // Can be LowCardinalityString in Clickhouse
	CityGeonameId          uint32
	OperatingSystem        string // Can be LowCardinalityString in Clickhouse
	OperatingSystemVersion string // Can be LowCardinalityString in Clickhouse
	Browser                string // Can be LowCardinalityString in Clickhouse
	BrowserVersion         string // Can be LowCardinalityString in Clickhouse
	UtmSource              string
	UtmMedium              string
	UtmCampaign            string
	UtmContent             string
	UtmTerm                string
	MetaData               map[string]string
}

func InitializeModels() {
    modelsOnce.Do(func() {
        parser = useragent.NewParser()

        reader, err := geoip2.Open(config.AppConfig.GeoliteDBPath)
        if err != nil {
            log.Fatalf("Failed to read GeoLite2 mmdb file: %v", err)
        }
        geoliteDBReader = reader

        globalJourneyKey = []byte(config.AppConfig.RovingJourneyKey)
        if len(globalJourneyKey) != 32 {
            log.Fatal("ROVING_JOURNEY_KEY must be exactly 32 bytes")
        }
    })
}

func (c *SiteIdCache) GetSiteId(hostname string) (uint32, bool) {
	c.mu.RLock()
	id, ok := c.siteIds[hostname]
	c.mu.RUnlock()
	return id, ok
}

func (c *SiteIdCache) SetSiteId(hostname string, id uint32) {
	c.mu.Lock()
	c.siteIds[hostname] = id
	c.mu.Unlock()
}

func (e *Event) PopulateUserAgent(request *SanitizedRequest) *Event {
	agent := parser.Parse(request.UserAgent)

	if e.ClickhouseEventAttrs == nil {
		e.ClickhouseEventAttrs = make(map[string]interface{})
	}

	e.ClickhouseEventAttrs["OperatingSystem"] = agent.OS()
	e.ClickhouseEventAttrs["OperatingSystemVersion"] = "" // TODO: the lib does not support OS version yet
	e.ClickhouseEventAttrs["Browser"] = agent.Browser()
	e.ClickhouseEventAttrs["BrowserVersion"] = agent.BrowserVersion()

	return e
}

func (e *Event) PopulateDomainInfo(request *SanitizedRequest) *Event {
	if e.ClickhouseEventAttrs == nil {
		e.ClickhouseEventAttrs = make(map[string]interface{})
	}

	e.ClickhouseEventAttrs["Domain"] = request.Hostname
	e.ClickhouseEventAttrs["Hostname"] = request.Hostname
	e.ClickhouseEventAttrs["Pathname"] = request.Pathname
	e.ClickhouseEventAttrs["Referrer"] = request.Referrer
	e.ClickhouseEventAttrs["EventName"] = request.EventName
	e.ClickhouseEventAttrs["Timestamp"] = request.Timestamp

	siteId, err := repositories.GetSiteRepository().GetSiteIdByHostname(request.Hostname)

	if err != nil {
		return e
	}
	
	e.ClickhouseEventAttrs["SiteId"] = siteId

	return e
}

func (e *Event) PopulateReferrer(request *SanitizedRequest) *Event {
	if e.ClickhouseEventAttrs == nil {
		e.ClickhouseEventAttrs = make(map[string]interface{})
	}

	referrerUri, err := ParseURI(request.Referrer)
	if err != nil {
		return e
	}

	// Reduce "Internal Noise" if user navigates between different pages on the same hostname
	if SanitizeHostname(referrerUri.Host) != SanitizeHostname(request.Uri.Host) {
		// This is a referrer from an external site
		e.ClickhouseEventAttrs["Referrer"] = SanitizeHostname(referrerUri.Host)

		utmSource, utmSourceExist := request.QueryParams["utm_source"]
		source, sourceExist := request.QueryParams["source"]
		ref, refExist := request.QueryParams["ref"]

		if utmSourceExist && len(utmSource) > 0 {
			e.ClickhouseEventAttrs["ReferrerSource"] = utmSource[0]
		} else if sourceExist && len(source) > 0 {
			e.ClickhouseEventAttrs["ReferrerSource"] = source[0]
		} else if refExist && len(ref) > 0 {
			e.ClickhouseEventAttrs["ReferrerSource"] = ref[0]
		} else {
			e.ClickhouseEventAttrs["ReferrerSource"] = SanitizeHostname(referrerUri.Host)
		}
	}

	return e
}

func (e *Event) PopulateUtmParams(request *SanitizedRequest) *Event {
	if e.ClickhouseEventAttrs == nil {
		e.ClickhouseEventAttrs = make(map[string]interface{})
	}

	params := map[string]string{
		"utm_source":   "UtmSource",
		"utm_medium":   "UtmMedium",
		"utm_campaign": "UtmCampaign",
		"utm_content":  "UtmContent",
		"utm_term":     "UtmTerm",
	}

	for queryKey, attrKey := range params {
		if values, ok := request.QueryParams[queryKey]; ok && len(values) > 0 {
			e.ClickhouseEventAttrs[attrKey] = values[0]
		}
	}

	return e
}

func (e *Event) PopulateLocation(request *SanitizedRequest) *Event {
	if e.ClickhouseEventAttrs == nil {
		e.ClickhouseEventAttrs = make(map[string]interface{})
	}

	ip, err := netip.ParseAddr(request.IP)
	if err != nil {
		log.Printf("Invalid IP address %s: %v", request.IP, err)
		return e
	}

	record, err := geoliteDBReader.City(ip)
	if err != nil {
		log.Printf("GeoIP lookup failed for %s: %v", request.IP, err)
		return e
	}

	if !record.HasData() {
		return e
	}

	alpha2 := record.Country.ISOCode
	country, err := iso3166.FromAlpha2(alpha2)

	if err == nil {
    	e.ClickhouseEventAttrs["CountryCode"] = country.Alpha3() 
	} else {
    	e.ClickhouseEventAttrs["CountryCode"] = alpha2 
	}

	if len(record.Subdivisions) > 0 {
		e.ClickhouseEventAttrs["Subdivision1Code"] = record.Subdivisions[0].ISOCode
	}

	if len(record.Subdivisions) > 1 {
		e.ClickhouseEventAttrs["Subdivision2Code"] = record.Subdivisions[1].ISOCode
	}

	e.ClickhouseEventAttrs["CityGeonameId"] = uint32(record.City.GeoNameID)

	return e
}

func (e *Event) PopulateJourneyId(request *SanitizedRequest) *Event {
	if e.ClickhouseEventAttrs == nil {
		e.ClickhouseEventAttrs = make(map[string]interface{})
	}

	input := request.UserAgent + request.IP + request.Hostname + request.Salt

	hash := highwayhash.Sum128([]byte(input), []byte(globalJourneyKey))

	e.ClickhouseEventAttrs["JourneyId"] = hex.EncodeToString(hash[:])

	return e
}

func (e *Event) ValidateClickHouseEvent() *Event {
	if e.ClickhouseEvent == nil {
		e.ClickhouseEvent = &ClickhouseEvent{}
	}

	if domain, ok := e.ClickhouseEventAttrs["Domain"].(string); ok {
		e.Domain = domain
	}

	// TODO: Do we need error handling if the value is not present?
	if eventName, ok := e.ClickhouseEventAttrs["EventName"].(string); ok {
		e.ClickhouseEvent.EventName = eventName
	}

	if siteId, ok := e.ClickhouseEventAttrs["SiteId"].(uint32); ok {
		e.ClickhouseEvent.SiteId = siteId
	}

	if hostName, ok := e.ClickhouseEventAttrs["Hostname"].(string); ok {
		e.ClickhouseEvent.Hostname = hostName
	}

	if pathName, ok := e.ClickhouseEventAttrs["Pathname"].(string); ok {
		e.ClickhouseEvent.Pathname = pathName
	}

	if journeyId, ok := e.ClickhouseEventAttrs["JourneyId"].(string); ok {
		e.ClickhouseEvent.JourneyId = journeyId
	}

	if timestamp, ok := e.ClickhouseEventAttrs["Timestamp"].(uint64); ok {
		e.ClickhouseEvent.Timestamp = timestamp
	}

	if referrer, ok := e.ClickhouseEventAttrs["Referrer"].(string); ok {
		e.ClickhouseEvent.Referrer = referrer
	}

	if referrerSource, ok := e.ClickhouseEventAttrs["ReferrerSource"].(string); ok {
		e.ClickhouseEvent.ReferrerSource = referrerSource
	}

	if countryCode, ok := e.ClickhouseEventAttrs["CountryCode"].(string); ok {
		e.ClickhouseEvent.CountryCode = countryCode
	}

	if subdivision1Code, ok := e.ClickhouseEventAttrs["Subdivision1Code"].(string); ok {
		e.ClickhouseEvent.Subdivision1Code = subdivision1Code
	}

	if subdivision2Code, ok := e.ClickhouseEventAttrs["Subdivision2Code"].(string); ok {
		e.ClickhouseEvent.Subdivision2Code = subdivision2Code
	}

	if cityGeonameId, ok := e.ClickhouseEventAttrs["CityGeonameId"].(uint32); ok {
		e.ClickhouseEvent.CityGeonameId = cityGeonameId
	}

	if operatingSystem, ok := e.ClickhouseEventAttrs["OperatingSystem"].(string); ok {
		e.ClickhouseEvent.OperatingSystem = operatingSystem
	}

	if operatingSystemVersion, ok := e.ClickhouseEventAttrs["OperatingSystemVersion"].(string); ok {
		e.ClickhouseEvent.OperatingSystemVersion = operatingSystemVersion
	}

	if browser, ok := e.ClickhouseEventAttrs["Browser"].(string); ok {
		e.ClickhouseEvent.Browser = browser
	}

	if browserVersion, ok := e.ClickhouseEventAttrs["BrowserVersion"].(string); ok {
		e.ClickhouseEvent.BrowserVersion = browserVersion
	}

	if utmSource, ok := e.ClickhouseEventAttrs["UtmSource"].(string); ok {
		e.ClickhouseEvent.UtmSource = utmSource
	}

	if utmMedium, ok := e.ClickhouseEventAttrs["UtmMedium"].(string); ok {
		e.ClickhouseEvent.UtmMedium = utmMedium
	}

	if utmCampaign, ok := e.ClickhouseEventAttrs["UtmCampaign"].(string); ok {
		e.ClickhouseEvent.UtmCampaign = utmCampaign
	}

	if utmContent, ok := e.ClickhouseEventAttrs["UtmContent"].(string); ok {
		e.ClickhouseEvent.UtmContent = utmContent
	}

	if utmTerm, ok := e.ClickhouseEventAttrs["UtmTerm"].(string); ok {
		e.ClickhouseEvent.UtmTerm = utmTerm
	}

	if metaData, ok := e.ClickhouseEventAttrs["MetaData"].(map[string]string); ok {
		e.ClickhouseEvent.MetaData = metaData
	}

	return e
}
