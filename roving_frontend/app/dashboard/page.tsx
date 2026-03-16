'use client';

import { useState } from 'react'
import { Flex, Grid } from '@tremor/react';

import { DateRangePickerValue } from '@tremor/react';

import TrafficOverview from '@/app/dashboard/trafficeoverview';
import DeviceCard from '@/app/dashboard/devices';
import ReferrerCard from '@/app/dashboard/referrer';
import CommonJourneyCard from '@/app/dashboard/commonjourney';
import CountryRankingCard from '@/app/dashboard/country';
import FunnelCard from '@/app/dashboard/funnel';
import TimeSelector from '@/app/dashboard/timeselector';

export default function DashbordPage() {
  const today = new Date()
  const sevenDaysAgo = new Date(today);
  sevenDaysAgo.setDate(today.getDate() - 7);

  const [dateRangeValue, setDateRangeValue] = useState<DateRangePickerValue>({
    from: sevenDaysAgo,
    to: today,
  });

  return (
    <main className="p-4 md:p-10 mx-auto max-w-7xl">
      <Flex justifyContent="end" className="w-full mb-5">
        <div className="w-full"></div>
        <TimeSelector value={dateRangeValue} setValue={setDateRangeValue} />
      </Flex>
      <div className="mb-8">
        <TrafficOverview dateRangeValue={dateRangeValue} />
      </div>
      <Grid numItemsSm={2} numItemsLg={2} className="mb-8 gap-6 sm:h-128">
        <ReferrerCard dateRangeValue={dateRangeValue}/>
        <DeviceCard dateRangeValue={dateRangeValue} />
      </Grid>
      <CountryRankingCard dateRangeValue={dateRangeValue}/>
      <CommonJourneyCard dateRangeValue={dateRangeValue}/>
      <FunnelCard dateRangeValue={dateRangeValue}/>
    </main>
  );
}
