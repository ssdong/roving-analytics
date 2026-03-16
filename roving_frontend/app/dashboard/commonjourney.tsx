import React, { useEffect, useState } from 'react';

import {
  Card,
  DateRangePickerValue,
  Title,
  Subtitle,
  BarChart,
  Text
} from '@tremor/react';

import { useInterval } from '@/app/utils/interval';
import { formatDate } from '@/app/utils/time';
import { SITE_ID } from '@/app/utils/site';

interface CommonJourneyProps {
  dateRangeValue: DateRangePickerValue;
}

interface SequenceData {
    name: string;
    Percentage: number;
    Sequence: string;
  }
  

interface RawSequenceData {
  Sequence: string[];
  Count: number;
  Percentage: number;
}

const convertSequenceData = (data: RawSequenceData[]): SequenceData[] => {
  return data.map((item) => {
    return {
      name: `${item.Count} Visitors`,
      Sequence: item.Sequence.join(' '),
      Percentage: item.Percentage
    };
  });
};

export default function CommonJourneyCard({
  dateRangeValue
}: CommonJourneyProps) {
  const [commonJourney, setCommonJourney] = useState<SequenceData[]>([]);

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
        `api/common-journey?siteId=${SITE_ID}&timestampStart=${startTime}&timestampEnd=${endTime}&timezone=${timeZone}`
      );
      const fetchedData: RawSequenceData[] = await response.json();

      setCommonJourney(convertSequenceData(fetchedData));
    } catch (error) {
      console.error('Error fetching sequence data:', error);
    }
  };

  useEffect(() => {
    fetchData();
  }, [dateRangeValue]);

  useInterval(fetchData, 10 * 1000);

  return (
    <Card decoration="top" decorationColor="amber">
      <Title>User Navigation Sequences</Title>
      <Subtitle>
        Analysis of common user navigation patterns(Hover to see sequence)
      </Subtitle>
      <Card className="mt-3">
        <Text className="mb-[-10px]">Percentage</Text>
        <BarChart
          data={commonJourney}
          index="name"
          categories={['Percentage', 'Sequence']}
          colors={['blue', 'indigo']}
          yAxisWidth={48}
        />
      </Card>
    </Card>
  );
}
