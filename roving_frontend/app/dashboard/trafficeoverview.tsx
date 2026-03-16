'use client';

import { useRef, useState, useEffect } from 'react';
import { DoorOpen, Eye, Speech, PartyPopper, Users } from 'lucide-react';

import {
  Card,
  DateRangePickerValue,
  Grid,
  Col,
  AreaChart,
  Metric,
  Title,
  Text
} from '@tremor/react';

import { useInterval } from '@/app/utils/interval';
import { SITE_ID } from '@/app/utils/site';
import {
  generateDays,
  formatDate,
  DataTemplateCallback
} from '@/app/utils/time';
import Skeleton from '@/components/ui/skeleton';

interface TrafficData {
  Date: string;
  PageViews: number;
  UniqueVisitors: number;
}

interface TrafficeOverviewProps {
  dateRangeValue: DateRangePickerValue;
}

// This needs to be able to be configured by the user
const targetPage = "/sign-up"

const trafficDataTemplate: DataTemplateCallback = (date) => ({
  Date: date,
  PageViews: 0,
  UniqueVisitors: 0
});

function Placeholder() {
  return (
    <div className="flex items-center space-x-4 pt-2">
      <div className="space-y-2">
        <Skeleton className="h-4 w-[160px]" />
        <Skeleton className="h-4 w-[160px]" />
      </div>
    </div>
  );
}

function CurrentVisitorPulseIcon() {
  return (
    <Skeleton className="h-2 w-2 bg-greenish rounded-full inline-block mr-2" />
  );
}

