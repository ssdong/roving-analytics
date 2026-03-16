import React, { useState, useEffect, useRef } from 'react';
import {
  Button,
  Card,
  Flex,
  Grid,
  DateRangePickerValue,
  Title,
  Subtitle,
  TextInput,
  List,
  ListItem
} from '@tremor/react';
import { BadgeHelp, Trash2, AlertTriangle, Filter } from 'lucide-react';

import { formatDate } from '@/app/utils/time';
import { SITE_ID } from '@/app/utils/site';
import Skeleton from '@/components/ui/skeleton';
import { ResponsiveFunnel } from '@nivo/funnel';

interface FunnelProps {
  dateRangeValue: DateRangePickerValue;
}

interface FunnelResult {
  Sequence: string[];
  Count: number;
}

/**
 * 1. ^: This asserts the start of the string.
   2. \/: Matches the literal / character.
   3. [a-z0-9-]+: Matches one or more alphanumeric characters or hyphens. This is for the URL segment after the initial /.
   4. (\/[a-z0-9-]+)*: This captures additional URL segments.
     i) \/: Matches the literal / character.
     ii) [a-z0-9-]+: Matches one or more alphanumeric characters or hyphens for the segment.
     iii) *: This allows for zero or more of these segments.
   5. (\?[a-z0-9-]+(=[a-z0-9-]*)?(&[a-z0-9-]+(=[a-z0-9-]*)?)*)?: This captures the query parameters.
     i) \?: Matches the literal ? character, which signals the start of the query parameters.
     ii) [a-z0-9-]+: Matches one or more alphanumeric characters or hyphens for the key of a query parameter.
     iii) (=[a-z0-9-]*)?: Optionally matches the = symbol followed by zero or more alphanumeric characters or hyphens for the value of the query parameter.
     iv) (&[a-z0-9-]+(=[a-z0-9-]*)?)*): Matches the subsequent query parameters.
        a) &: Matches the literal & character, which separates query parameters.
        b) [a-z0-9-]+: Matches one or more alphanumeric characters or hyphens for the key of a subsequent query parameter.
        c) (=[a-z0-9-]*)?: Optionally matches the = symbol followed by zero or more alphanumeric characters or hyphens for the value of a subsequent query parameter.
     v) The outer ? makes the entire query parameter section optional.
    6. (#[a-z0-9-]+)?: Captures the URL hash fragment.
      i) #: Matches the literal # character.
      ii) [a-z0-9-]+: Matches one or more alphanumeric characters or hyphens for the hash fragment.
      iii) ?: This whole hash fragment section is optional.
    7. $: This asserts the end of the string.
    8. /i: This flag makes the regex case-insensitive.

    Regex will match patterns like:

    /example-url
    /example-url/second-segment
    /example-url?query=value
    /example-url?query=value&secondQuery=secondValue
    /example-url#hashValue
    /example-url/second-segment#hashValue
    /example-url/second-segment?query=value#hashValue
 * 
 */
