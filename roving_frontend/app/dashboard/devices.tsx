'use client';

import { useState, useEffect } from 'react';
import {
  Card,
  BarList,
  DateRangePickerValue,
  Title,
  Flex,
  Text
} from '@tremor/react';

import { useInterval } from '@/app/utils/interval';
import { formatDate } from '@/app/utils/time';
import { SITE_ID } from '@/app/utils/site';

import ChromeIcon from '@/components/ui/icons/chrome';
import FirefoxIcon from '@/components/ui/icons/firefox';
import InternetExplorerIcon from '@/components/ui/icons/internetexplorer';
import OperaIcon from '@/components/ui/icons/opera';
import EdgeIcon from '@/components/ui/icons/edge';
import SafariIcon from '@/components/ui/icons/safari';
import UnknownIcon from '@/components/ui/icons/unknown';

interface DeviceResult {
  Device: string;
  PercentageOfDevice: number;
}

interface DeviceProps {
  dateRangeValue: DateRangePickerValue;
}

const isChrome = (device: string): boolean => {
  return device.toLowerCase().includes('chrome');
};

const isFirefox = (device: string): boolean => {
  return device.toLowerCase().includes('firefox');
};

const isOpera = (device: string): boolean => {
  return device.toLowerCase().includes('opera');
};

const isEdge = (device: string): boolean => {
  return device.toLowerCase().includes('edge');
};

const isSafari = (device: string): boolean => {
  return device.toLowerCase().includes('safari');
};

const isInternetExplorer = (device: string): boolean => {
  return (
    device.toLowerCase().includes('internet explorer') ||
    device.toLowerCase() === 'ie'
  );
};

const getIconForDevice = (device: string) => {
  if (isChrome(device)) return ChromeIcon;
  if (isFirefox(device)) return FirefoxIcon;
  if (isInternetExplorer(device)) return InternetExplorerIcon;
  if (isOpera(device)) return OperaIcon;
  if (isEdge(device)) return EdgeIcon;
  if (isSafari(device)) return SafariIcon;
  return UnknownIcon;
};

export default function DeviceCard({ dateRangeValue }: DeviceProps) {
  const today = new Date();
  const sevenDaysAgo = new Date(today);
  sevenDaysAgo.setDate(today.getDate() - 7);

  const to = dateRangeValue.to ? dateRangeValue.to : today;
  const from = dateRangeValue.from ? dateRangeValue.from : sevenDaysAgo;

  const [data, setData] = useState<DeviceResult[]>([]);

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
        `api/top-devices?siteId=${SITE_ID}&timestampStart=${startTime}&timestampEnd=${endTime}&timezone=${timeZone}`
      );
      const fetchedData: DeviceResult[] = await response.json();
      setData(fetchedData);
    } catch (error) {
      console.error('Error fetching device data:', error);
    }
  };

  useEffect(() => {
    fetchData();
  }, [dateRangeValue]);

  useInterval(fetchData, 10 * 1000);

  return (
    <Card
      decoration="top"
      decorationColor="emerald"
      className="h-128 flex flex-col overflow-y-auto"
    >
      <Title>Top Devices</Title>
      <Flex className="mt-6">
        <Text>Device</Text>
        <Text className="text-right">Percentage</Text>
      </Flex>
      <BarList
        data={data.map((item) => ({
          name: item.Device,
          value: item.PercentageOfDevice,
          icon: getIconForDevice(item.Device)
        }))}
        valueFormatter={(number: number) => `${number.toFixed(2)}%`}
        className="mt-2"
      />
    </Card>
  );
}