export default function TrafficeOverview({
  dateRangeValue
}: TrafficeOverviewProps) {
  const hasMounted = useRef(false);

  const today = new Date();
  const sevenDaysAgo = new Date(today);
  sevenDaysAgo.setDate(today.getDate() - 7);

  const to = dateRangeValue.to ? dateRangeValue.to : today;
  const from = dateRangeValue.from ? dateRangeValue.from : sevenDaysAgo;

  const [trafficData, setTrafficData] = useState<TrafficData[]>([]);
  const [isInitialLoading, setIsInitialLoading] = useState(true);

  const [totalPageViews, setTotalPageViews] = useState<number>(0);
  const [totalUniqueVisitors, setTotalUniqueVisitors] = useState<number>(0);
  const [currentVisitors, setCurrentVisitors] = useState<number>(0);
  const [converionRate, setConversionRate] = useState<number>(0);
  const [bounceRate, setBounceRate] = useState<number>(0);

  const formatNumber = (num: number): string => {
    if (num < 1000000) {
      return Intl.NumberFormat('us').format(num);
    } else if (num === 1000000) {
      return '1 M';
    } else {
      const millionFormat = parseFloat((num / 1000000).toFixed(3)) + ' M';
      return millionFormat
    }
  }

  const fetchCurrentVisitorsData = async () => {
    try {
      const response = await fetch(`api/current-visitors?siteId=${SITE_ID}`);
      if (!response.ok) {
        throw new Error('Failed to fetch current visitors data.');
      }
      const data = await response.json();
      setCurrentVisitors(data.Count);
    } catch (error) {
      console.error('Error fetching current visitors:', error);
      setCurrentVisitors(0); // Set currentVisitors to 0 if the API call fails.
    }
  };

  const fetchConversionRate = async (startTime: string, endTime: string, targetPage: string, timeZone: string) => {
    try {
      const response = await fetch(
        `api/conversion-rate?siteId=${SITE_ID}&timestampStart=${startTime}&timestampEnd=${endTime}&targetConversionPage=${targetPage}&timezone=${timeZone}`
      );
      if (!response.ok) {
        throw new Error('Failed to fetch conversion rate data.');
      }
      const data = await response.json();
      setConversionRate(data.ConversionRatePercentage);
    } catch (error) {
      console.error('Error fetching conversion rate:', error);
      setConversionRate(0); // Set conversion rate to 0 if the API call fails.
    }
  };

  const fetchBounceRate = async (startTime: string, endTime: string, timeZone: string) => {
    try {
      const response = await fetch(
        `api/bounce-rate?siteId=${SITE_ID}&timestampStart=${startTime}&timestampEnd=${endTime}&timezone=${timeZone}`
      );
      if (!response.ok) {
        throw new Error('Failed to fetch bounce rate data.');
      }
      const data = await response.json();
      setBounceRate(data.BounceRatePercentage);
    } catch (error) {
      console.error('Error fetching bounce rate:', error);
      setBounceRate(0); // Set bounce rate to 0 if the API call fails.
    }
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

      const filledData = generateDays(from, to, trafficDataTemplate);

      await fetchCurrentVisitorsData();

      await fetchConversionRate(
        startTime,
        endTime,
        targetPage,
        timeZone
      );

      await fetchBounceRate(
        startTime,
        endTime,
        timeZone
      );

      const pageViewRes = await fetch(
        `api/page-views?siteId=${SITE_ID}&timestampStart=${startTime}&timestampEnd=${endTime}&timezone=${timeZone}`
      );
      const pageViewData: TrafficData[] = await pageViewRes.json();

      pageViewData.forEach((dataPoint) => {
        const index = filledData.findIndex((x) => x.Date === dataPoint.Date);
        if (index > -1) {
          filledData[index].PageViews = dataPoint.PageViews;
        }
      });

      const uniqueVisitorRes = await fetch(
        `api/unique-visitors?siteId=${SITE_ID}&timestampStart=${startTime}&timestampEnd=${endTime}&timezone=${timeZone}`
      );

      const uniqueVisitorData: TrafficData[] = await uniqueVisitorRes.json();

      uniqueVisitorData.forEach((dataPoint) => {
        const index = filledData.findIndex((x) => x.Date === dataPoint.Date);
        if (index > -1) {
          filledData[index].UniqueVisitors = dataPoint.UniqueVisitors;
        }
      });

      const totalPageViews = pageViewData.reduce(
        (acc, curr) => acc + curr.PageViews,
        0
      );
      const totalUniqueVisitors = uniqueVisitorData.reduce(
        (acc, curr) => acc + curr.UniqueVisitors,
        0
      );

      setTotalPageViews(totalPageViews);
      setTotalUniqueVisitors(totalUniqueVisitors);
      setTrafficData(filledData);
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  };

  useEffect(() => {
    fetchData();
    setIsInitialLoading(false);
  }, []);

  // This setup ensures the data is fetched only once on the initial render,
  // and then fetched again only when dateRangeValue changes (excluding the first render).
  // i.e. only when dateRangeValue changes(but excluding the first time when page loads cuz
  // data has already been fetched)
  useEffect(() => {
    if (hasMounted.current) {
      fetchData();
    } else {
      hasMounted.current = true;
    }
  }, [dateRangeValue]);

  // useInterval(fetchCurrentVisitorsData, 3 * 1000); // Fetch current visitors data every 10 seconds
  
  useInterval(fetchData, 3 * 1000);

  return (
    <Card>
      <Title className="mb-8">justinsdong.dev</Title>
      <Grid numItems={1} numItemsSm={2} numItemsLg={5} className="gap-6">
        <Col>
          <Card decoration="top" decorationColor="sky">
            <div className="flex items-center">
              <CurrentVisitorPulseIcon />
              <Text>Current Visitors</Text>
            </div>
            {isInitialLoading ? (
              <Placeholder />
            ) : (
              <Metric className="mt-3 flex justify-between">
                <span className="inline-block align-middle">
                  {formatNumber(currentVisitors)}
                </span>
                <Speech className="inline-block align-middle mt-1.5" />
              </Metric>
            )}
          </Card>
        </Col>
        <Col>
          <Card decoration="top" decorationColor="sky">
            <Text>Total Pageviews</Text>
            {isInitialLoading ? (
              <Placeholder />
            ) : (
              <Metric className="mt-3 flex justify-between">
                <span className="inline-block align-middle">
                  {formatNumber(totalPageViews)}
                </span>
                <Eye className="inline-block align-middle mt-1.5" />
              </Metric>
            )}
          </Card>
        </Col>
        <Col>
          <Card decoration="top" decorationColor="sky">
            <Text>Total Unique Visitors</Text>
            {isInitialLoading ? (
              <Placeholder />
            ) : (
              <Metric className="mt-3 flex justify-between">
                <span className="inline-block align-middle">
                  {formatNumber(totalUniqueVisitors)}
                </span>
                <Users className="inline-block align-middle mt-1.5" />
              </Metric>
            )}
          </Card>
        </Col>
        <Col>
          <Card decoration="top" decorationColor="sky">
            <Text>Conversion Rate</Text>
            {isInitialLoading ? (
              <Placeholder />
            ) : (
              <Metric className="mt-3 flex justify-between">
                <span className="inline-block align-middle">
                  {Intl.NumberFormat('us').format(converionRate)}%
                </span>
                <PartyPopper className="inline-block align-middle mt-1.5" />
              </Metric>
            )}
          </Card>
        </Col>
        <Col>
          <Card decoration="top" decorationColor="sky">
            <Text>Bounce Rate</Text>
            {isInitialLoading ? (
              <Placeholder />
            ) : (
              <Metric className="mt-3 flex justify-between">
                <span className="inline-block align-middle">
                  {Intl.NumberFormat('us').format(bounceRate)}%
                </span>
                <DoorOpen className="inline-block align-middle mt-1.5" />
              </Metric>
            )}
          </Card>
        </Col>
      </Grid>
      <Card className="mt-8" decoration="top" decorationColor="violet">
        <Title>Traffic Overview</Title>
        <AreaChart
          className="mt-4 h-80"
          data={trafficData}
          categories={['PageViews', 'UniqueVisitors']}
          index="Date"
          colors={['indigo', 'fuchsia']}
          valueFormatter={(number: number) =>
            Intl.NumberFormat('us').format(number).toString()
          }
          yAxisWidth={80}
        />
      </Card>
    </Card>
  );
}