const urlRegex =
  /^\/[a-z0-9-]+(\/[a-z0-9-]+)*(\?[a-z0-9-]+(=[a-z0-9-]*)?(&[a-z0-9-]+(=[a-z0-9-]*)?)*)?(#[a-z0-9-]+)?$/i;

export default function FunnelCard({ dateRangeValue }: FunnelProps) {
  const listRef = useRef<HTMLUListElement>(null);

  const [url, setUrl] = useState('');
  const [urlList, setUrlList] = useState<{ id: number; url: string }[]>([]);
  const [placeholderWarning, setPlaceholderWarning] = useState(false);

  const [funnelSearchInitiated, setFunnelSearchInitiated] = useState(false);
  const [funnelDataLoaded, setFunnelDataLoaded] = useState(false);

  const [funnelData, setFunnelData] = useState<FunnelResult[]>([]);

  const today = new Date();
  const sevenDaysAgo = new Date(today);
  sevenDaysAgo.setDate(today.getDate() - 7);

  const options: Intl.DateTimeFormatOptions = {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  };

  const to = dateRangeValue.to ? dateRangeValue.to : today;
  const from = dateRangeValue.from ? dateRangeValue.from : sevenDaysAgo;

  const isFunnelDataEmpty = (funnelData: FunnelResult[]): boolean => {
    return funnelData.every((item) => item.Count === 0);
  };

  const handleAddUrl = () => {
    if (url && urlRegex.test(url)) {
      setUrlList((prevList) => [...prevList, { id: Date.now(), url }]);
      setUrl('');
    } else {
      setUrl(''); // Clear the input field
      setPlaceholderWarning(true); // This will trigger a change in the placeholder text
    }
  };

  const handleDeleteUrl = (idToDelete: number) => {
    const filteredUrls = urlList.filter((item) => item.id !== idToDelete);
    setUrlList(filteredUrls);
  };

  const handleSearch = async () => {
    setFunnelSearchInitiated(true);
    setFunnelDataLoaded(false);

    // TODO: take out setTimeout
    setTimeout(() => {
      fetchData();
    }, 3000);
  };

  const fetchData = async () => {
    try {
      // We adjust the "to" date to the end of the day in the user's timezone.
      // This ensures that the selected date will encompass all data up to the very
      // last millisecond of the day in the user's timezone, even when converted to UTC.

      // We also adjust the "from" date to the beginning of the day in the user's timezone
      // to ensure that the selected date will encompass all data from the very first
      // millisecond of the day in the user's timezone, even when converted to UTC.
      from.setHours(0, 0, 0, 0);
      to.setHours(23, 59, 59, 999);

      const startTime = formatDate(from);
      const endTime = formatDate(to);

      let timeZone = undefined;

      try {
        timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;
      } catch (e) {
        timeZone = 'UTC';
      }

      if (!timeZone) {
        timeZone = 'UTC';
      }

      const url = urlList.map((item) => item.url).join(',');

      const response = await fetch(
        `api/funnel?siteId=${SITE_ID}&timestampStart=${startTime}&timestampEnd=${endTime}&timezone=${timeZone}&url=${url}`
      );
      const fetchedData = await response.json();

      // Shorter sequence should come before longer sequence in Funnel's context
      const sortedFunnelData = [...fetchedData].sort((a, b) => a.Sequence.length - b.Sequence.length);

      setFunnelData(sortedFunnelData);
      setFunnelSearchInitiated(false);
      setFunnelDataLoaded(true);
    } catch (error) {
      console.error('Error fetching sequence data:', error);
    }
  };

  const NotInitiatedPlaceholder = () => {
    return (
      <div className="flex flex-col items-center justify-center h-full text-center">
        <Filter className="w-32 h-32 mb-4" />
        <Subtitle>Enter URL sequence to visualize your funnel</Subtitle>
      </div>
    );
  };

  const NoFunnelAvailablePlaceholder = () => {
    return (
      <div className="flex flex-col items-center justify-center h-full text-center">
        <BadgeHelp className="w-32 h-32 mb-4" />
        <Subtitle>
          It seems no users have followed this exact path. Try a different
          sequence
        </Subtitle>
      </div>
    );
  };

  const LoadingPlaceholder = () => {
    return (
      <div className="flex flex-col items-center justify-center h-full space-x-4 pt-2">
        <div className="flex items-center space-x-4">
          <Skeleton className="h-12 w-12 md:h-16 md:w-16 rounded-full" />
          <div className="space-y-2">
            <Skeleton className="h-6 w-[200px] md:w-[360px]" />
            <Skeleton className="h-6 w-[200px] md:w-[360px]" />
          </div>
        </div>
        <Subtitle className="mt-4">Generating funnel...</Subtitle>
      </div>
    );
  };

  // `useEffect` hook to automatically scroll the list to its bottom whenever `urlList` changes.
  // This ensures that whenever a new URL is added to the list (or a URL is removed),
  // the list auto-scrolls to show the latest (or last remaining) URL entry.
  // Otherwise, if this effect wasn't used, users might need to manually scroll down to view the latest entries,
  // which can be inconvenient especially when the list grows long.
  useEffect(() => {
    if (listRef.current) {
      listRef.current.scrollTop = listRef.current.scrollHeight;
    }
  }, [urlList]);

  useEffect(() => {
    // Reset state values when dateRangeValue changes
    setFunnelSearchInitiated(false);
    setFunnelDataLoaded(false);
  }, [dateRangeValue]);

  return (
    <Card decoration="top" decorationColor="fuchsia" className="mt-8">
      <Flex>
        <div>
          <Title>Funnels</Title>
          <Subtitle>Enter your desired page sequence and search</Subtitle>
        </div>
        <Subtitle className="mt-6">
          {from.toLocaleDateString('en-US', options)} -{' '}
          {to.toLocaleDateString('en-US', options)}
        </Subtitle>
      </Flex>
      <Grid className="grid-cols-1 sm:grid-cols-3 mt-5 gap-6">
        <Card className="col-span-1 h-128">
          <div className="mt-5">
            <TextInput
              placeholder={
                placeholderWarning
                  ? 'Invalid URL format'
                  : 'Enter URL that starts with / (e.g., /about-us)'
              }
              value={url}
              onChange={(e) => {
                setPlaceholderWarning(false);
                setUrl(e.target.value);
              }}
              className="mb-2"
              icon={placeholderWarning ? AlertTriangle : undefined}
            />
            <div className="flex space-x-4 mt-2">
              <Button onClick={handleAddUrl} size="sm" className="mb-5">
                Add URL
              </Button>
              <Button
                onClick={handleSearch}
                size="sm"
                className="mb-5"
                variant="secondary"
                disabled={urlList.length === 0}
                loading={funnelSearchInitiated && !funnelDataLoaded}
              >
                Search
              </Button>
            </div>
          </div>

          <Card className="max-w-xs">
            <Title>URL Sequence</Title>
            <List className="overflow-y-auto max-h-60" ref={listRef}>
              {urlList.map((item, index) => (
                <ListItem
                  key={item.id}
                  className="flex justify-between items-center"
                >
                  <div className="max-w-4/5 overflow-x-auto">
                    <span>
                      {index + 1}. {item.url}
                    </span>
                  </div>
                  <Button
                    size="xs"
                    variant="secondary"
                    className="ml-2"
                    onClick={() => handleDeleteUrl(item.id)}
                  >
                    <Trash2 className="h-3 w-3" />
                  </Button>
                </ListItem>
              ))}
            </List>
          </Card>
        </Card>
        <Card className="sm:col-span-2 h-128">
          {!funnelSearchInitiated && !funnelDataLoaded && (
            <NotInitiatedPlaceholder />
          )}
          {funnelSearchInitiated && !funnelDataLoaded && <LoadingPlaceholder />}
          {!funnelSearchInitiated &&
            funnelDataLoaded &&
            isFunnelDataEmpty(funnelData) && <NoFunnelAvailablePlaceholder />}
          {!funnelSearchInitiated &&
            funnelDataLoaded &&
            !isFunnelDataEmpty(funnelData) &&
             (
              <ResponsiveFunnel
                data={funnelData.map((item) => ({
                  id: item.Sequence.join('_'), // concatenate entire sequence for uniqueness
                  label: `${item.Sequence[item.Sequence.length - 1]} - users`,
                  value: item.Count
                }))}
                margin={{ top: 20, right: 20, bottom: 20, left: 20 }}
                colors={{ scheme: 'pastel1' }}
                borderWidth={20}
                labelColor={{
                  from: 'color',
                  modifiers: [['darker', 3]]
                }}
                beforeSeparatorLength={100}
                beforeSeparatorOffset={20}
                afterSeparatorLength={100}
                afterSeparatorOffset={20}
                currentPartSizeExtension={10}
                currentBorderWidth={40}
                motionConfig="wobbly"
              />
            )}
        </Card>
      </Grid>
    </Card>
  );
}
