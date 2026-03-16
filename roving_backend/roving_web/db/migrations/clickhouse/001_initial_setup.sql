CREATE DATABASE IF NOT EXISTS roving

CREATE TABLE roving.hostname_to_site_id
(
	`Hostname` String,
	`SiteId` UInt32,
) ENGINE = MergeTree()
ORDER BY (Hostname)

CREATE TABLE roving.web_traffic_event
(
    `EventName` String,
    `SiteId` UInt32,
    `HostName` String,
    `PathName` String,
    `JourneyId` String,
    `Timestamp` DateTime('UTC'),
    `Referrer` String,
	`ReferrerSource` String,
	`CountryCode` String,
	`Subdivision1Code` String,
	`Subdivision2Code` String,
	`CityGeonameId` UInt32,
	`OperatingSystem` String,
	`OperatingSystemVersion` String,
	`Browser` String,
	`BrowserVersion` String,
	`UtmSource` String,
	`UtmMedium` String,
	`UtmCampaign` String,
	`UtmContent` String,
	`UtmTerm` String,
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(Timestamp)
ORDER BY (SiteId, Timestamp)