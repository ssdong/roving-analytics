'use client';

import { useState, useEffect } from 'react';
import {
  Card,
  DateRangePickerValue,
  BarList,
  Title,
  Flex,
  Text
} from '@tremor/react';

import { SITE_ID } from '@/app/utils/site';

import { useInterval } from '@/app/utils/interval';
import { formatDate } from '@/app/utils/time';
import FacebookIcon from '@/components/ui/icons/facebook';
import GoogleIcon from '@/components/ui/icons/google';
import GithubIcon from '@/components/ui/icons/github';
import LinkedinIcon from '@/components/ui/icons/linkedin';
import InstagramIcon from '@/components/ui/icons/instagram';
import RedditIcon from '@/components/ui/icons/reddit';
import TwitterIcon from '@/components/ui/icons/twitter';
import YoutubeIcon from '@/components/ui/icons/youtube';
import UnknownIcon from '@/components/ui/icons/unknown';

// Interface for Referrer Data
interface ReferrerSourceResult {
  ReferrerSource: string;
  PercentageOfReferrer: number;
}

interface ReferrerProps {
  dateRangeValue: DateRangePickerValue;
}

const isFromGoogle = (referrer: string): boolean => {
  return referrer.toLowerCase().includes('google');
};

const isFromTwitter = (referrer: string): boolean => {
  return referrer.toLowerCase().includes('twitter');
};

const isFromYouTube = (referrer: string): boolean => {
  return referrer.toLowerCase().includes('youtube');
};

const isFromGithub = (referrer: string): boolean => {
  return referrer.toLowerCase().includes('github');
};

const isFromReddit = (referrer: string): boolean => {
  return referrer.toLowerCase().includes('reddit');
};

const isFromFacebook = (referrer: string): boolean => {
  return referrer.toLowerCase().includes('facebook');
};

const isFromInstagram = (referrer: string): boolean => {
  return referrer.toLowerCase().includes('instagram');
};

const isFromLinkedin = (referrer: string): boolean => {
  return referrer.toLowerCase().includes('linkedin');
};

const getIconForReferrer = (referrer: string) => {
  if (isFromGoogle(referrer)) return GoogleIcon;
  if (isFromTwitter(referrer)) return TwitterIcon;
  if (isFromYouTube(referrer)) return YoutubeIcon;
  if (isFromGithub(referrer)) return GithubIcon;
  if (isFromReddit(referrer)) return RedditIcon;
  if (isFromFacebook(referrer)) return FacebookIcon;
  if (isFromInstagram(referrer)) return InstagramIcon;
  if (isFromLinkedin(referrer)) return LinkedinIcon;

  return UnknownIcon;
};

export default function ReferrerCard({ dateRangeValue }: ReferrerProps) {
  const today = new Date();
  const sevenDaysAgo = new Date(today);
  sevenDaysAgo.setDate(today.getDate() - 7);

  const to = dateRangeValue.to ? dateRangeValue.to : today;
  const from = dateRangeValue.from ? dateRangeValue.from : sevenDaysAgo;

  const [data, setData] = useState<ReferrerSourceResult[]>([]);

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
        `api/top-referrers?siteId=${SITE_ID}&timestampStart=${startTime}&timestampEnd=${endTime}&timezone=${timeZone}`
      );
      const fetchedData: ReferrerSourceResult[] = await response.json();
      setData(fetchedData);
    } catch (error) {
      console.error('Error fetching referrer data:', error);
    }
  };

  useEffect(() => {
    fetchData();
  }, [dateRangeValue]);

  useInterval(fetchData, 10 * 1000);

  return (
    <Card
      decoration="top"
      decorationColor="green"
      className="h-128 flex flex-col overflow-y-auto"
    >
      <Title>Top Referrers</Title>
      <Flex className="mt-6">
        <Text>Referrer</Text>
        <Text className="text-right">Percentage</Text>
      </Flex>
      <BarList
        data={data.map((item) => ({
          name: item.ReferrerSource,
          value: item.PercentageOfReferrer,
          icon: getIconForReferrer(item.ReferrerSource)
        }))}
        valueFormatter={(number: number) => `${number.toFixed(2)}%`}
        className="mt-2"
      />
    </Card>
  );
}
