import React, { useEffect, useState } from 'react';

import {
  Card,
  DateRangePickerValue,
  Title,
  Subtitle,
  Text
} from '@tremor/react';

import { ResponsiveChoroplethCanvas } from '@nivo/geo';

import { useInterval } from '@/app/utils/interval';
import { formatDate } from '@/app/utils/time';
import { features } from '@/app/utils/features';
import { SITE_ID } from '@/app/utils/site';

interface CountryRankingProps {
  dateRangeValue: DateRangePickerValue;
}

interface RawCountryData {
  CountryCode: string;
  UniqueVisitors: number;
}

interface CountryData {
  id: string;
  value: number;
}

const convertRawCountryRankingData = (
  data: RawCountryData[]
): CountryData[] => {
  return data.map((item) => {
    return {
      id: item.CountryCode,
      value: item.UniqueVisitors
    };
  });
};

export default function CountryRankingCard({
  dateRangeValue
}: CountryRankingProps) {
  const [countryRankingData, setCountryRankingData] = useState<CountryData[]>(
    []
  );

  const today = new Date();
  const sevenDaysAgo = new Date(today);
  sevenDaysAgo.setDate(today.getDate() - 7);

  const to = dateRangeValue.to ? dateRangeValue.to : today;
  const from = dateRangeValue.from ? dateRangeValue.from : sevenDaysAgo;

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
      const response = await fetch(
        `api/country-ranking?siteId=${SITE_ID}&timestampStart=${startTime}&timestampEnd=${endTime}&timezone=${timeZone}`
      );
      const fetchedData: RawCountryData[] = await response.json();

      setCountryRankingData(convertRawCountryRankingData(fetchedData));
    } catch (error) {
      console.error('Error fetching sequence data:', error);
    }
  };

  useEffect(() => {
    fetchData();
  }, [dateRangeValue]);

  useInterval(fetchData, 10 * 1000);

  return (
    <Card decoration="top" decorationColor="stone" className="mb-8">
      <Title>Country-wise User Distribution</Title>
      <Subtitle>Visual representation of unique visitors by country</Subtitle>
      <Card className="mt-3 h-128">
        <ResponsiveChoroplethCanvas
          data={countryRankingData}
          features={features['features']}
          margin={{ top: 0, right: 0, bottom: 0, left: 0 }}
          colors="RdBu"
          domain={[0, 1000]}
          unknownColor="#101b42"
          label="properties.name"
          valueFormat=".2s"
          projectionTranslation={[0.5, 0.5]}
          projectionRotation={[0, 0, 0]}
          enableGraticule={true}
          graticuleLineColor="rgba(0, 0, 0, .2)"
          borderWidth={0.5}
          borderColor="#101b42"
          legends={[
            {
              anchor: 'bottom-left',
              direction: 'column',
              justify: true,
              translateX: 20,
              translateY: -60,
              itemsSpacing: 0,
              itemWidth: 92,
              itemHeight: 18,
              itemDirection: 'left-to-right',
              itemOpacity: 0.85,
              symbolSize: 18
            }
          ]}
        />
      </Card>
    </Card>
  );
}
